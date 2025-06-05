package models

import (
	"errors"
	"fmt"

	helga_errors "github.com/fennet82/helga/pkg/errors"
)

type Repo struct {
	Name            string   `yaml:"name"`
	DecideByVersion bool     `yaml:"decideByVersion"` // will decide by date if not true
	Paths           []string `yaml:"paths"`
}

func (r *Repo) String() string {
	return r.Name
}

func (r *Repo) Validate() []error {
	var (
		validationErrs []error
		structName     = "Repo"
	)

	if r.Name == "" {
		validationErrs = append(validationErrs, helga_errors.ErrValidation{StructName: structName, DerivedFromErr: errors.New("repo name cannot be empty")})
	}

	if len(r.Paths) == 0 {
		validationErrs = append(validationErrs, helga_errors.ErrValidation{StructName: structName, DerivedFromErr: errors.New("length of repo paths list cannot be empty")})
	}

	helga_errors.HandleErrors(validationErrs)

	return validationErrs
}

func (dest *Repo) Sync(src *Repo) error {
	if dest == nil || src == nil {
		return fmt.Errorf("cannot sync nil Repo objects")
	}

	if dest.Name != src.Name {
		return fmt.Errorf("repo: %s could be synced with repo: %s. repos name do not match", dest.Name, src.Name)
	}

	if src.DecideByVersion && !dest.DecideByVersion {
		dest.DecideByVersion = src.DecideByVersion
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
