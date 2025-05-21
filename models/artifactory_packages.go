package models

import (
	helga_errors "cicd/operators/helga/errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/mod/semver"
)

type ArtifactHelmPackage struct {
	Repo         string `yaml:"repo"`
	Path         string `yaml:"path"`
	Name         string `json:"name"`
	TimeCreated  string `json:"created"`
	TimeModified string `json:"modified"`
}

func (ahp *ArtifactHelmPackage) GetNameAndVersion() (pkgName string, pkgVersion string, err error) {
	if ahp.Name == "" {
		err = fmt.Errorf("package: %v name is invalid please check again", ahp)

		return
	}
	s := strings.Split(ahp.Name, "-")
	pkgVersion = s[len(s)-1]
	pkgName = strings.Join(s[:len(s)-1], "-")

	return
}

func DetermineNewerPkg(pkgA *ArtifactHelmPackage, pkgB *ArtifactHelmPackage, decideByVersion bool) (*ArtifactHelmPackage, error) {
	nameA, verA, _ := pkgA.GetNameAndVersion()
	nameB, verB, _ := pkgB.GetNameAndVersion()
	retPkg := pkgB

	if nameA != nameB {
		return nil, &helga_errors.ErrPkgsDoNotMatch{
			ErrMsg: fmt.Sprintf("pkg's names: %s and %s do not match please determine which newer", nameA, nameB),
		}
	}

	if semver.Compare("v"+verA, "v"+verB) > 0 {
		retPkg = pkgA
	}

	if !decideByVersion {
		tA, _ := time.Parse(time.RFC3339, pkgA.TimeModified)
		tB, _ := time.Parse(time.RFC3339, pkgB.TimeModified)

		if tA.After(tB) {
			retPkg = pkgA
		}
	}

	return retPkg, nil
}
