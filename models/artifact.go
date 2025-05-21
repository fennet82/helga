package models

import (
	"bytes"
	helga_errors "cicd/operators/helga/errors"
	"cicd/operators/helga/internal/logger"
	"cicd/operators/helga/internal/utils"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Artifact struct {
	DecideByVersion *bool   `yaml:"decideByVersion"` // will decide by date if not true
	Domain          string  `yaml:"domain"`
	Username        string  `yaml:"username"`
	Password        string  `yaml:"password"`
	Repos           []*Repo `yaml:"repos"`
}

func (a *Artifact) Validate() error {
	var (
		validationErr error  = nil
		structName    string = "Artifact"
	)

	if _, err := url.ParseRequestURI(a.Domain); err != nil {
		validationErr = &helga_errors.ErrValidation{StructName: structName, DerivedFromErr: err}
	}

	if a.Username == "" {
		validationErr = &helga_errors.ErrValidation{StructName: structName, DerivedFromErr: errors.New("username field cannot be empty")}
	}

	if a.Password == "" {
		validationErr = &helga_errors.ErrValidation{StructName: structName, DerivedFromErr: errors.New("password field cannot be empty")}
	}

	errs, filteredRepos := utils.FilterByValidation(utils.ToValidatableSlice(a.Repos), "repo: %v, did not pass validation, changing availability to false")
	helga_errors.HandleErrors(errs)

	a.Repos, validationErr = utils.FromValidatableSlice[*Repo](filteredRepos)

	if len(a.Repos) == 0 {
		validationErr = &helga_errors.ErrValidation{StructName: structName, DerivedFromErr: errors.New("repos list cannot be empty")}
	}

	helga_errors.HandleError(validationErr)

	return validationErr
}

func (dest_a *Artifact) SyncArtifact(src_a *Artifact) {
	if src_a.Domain != "" && dest_a.Domain == "" {
		dest_a.Domain = src_a.Domain
	}

	if src_a.Username != "" && dest_a.Domain == "" {
		dest_a.Username = src_a.Username
	}

	if src_a.Password != "" && dest_a.Domain == "" {
		dest_a.Password = src_a.Password
	}

	if len(src_a.Repos) > 0 && len(dest_a.Repos) == 0 {
		dest_a.Repos = src_a.Repos
	}

	if src_a.DecideByVersion != nil && dest_a.DecideByVersion == nil {
		dest_a.DecideByVersion = src_a.DecideByVersion
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
				a.Domain+"/artifactory/api/search/aql",
				bytes.NewBuffer([]byte(fmt.Sprintf(aqlTemplate, r.Name, p))),
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
