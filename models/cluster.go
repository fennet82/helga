package models

import (
	"errors"
	"fmt"
	"regexp"

	helga_errors "cicd/operators/helga/errors"
	"cicd/operators/helga/internal/utils"
	"cicd/operators/helga/internal/vars"
)

type Cluster struct {
	Name       string       `yaml:"name"`
	Server     string       `yaml:"server"`
	Username   string       `yaml:"username"`
	Password   string       `yaml:"password"`
	Namespaces []*Namespace `yaml:"namespaces"`
}

func (c *Cluster) Validate() []error {
	var (
		validationErrs []error
		structName     = "Cluster"
		dReg           = regexp.MustCompile(vars.DOMAIN_VALIDATION_REGEX)
	)

	if c.Name == "" {
		validationErrs = append(validationErrs, helga_errors.ErrValidation{StructName: structName, DerivedFromErr: errors.New("name field cannot be empty")})
	}

	if !dReg.MatchString(c.Server) {
		validationErrs = append(validationErrs, helga_errors.ErrValidation{StructName: structName, DerivedFromErr: fmt.Errorf(
			"server: %s, did not pass regex validation please refer to this regex for fixing: %s", c.Server, vars.DOMAIN_VALIDATION_REGEX,
		)})
	}

	if c.Username == "" {
		validationErrs = append(validationErrs, helga_errors.ErrValidation{StructName: structName, DerivedFromErr: errors.New("username field cannot be empty")})
	}

	if c.Password == "" {
		validationErrs = append(validationErrs, helga_errors.ErrValidation{StructName: structName, DerivedFromErr: errors.New("password field cannot be empty")})
	}

	errs, filteredNameSpace := utils.FilterByValidation(utils.ToValidatableSlice(c.Namespaces), "namespace: %v did not pass validation, changing availability to false")
	helga_errors.HandleErrors(errs)

	c.Namespaces = utils.FromValidatableSlice[*Namespace](filteredNameSpace)

	if len(c.Namespaces) == 0 {
		validationErrs = append(validationErrs, helga_errors.ErrValidation{StructName: structName, DerivedFromErr: errors.New("namespaces list cannot be empty")})
	}

	helga_errors.HandleErrors(validationErrs)

	return validationErrs
}
