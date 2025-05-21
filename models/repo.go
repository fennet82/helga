package models

import (
	helga_errors "cicd/operators/helga/errors"
	"errors"
)

type Repo struct {
	Name    string   `yaml:"name"`
	Paths   []string `yaml:"paths"`
}

func (r *Repo) Validate() error {
	var (
		validationErr error  = nil
		structName    string = "Repo"
	)

	if r.Name == "" {
		validationErr = &helga_errors.ErrValidation{StructName: structName, DerivedFromErr: errors.New("repo name cannot be empty")}
	}

	if len(r.Paths) == 0 {
		validationErr = &helga_errors.ErrValidation{StructName: structName, DerivedFromErr: errors.New("length of repo paths list cannot be empty")}
	}

	helga_errors.HandleError(validationErr)

	return validationErr
}
