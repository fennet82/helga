package models

import (
	helga_errors "cicd/operators/helga/errors"
	"cicd/operators/helga/internal/utils"
	"errors"
	"net/url"
)

type Cluster struct {
	Name       string       `yaml:"name"`
	Server     string       `yaml:"server"`
	Username   string       `yaml:"username"`
	Password   string       `yaml:"password"`
	Namespaces []*Namespace `yaml:"namespaces"`
}

func (c *Cluster) Validate() error {
	var (
		validationErr error  = nil
		structName    string = "Cluster"
	)

	if c.Name == "" {
		validationErr = &helga_errors.ErrValidation{StructName: structName, DerivedFromErr: errors.New("name field cannot be empty")}
	}

	if _, err := url.ParseRequestURI(c.Server); err != nil {
		validationErr = &helga_errors.ErrValidation{StructName: structName, DerivedFromErr: err}
	}

	if c.Username == "" {
		validationErr = &helga_errors.ErrValidation{StructName: structName, DerivedFromErr: errors.New("username field cannot be empty")}
	}

	if c.Password == "" {
		validationErr = &helga_errors.ErrValidation{StructName: structName, DerivedFromErr: errors.New("password field cannot be empty")}
	}

	errs, filteredNameSpace := utils.FilterByValidation(utils.ToValidatableSlice(c.Namespaces), "namespace: %v did not pass validation, changing availability to false")
	helga_errors.HandleErrors(errs)

	c.Namespaces, validationErr = utils.FromValidatableSlice[*Namespace](filteredNameSpace)

	if len(c.Namespaces) == 0 {
		validationErr = &helga_errors.ErrValidation{StructName: structName, DerivedFromErr: errors.New("namespaces list cannot be empty")}
	}

	helga_errors.HandleError(validationErr)

	return validationErr
}
