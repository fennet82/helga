package models

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"

	helga_errors "cicd/operators/helga/errors"
	"cicd/operators/helga/internal/logger"
	"cicd/operators/helga/internal/utils"
	"cicd/operators/helga/internal/vars"
)

type Artifact struct {
	DecideByVersion *bool   `yaml:"decideByVersion"` // will decide by date if not true
	Domain          string  `yaml:"domain"`
	Username        string  `yaml:"username"`
	Password        string  `yaml:"password"`
	Repos           []*Repo `yaml:"repos"`
}

func (a *Artifact) Validate() []error {
	var (
		validationErrs []error
		structName     = "Artifact"
		dReg           = regexp.MustCompile(vars.DOMAIN_VALIDATION_REGEX)
	)

	if !dReg.MatchString(a.Domain) {
		validationErrs = append(validationErrs, helga_errors.ErrValidation{StructName: structName, DerivedFromErr: fmt.Errorf(
			"domain: %s, did not pass regex validation please refer to this regex for fixing: %s", a.Domain, vars.DOMAIN_VALIDATION_REGEX,
		)})
	}

	if a.Username == "" {
		validationErrs = append(validationErrs, helga_errors.ErrValidation{StructName: structName, DerivedFromErr: errors.New("username field cannot be empty")})
	}

	if a.Password == "" {
		validationErrs = append(validationErrs, helga_errors.ErrValidation{StructName: structName, DerivedFromErr: errors.New("password field cannot be empty")})
	}

	errs, filteredRepos := utils.FilterByValidation(utils.ToValidatableSlice(a.Repos), "repo: %v, did not pass validation, changing availability to false")
	helga_errors.HandleErrors(errs)

	a.Repos = utils.FromValidatableSlice[*Repo](filteredRepos)

	if len(a.Repos) == 0 {
		validationErrs = append(validationErrs, helga_errors.ErrValidation{StructName: structName, DerivedFromErr: errors.New("repos list cannot be empty")})
	}

	helga_errors.HandleErrors(validationErrs)

	return validationErrs
}

func (dest *Artifact) Sync(src *Artifact) {
	if src.Domain != "" && dest.Domain == "" {
		dest.Domain = src.Domain
	}

	if src.Username != "" && dest.Domain == "" {
		dest.Username = src.Username
	}

	if src.Password != "" && dest.Domain == "" {
		dest.Password = src.Password
	}

	if src.DecideByVersion != nil && dest.DecideByVersion == nil {
		dest.DecideByVersion = src.DecideByVersion
	}

	syncReposList(&dest.Repos, &src.Repos)
}

func syncReposList(destRepos *[]*Repo, srcRepos *[]*Repo) {
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
}

func (a *Artifact) FetchHelmPacakgesByAQL() []*ArtifactHelmPackage {
	artifactoryErr := helga_errors.ErrArtifactoryAPI{}
	pkgs := []*ArtifactHelmPackage{}
	client := http.Client{}
	aqlTemplate := `
items.find({
	"repo": {"$eq": "%s"},
	"path": {"$eq": "%s"},
	"name": {"$match": "*.tgz"}
}).include("repo", "path", "name", "created", "modified")
`

	type ArtifactResBody struct {
		Results []*ArtifactHelmPackage `json:"results"`
	}

	for _, r := range a.Repos {
		artifactoryErr.Repo = r.Name

		for _, p := range r.Paths {
			logger.GetLoggerInstance().Info(fmt.Sprintf("initiating req - for repo: %s, path: %s", r.Name, p))

			artifactoryErr.DerivedFromErr = nil
			artifactoryErr.Path = p

			req, err := http.NewRequest("POST",
				a.Domain+vars.AQL_ARTIFACT_PATH_POSTFIX,
				bytes.NewBuffer(fmt.Appendf(nil, aqlTemplate, r.Name, p)),
			)

			if err != nil {
				artifactoryErr.DerivedFromErr = err
				helga_errors.HandleError(&artifactoryErr)

				continue
			}

			req.Header.Set("Content-Type", "text/plain")
			req.Header.Set("Content-Type", "Basic "+base64.StdEncoding.EncodeToString([]byte(a.Username+":"+a.Password)))

			resp, err := client.Do(req)
			if err != nil {
				artifactoryErr.DerivedFromErr = err
				helga_errors.HandleError(&artifactoryErr)

				panic(err)
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				artifactoryErr.DerivedFromErr = err
				helga_errors.HandleError(&artifactoryErr)

				panic(err)
			}

			defer resp.Body.Close()

			logMsg := fmt.Sprintf("req - on repo: %s, path: %s, gave status code: %d", r.Name, p, resp.StatusCode)

			if resp.StatusCode == http.StatusOK {
				logger.GetLoggerInstance().Info(logMsg)

				var res ArtifactResBody

				json.Unmarshal(body, &res)

				logger.GetLoggerInstance().Info(fmt.Sprintf("found %d pacakges for repo: %s on path: %s", len(res.Results), r.Name, p))

				pkgs = append(pkgs, res.Results...)
			} else {
				logger.GetLoggerInstance().Warn(logMsg)
			}
		}
	}

	return pkgs
}
