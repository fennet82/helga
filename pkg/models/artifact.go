package models

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"

	"github.com/fennet82/helga/internal/logger"
	"github.com/fennet82/helga/internal/utils"
	"github.com/fennet82/helga/internal/vars"
	helga_errors "github.com/fennet82/helga/pkg/errors"
	"helm.sh/helm/v3/pkg/repo"
)

type Artifact struct {
	Domain   string  `yaml:"domain"`
	Username string  `yaml:"username"`
	Password string  `yaml:"password"`
	Repos    []*Repo `yaml:"repos"`
}

func (a *Artifact) String() string {
	return a.Domain
}

func (a *Artifact) Validate() []error {
	logger.GetLoggerInstance().Info(fmt.Sprintf("starting validation for artifact: %s", a.String()))

	var (
		validationErrs []error
		structName     = "Artifact"
		dReg           = regexp.MustCompile(vars.ARTIFACTORY_VALIDATION_REGEX)
	)

	if !dReg.MatchString(a.Domain) {
		validationErrs = append(validationErrs, helga_errors.ErrValidation{StructName: structName, DerivedFromErr: fmt.Errorf(
			"domain: %s, did not pass regex validation please refer to this regex for fixing: %s", a.Domain, vars.ARTIFACTORY_VALIDATION_REGEX,
		)})
	}

	if a.Username == "" {
		validationErrs = append(validationErrs, helga_errors.ErrValidation{StructName: structName, DerivedFromErr: errors.New("username field cannot be empty")})
	}

	if a.Password == "" {
		validationErrs = append(validationErrs, helga_errors.ErrValidation{StructName: structName, DerivedFromErr: errors.New("password field cannot be empty")})
	}

	errs, filteredRepos := utils.FilterByValidation(utils.ToValidatableSlice(a.Repos), "repo: %s, did not pass validation, changing availability to false")
	helga_errors.HandleErrors(errs)

	a.Repos = utils.FromValidatableSlice[*Repo](filteredRepos)

	if len(a.Repos) == 0 {
		validationErrs = append(validationErrs, helga_errors.ErrValidation{StructName: structName, DerivedFromErr: errors.New("repos list cannot be empty")})
	}

	helga_errors.HandleErrors(validationErrs)

	return validationErrs
}

func (dest *Artifact) Sync(src *Artifact) error {
	if dest == nil || src == nil {
		return fmt.Errorf("cannot sync nil Artifacts objects")
	}

	if src.Domain != "" && dest.Domain == "" {
		dest.Domain = src.Domain
	}

	if src.Username != "" && dest.Username == "" {
		dest.Username = src.Username
	}

	if src.Password != "" && dest.Password == "" {
		dest.Password = src.Password
	}

	syncReposList(&dest.Repos, &src.Repos)

	return nil
}

func syncReposList(destRepos, srcRepos *[]*Repo) error {
	if destRepos == nil || srcRepos == nil {
		return fmt.Errorf("cannot sync nil Repo slices")
	}

	seen := make(map[string]*Repo)

	for _, r := range *destRepos {
		seen[r.Name] = r
	}

	for _, r := range *srcRepos {
		if existing, exists := seen[r.Name]; exists {
			existing.Sync(r)
		} else {
			*destRepos = append(*destRepos, r)
		}
	}

	return nil
}

func (a *Artifact) GetArtifactReposAsEntries() []repo.Entry {
	helmRepoEntries := []repo.Entry{}

	for _, r := range a.Repos {
		e := r.GetAsHelmRepoEntry()

		e.URL = a.Domain + "/" + r.Name
		e.Username = a.Username
		e.Password = a.Password

		helmRepoEntries = append(helmRepoEntries, *e)
	}

	return helmRepoEntries
}

func (a *Artifact) GetChartPkgsInArtifact() map[string]HelmChart {
	aqlQuery := `
items.find({
	"repo": {"eq": "%s"},
	"path": {"eq": "%s"},
	"name": {"match": "*.tgz"}
})
`
	type ArtifactoryResponse struct {
		Results []ArtifactHelmPackage `json:"results"`
	}

	artifactoryHelmPackages := make(map[string]HelmChart)
	client := http.Client{}

	for _, r := range a.Repos {
		for _, p := range r.Paths {
			logger.GetLoggerInstance().Info(fmt.Sprintf("sending request to fetch helm pkgs for repo: %s, path: %s", r.String(), p))

			req, err := http.NewRequest("POST", a.Domain+"/"+vars.AQL_ARTIFACT_PATH_POSTFIX, bytes.NewBufferString(fmt.Sprintf(aqlQuery, r.Name, p)))
			if err != nil {
				err := helga_errors.ErrArtifactoryAPI{
					DerivedFromErr: fmt.Errorf("generating request to the artifactory was unsuccesful"),
					Repo:           r.String(),
					Path:           p,
				}
				helga_errors.HandleError(err)
				panic(err)
			}

			req.SetBasicAuth(a.Username, a.Password)
			req.Header.Set("Content-Type", "text/plain")

			resp, err := client.Do(req)
			if err != nil {
				err := helga_errors.ErrArtifactoryAPI{
					DerivedFromErr: fmt.Errorf("request to the artifactory was unsuccesful"),
					Repo:           r.String(),
					Path:           p,
				}
				helga_errors.HandleError(err)
				panic(err)
			}

			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				err := helga_errors.ErrArtifactoryAPI{
					DerivedFromErr: fmt.Errorf("request to the artifactory was unsuccesful returned status code: %d, needs to be %d", resp.StatusCode, http.StatusOK),
					Repo:           r.String(),
					Path:           p,
				}

				helga_errors.HandleError(err)
				panic(err)
			}

			var ar ArtifactoryResponse
			if err := json.NewDecoder(resp.Body).Decode(&ar); err != nil {
				err := helga_errors.ErrArtifactoryAPI{
					DerivedFromErr: fmt.Errorf("couldn't parse response to struct for"),
					Repo:           r.String(),
					Path:           p,
				}

				helga_errors.HandleError(err)
				panic(err)
			}

			for _, resPkg := range ar.Results {
				if err := resPkg.Validate(); err != nil {
					helga_errors.HandleError(fmt.Errorf("validation failed for pkg fetched from artifactory api reason: %s", err.Error()))
					continue
				}

				seenHelmPkg, exists := artifactoryHelmPackages[resPkg.Name()]
				if exists {
					pkg, err := DetermineNewerPkg(seenHelmPkg, resPkg, r.DecideByVersion)
					if err != nil {
						helga_errors.HandleError(err)
						continue
					}

					artifactoryHelmPackages[resPkg.Name()] = pkg
				} else {
					artifactoryHelmPackages[resPkg.Name()] = resPkg
				}
			}
		}
	}

	return artifactoryHelmPackages
}

func (a *Artifact) GetRepoByName(repoName string) (repo *Repo) {
	repo = nil

	for _, r := range a.Repos {
		if r.Name == repoName {
			repo = r
		}
	}

	return
}
