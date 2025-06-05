package models

import (
	"fmt"
	"strings"
	"time"

	helga_errors "github.com/fennet82/helga/pkg/errors"
	"golang.org/x/mod/semver"
)

type HelmChart interface {
	GetNameAndVersion() (pkgName string, pkgVersion string, err error)
	GetTimeForComparison() time.Time
}

// helm chart fetched from namespace by go-helm-client
type ArtifactHelmPackage struct {
	Repo         string    `yaml:"repo"`
	Path         string    `yaml:"path"`
	Name         string    `json:"name"`
	TimeModified time.Time `json:"modified"`
}

func (ahp ArtifactHelmPackage) GetNameAndVersion() (pkgName string, pkgVersion string, err error) {
	if ahp.Name == "" {
		err = fmt.Errorf("package: %+v name is invalid please check again", ahp)

		return
	}

	s := strings.Split(strings.TrimSuffix(ahp.Name, ".tgz"), "-")
	pkgVersion = s[len(s)-1]
	pkgName = strings.Join(s[:len(s)-1], "-")

	return
}

func (hri ArtifactHelmPackage) GetTimeForComparison() time.Time {
	return hri.TimeModified
}

// helm chart fetched from namespace by go-helm-client
type HelmReleaseInfo struct {
	Namespace string `json:"namespace"`
	Chart     struct {
		Metadata struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"metadata"`
	} `json:"chart"`
	Info struct {
		LastDeployed time.Time `json:"lastDeployed"`
		Status       string    `json:"status"`
	} `json:"info"`
}

func (hri HelmReleaseInfo) GetNameAndVersion() (pkgName string, pkgVersion string, err error) {
	pkgName = hri.Chart.Metadata.Name
	pkgVersion = hri.Chart.Metadata.Version

	return
}

func (hri HelmReleaseInfo) GetTimeForComparison() time.Time {
	return hri.Info.LastDeployed
}

func DetermineNewerPkg(pkgA HelmChart, pkgB HelmChart, decideByVersion bool) (HelmChart, error) {
	nameA, verA, err := pkgA.GetNameAndVersion()
	if err != nil {
		return nil, err
	}

	nameB, verB, err := pkgB.GetNameAndVersion()
	if err != nil {
		return nil, err
	}

	retPkg := pkgB

	if nameA != nameB {
		return nil, &helga_errors.ErrPkgsDoNotMatch{
			ErrMsg: fmt.Sprintf("pkg's names: %s and %s do not match please determine which newer", nameA, nameB),
		}
	}

	if decideByVersion && semver.Compare("v"+verA, "v"+verB) > 0 {
		retPkg = pkgA
	} else if pkgA.GetTimeForComparison().After(pkgB.GetTimeForComparison()) {
		retPkg = pkgA
	}

	return retPkg, nil
}
