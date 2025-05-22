package models

import (
	helga_errors "cicd/operators/helga/errors"
	"errors"
	"fmt"
)

type Repo struct {
	Name  string   `yaml:"name"`
	Paths []string `yaml:"paths"`
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

func (dest *Repo) Sync(src *Repo) error {
	if dest.Name != src.Name {
		return fmt.Errorf("repo: %s could be synced with repo: %s. repos name do not match", dest.Name, src.Name)
	}

	seenPaths := make(map[string]any)

	for _, p := range append(dest.Paths, src.Paths...) {
		if _, exists := seenPaths[p]; !exists {
			seenPaths[p] = struct{}{}
		}
	}

	dest.Paths = make([]string, len(seenPaths))

	i := 0
	for p := range seenPaths {
		dest.Paths[i] = p
		i++
	}

	return nil
}
