package models

import (
	"errors"
	"fmt"

	helga_errors "cicd/operators/helga/errors"
)

type Namespace struct {
	Name     string    `yaml:"name"`
	Artifact *Artifact `yaml:"artifact"`
}

func (ns *Namespace) Validate() []error {
	var (
		validationErrs []error
		structName     = "Namespace"
	)

	if ns.Name == "" {
		validationErrs = append(validationErrs, helga_errors.ErrValidation{StructName: structName, DerivedFromErr: errors.New("namespace name cannot be empty")})
	}

	if errs := ns.Artifact.Validate(); errs != nil {
		validationErrs = append(validationErrs, helga_errors.ErrValidation{StructName: structName, DerivedFromErr: fmt.Errorf("error artifact %s did not pass validation", ns.Artifact.Domain)})
	}

	helga_errors.HandleErrors(validationErrs)

	return validationErrs
}

func (ns *Namespace) OrganizeHelmPackages() []*ArtifactHelmPackage {
	pkgsMapped := map[string]*ArtifactHelmPackage{}
	finalPkgs := []*ArtifactHelmPackage{}

	for _, pkg := range ns.Artifact.FetchHelmPacakgesByAQL() {
		n, _, err := pkg.GetNameAndVersion()
		if err != nil {
			helga_errors.HandleError(err)

			continue
		}

		if currPkg, exists := pkgsMapped[n]; exists {
			p, err := DetermineNewerPkg(pkg, currPkg, *ns.Artifact.DecideByVersion)
			if err != nil {
				helga_errors.HandleError(err)

				continue
			}

			pkgsMapped[n] = p
		} else {
			pkgsMapped[n] = pkg
		}
	}

	for _, pkg := range pkgsMapped {
		finalPkgs = append(finalPkgs, pkg)
	}

	return finalPkgs
}
