package config

import (
	helga_errors "cicd/operators/helga/errors"
	"cicd/operators/helga/internal/utils"
	"cicd/operators/helga/internal/vars"
	"cicd/operators/helga/models"
	"errors"
	"net/url"
	"os"

	"gopkg.in/yaml.v3"
)

type Global struct {
	artifact *models.Artifact `yaml:"artifact"`
}

type Config struct {
	Global   *Global           `yaml:"global"`
	Clusters []*models.Cluster `yaml:"clusters"`
}

func (g *Global) Validate() error {
	var (
		validationErr error  = nil
		structName    string = "Global"
	)

	if g.artifact.Domain != "" {
		if _, err := url.ParseRequestURI(g.artifact.Domain); err != nil {
			validationErr = &helga_errors.ErrValidation{StructName: structName, DerivedFromErr: err}
		}
	}

	if len(g.artifact.Repos) > 0 {
		errs, filteredRepos := utils.FilterByValidation(utils.ToValidatableSlice(g.artifact.Repos), "repo: %v, did not pass validation, changing availability to false")
		helga_errors.HandleErrors(errs)

		g.artifact.Repos, validationErr = utils.FromValidatableSlice[*models.Repo](filteredRepos)
	}

	helga_errors.HandleError(validationErr)

	return validationErr
}

func (c *Config) validate() error {
	var (
		validationErr error  = nil
		structName    string = "Config"
	)

	if err := c.Global.Validate(); err != nil {
		validationErr = &helga_errors.ErrValidation{StructName: structName, DerivedFromErr: err}
	}

	errs, filteredClusters := utils.FilterByValidation(utils.ToValidatableSlice(c.Clusters), "namespace: %v did not pass validation, changing availability to false")
	helga_errors.HandleErrors(errs)

	c.Clusters, validationErr = utils.FromValidatableSlice[*models.Cluster](filteredClusters)

	if len(c.Clusters) == 0 {
		validationErr = &helga_errors.ErrValidation{StructName: structName, DerivedFromErr: errors.New("clusters list cannot be empty")}
	}

	helga_errors.HandleError(validationErr)

	return validationErr
}

func (c *Config) syncClusterArtifactsWithGlobal() {
	for _, cl := range c.Clusters {
		for _, ns := range cl.Namespaces {
			ns.Artifact.Sync(c.Global.artifact)
		}
	}
}

func (c *Config) syncWithGlobal() {
	c.syncClusterArtifactsWithGlobal()
}

func (c *Config) UnmarshalYAMLConfig() error {
	var configErr error

	// validate correct unmarshelling
	f, err := os.ReadFile(vars.HELGA_CONF_FILE_PATH)

	if err != nil {
		configErr = &helga_errors.ErrConfigLoadingError{DerivedFromErr: err}
		helga_errors.HandleError(configErr)

		return configErr
	}

	if err := yaml.Unmarshal(f, c); err != nil {
		configErr = &helga_errors.ErrConfigLoadingError{DerivedFromErr: err}
		helga_errors.HandleError(configErr)

		return configErr
	}

	c.syncWithGlobal()

	return c.validate()
}
