package models

import (
	"fmt"

	"github.com/fennet82/helga/internal/vars"
	helga_errors "github.com/fennet82/helga/pkg/errors"
)

type Namespace struct {
	Name         string    `yaml:"name"`
	SyncInterval uint16    `yaml:"sync_interval"`
	Artifact     *Artifact `yaml:"artifact"`
}

func (ns *Namespace) String() string {
	return ns.Name
}

func (ns *Namespace) Validate() []error {
	var (
		validationErrs []error
		structName     = "Namespace"
	)

	if ns.Name == "" {
		validationErrs = append(validationErrs, helga_errors.ErrValidation{StructName: structName, DerivedFromErr: fmt.Errorf("namespace name cannot be empty")})
	}

	if ns.SyncInterval <= vars.SYNC_INTERVAL_DEFAULT_RETENTION {
		validationErrs = append(validationErrs, helga_errors.ErrValidation{StructName: structName, DerivedFromErr: fmt.Errorf("sync interval for namespace: %s, needs to be above: %d currently: %d", ns.Name, vars.SYNC_INTERVAL_DEFAULT_RETENTION, ns.SyncInterval)})
	}

	if errs := ns.Artifact.Validate(); len(errs) > 0 {
		validationErrs = append(validationErrs, helga_errors.ErrValidation{StructName: structName, DerivedFromErr: fmt.Errorf("error artifact with domain: %s, did not pass validation", ns.Artifact.Domain)})
	}

	helga_errors.HandleErrors(validationErrs)

	return validationErrs
}
