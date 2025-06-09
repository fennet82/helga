package config

import (
	"errors"
	"fmt"
	"os"
	"regexp"

	"github.com/fennet82/helga/internal/logger"
	"github.com/fennet82/helga/internal/utils"
	"github.com/fennet82/helga/internal/vars"
	helga_errors "github.com/fennet82/helga/pkg/errors"
	"github.com/fennet82/helga/pkg/models"
	"gopkg.in/yaml.v2"
)

type Global struct {
	Cluster  *models.Cluster  `yaml:"cluster"`
	Artifact *models.Artifact `yaml:"artifact"`
}

func (g *Global) Validate() []error {
	logger.GetLoggerInstance().Info("starting validation for global")

	var (
		validationErrs []error
		structName     = "Global"
		dReg           = regexp.MustCompile(vars.ARTIFACTORY_VALIDATION_REGEX)
	)

	// artifact validation
	if !dReg.MatchString(g.Artifact.Domain) {
		validationErrs = append(validationErrs, helga_errors.ErrValidation{StructName: structName, DerivedFromErr: fmt.Errorf(
			"domain: %s, did not pass regex validation please refer to this regex for fixing: %s", g.Artifact.Domain, vars.ARTIFACTORY_VALIDATION_REGEX,
		)})
	}

	if len(g.Artifact.Repos) > 0 {
		errs, filteredRepos := utils.FilterByValidation(utils.ToValidatableSlice(g.Artifact.Repos), "repo: %s, did not pass validation, changing availability to false")
		helga_errors.HandleErrors(errs)

		g.Artifact.Repos = utils.FromValidatableSlice[*models.Repo](filteredRepos)
	}

	helga_errors.HandleErrors(validationErrs)

	return validationErrs
}

type Config struct {
	Global   *Global           `yaml:"global"`
	Clusters []*models.Cluster `yaml:"clusters"`
}

func (c *Config) validate() []error {
	var (
		validationErrs []error
		structName     = "Config"
	)

	if errs := c.Global.Validate(); len(errs) > 0 {
		validationErrs = append(validationErrs, helga_errors.ErrValidation{StructName: structName, DerivedFromErr: fmt.Errorf("global did not pass validation refer to logs and fix")})
	}

	errs, filteredClusters := utils.FilterByValidation(utils.ToValidatableSlice(c.Clusters), "cluster: %s did not pass validation, changing availability to false")
	helga_errors.HandleErrors(errs)

	c.Clusters = utils.FromValidatableSlice[*models.Cluster](filteredClusters)

	if len(c.Clusters) == 0 {
		validationErrs = append(validationErrs, helga_errors.ErrValidation{StructName: structName, DerivedFromErr: errors.New("clusters list cannot be empty")})
	}

	helga_errors.HandleErrors(validationErrs)

	return validationErrs
}

func (c *Config) syncWithGlobal() (errs []error) {
	for _, cl := range c.Clusters {
		err := cl.Sync(c.Global.Cluster)
		if err != nil {
			errs = append(errs, helga_errors.ErrSync{DerivedFromErr: err})
		}

		for _, ns := range cl.Namespaces {
			err := ns.Artifact.Sync(c.Global.Artifact)
			if err != nil {
				errs = append(errs, helga_errors.ErrSync{DerivedFromErr: err})
			}
		}
	}

	return
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

	sync_errs := c.syncWithGlobal()
	if len(sync_errs) > 0 {
		helga_errors.HandleErrors(sync_errs)
	}

	return c.validate()
}
