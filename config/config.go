package config

import (
	"errors"
	"fmt"
	"os"
	"regexp"

	helga_errors "cicd/operators/helga/errors"
	"cicd/operators/helga/internal/utils"
	"cicd/operators/helga/internal/vars"
	"cicd/operators/helga/models"

	"gopkg.in/yaml.v3"
)

type Global struct {
	cluster  *models.Cluster  `yaml:"cluster"`
	artifact *models.Artifact `yaml:"artifact"`
}

type Config struct {
	Global   *Global           `yaml:"global"`
	Clusters []*models.Cluster `yaml:"clusters"`
}

func (g *Global) Validate() []error {
	var (
		validationErrs []error
		structName     = "Global"
		dReg           = regexp.MustCompile(vars.DOMAIN_VALIDATION_REGEX)
	)

	// artifact validation
	if !dReg.MatchString(g.artifact.Domain) {
		validationErrs = append(validationErrs, helga_errors.ErrValidation{StructName: structName, DerivedFromErr: fmt.Errorf(
			"domain: %s, did not pass regex validation please refer to this regex for fixing: %s", g.artifact.Domain, vars.DOMAIN_VALIDATION_REGEX,
		)})
	}

	if len(g.artifact.Repos) > 0 {
		errs, filteredRepos := utils.FilterByValidation(utils.ToValidatableSlice(g.artifact.Repos), "repo: %v, did not pass validation, changing availability to false")
		helga_errors.HandleErrors(errs)

		g.artifact.Repos = utils.FromValidatableSlice[*models.Repo](filteredRepos)
	}

	// cluster validation
	if g.cluster.Name == "" {
		validationErrs = append(validationErrs, helga_errors.ErrValidation{StructName: structName, DerivedFromErr: errors.New("name field cannot be empty")})
	}

	if !dReg.MatchString(g.cluster.Server) {
		validationErrs = append(validationErrs, helga_errors.ErrValidation{StructName: structName, DerivedFromErr: fmt.Errorf(
			"server: %s, did not pass regex validation please refer to this regex for fixing: %s", g.cluster.Server, vars.DOMAIN_VALIDATION_REGEX,
		)})
	}

	if g.cluster.Username == "" {
		validationErrs = append(validationErrs, helga_errors.ErrValidation{StructName: structName, DerivedFromErr: errors.New("username field cannot be empty")})
	}

	if g.cluster.Password == "" {
		validationErrs = append(validationErrs, helga_errors.ErrValidation{StructName: structName, DerivedFromErr: errors.New("password field cannot be empty")})
	}

	helga_errors.HandleErrors(validationErrs)

	return validationErrs
}

func (c *Config) validate() []error {
	var (
		validationErrs []error
		structName     = "Config"
	)

	if err := c.Global.Validate(); err != nil {
		validationErrs = append(validationErrs, helga_errors.ErrValidation{StructName: structName, DerivedFromErr: fmt.Errorf("global did not pass validation refer to logs and fix")})
	}

	errs, filteredClusters := utils.FilterByValidation(utils.ToValidatableSlice(c.Clusters), "namespace: %v did not pass validation, changing availability to false")
	helga_errors.HandleErrors(errs)

	c.Clusters = utils.FromValidatableSlice[*models.Cluster](filteredClusters)

	if len(c.Clusters) == 0 {
		validationErrs = append(validationErrs, helga_errors.ErrValidation{StructName: structName, DerivedFromErr: errors.New("clusters list cannot be empty")})
	}

	helga_errors.HandleErrors(validationErrs)

	return validationErrs
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

func (c *Config) UnmarshalYAMLConfig() []error {
	var configErrs []error

	// validate correct unmarshelling
	f, err := os.ReadFile(vars.HELGA_CONF_FILE_PATH)
	if err != nil {
		configErrs = append(configErrs, helga_errors.ErrConfigLoadingError{DerivedFromErr: err})
		helga_errors.HandleErrors(configErrs)

		return configErrs
	}

	if err := yaml.Unmarshal(f, c); err != nil {
		configErrs = append(configErrs, helga_errors.ErrConfigLoadingError{DerivedFromErr: err})
		helga_errors.HandleErrors(configErrs)

		return configErrs
	}

	c.syncWithGlobal()

	return c.validate()
}
