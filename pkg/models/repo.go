package models

import (
	"errors"
	"fmt"

	"github.com/fennet82/helga/internal/logger"
	helga_errors "github.com/fennet82/helga/pkg/errors"
	"helm.sh/helm/v3/pkg/repo"
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
	logger.GetLoggerInstance().Info(fmt.Sprintf("starting validation for repo: %s", r.String()))

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

	syncPathsList(&dest.Paths, &src.Paths)

	return nil
}

func syncPathsList(destPaths, srcPaths *[]string) {
	seenPaths := make(map[string]any)

	for _, p := range append(*destPaths, *srcPaths...) {
		if _, exists := seenPaths[p]; !exists {
			seenPaths[p] = struct{}{}
		}
	}

	*destPaths = make([]string, len(seenPaths))

	i := 0
	for p := range seenPaths {
		(*destPaths)[i] = p
		i++
	}
}

func (r *Repo) GetAsHelmRepoEntry() *repo.Entry {
	return &repo.Entry{
		Name:                  r.Name,
		InsecureSkipTLSverify: true,
		PassCredentialsAll:    false,
	}
}
