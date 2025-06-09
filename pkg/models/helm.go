package models

import (
	"fmt"
	"strings"
	"time"

	helga_errors "github.com/fennet82/helga/pkg/errors"
	"golang.org/x/mod/semver"
	"helm.sh/helm/v3/pkg/release"
)

type HelmChart interface {
	Validate() error
	Name() string
	Version() string
	Time() time.Time
}

// helm chart fetched from namespace by go-helm-client
type ArtifactHelmPackage struct {
	Repo         string    `yaml:"repo"`
	Path         string    `yaml:"path"`
	FullName     string    `json:"name"`
	TimeModified time.Time `json:"modified"`
}

func (ahp ArtifactHelmPackage) Validate() error {
	if ahp.FullName == "" {
		err := fmt.Errorf("package: %+v name is invalid please check again", ahp)

		return err
	}

	return nil
}

func (ahp ArtifactHelmPackage) Name() string {
	s := strings.Split(strings.TrimSuffix(ahp.FullName, ".tgz"), "-")
	pkgName := strings.Join(s[:len(s)-1], "-")

	return pkgName
}

func (ahp ArtifactHelmPackage) Version() string {
	s := strings.Split(strings.TrimSuffix(ahp.FullName, ".tgz"), "-")
	pkgVersion := s[len(s)-1]

	return pkgVersion
}

func (hri ArtifactHelmPackage) Time() time.Time {
	return hri.TimeModified
}

// helm chart fetched from namespace by go-helm-client
type HelmReleaseInfo struct {
	release.Release
}

func (hri HelmReleaseInfo) Validate() error {
	return nil
}

func (hri HelmReleaseInfo) Name() string {
	return hri.Chart.Metadata.Name
}

func (hri HelmReleaseInfo) Version() string {
	return hri.Chart.Metadata.Version
}

func (hri HelmReleaseInfo) Time() time.Time {
	return hri.Info.LastDeployed.Time
}

func DetermineNewerPkg(pkgA HelmChart, pkgB HelmChart, decideByVersion bool) (HelmChart, error) {
	if err := pkgA.Validate(); err != nil {
		return nil, err
	}

	if err := pkgB.Validate(); err != nil {
		return nil, err
	}

	retPkg := pkgB

	if pkgA.Name() != pkgB.Name() {
		return nil, &helga_errors.ErrPkgsDoNotMatch{
			ErrMsg: fmt.Sprintf("pkg's names: %s and %s do not match", pkgA.Name(), pkgB.Name()),
		}
	}

	if decideByVersion && semver.Compare("v"+pkgA.Version(), "v"+pkgB.Version()) > 0 {
		retPkg = pkgA
	} else if pkgA.Time().After(pkgB.Time()) {
		retPkg = pkgA
	}

	return retPkg, nil
}
