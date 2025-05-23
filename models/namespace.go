package models

import (
	helga_errors "cicd/operators/helga/errors"
	"errors"
)

type Namespace struct {
	Name     string    `yaml:"name"`
	Artifact *Artifact `yaml:"artifact"`
}

func (ns *Namespace) Validate() error {
	var (
		validationErr error  = nil
		structName    string = "Namespace"
	)

	if ns.Name == "" {
		validationErr = &helga_errors.ErrValidation{StructName: structName, DerivedFromErr: errors.New("repo name cannot be empty")}
	}

	if err := ns.Artifact.Validate(); err != nil {
		validationErr = &helga_errors.ErrValidation{StructName: structName, DerivedFromErr: err}
	}

	helga_errors.HandleError(validationErr)

	return validationErr
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
