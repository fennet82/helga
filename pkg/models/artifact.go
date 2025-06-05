package models

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/fennet82/helga/internal/utils"
	"github.com/fennet82/helga/internal/vars"
	helga_errors "github.com/fennet82/helga/pkg/errors"
	helmclient "github.com/mittwald/go-helm-client"
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

func syncReposList(destRepos *[]*Repo, srcRepos *[]*Repo) error {
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

func (a *Artifact) AddOrUpdateHelmRepos(hc helmclient.Client) {
	for _, r := range a.Repos {
		err := hc.AddOrUpdateChartRepo(repo.Entry{
			Name:                  r.Name,
			URL:                   a.Domain + "/" + r.Name,
			Username:              a.Username,
			Password:              a.Password,
			InsecureSkipTLSverify: true,
			PassCredentialsAll:    false,
		})

		if err != nil {
			helga_errors.HandleError(err)
		}
	}
}
