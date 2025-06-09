package models

import (
	"context"
	"fmt"
	"time"

	"github.com/fennet82/helga/internal/logger"
	"github.com/fennet82/helga/internal/vars"
	helga_errors "github.com/fennet82/helga/pkg/errors"
	helmclient "github.com/mittwald/go-helm-client"
	"slices"
)

type Namespace struct {
	Name         string    `yaml:"name"`
	SyncInterval uint16    `yaml:"sync_interval"`
	Artifact     *Artifact `yaml:"artifact"`
	helmClient   helmclient.Client
}

func (ns *Namespace) String() string {
	return ns.Name
}

func (ns *Namespace) Validate() []error {
	logger.GetLoggerInstance().Info(fmt.Sprintf("starting validation for namespace: %s", ns.String()))

	var (
		validationErrs []error
		structName     = "Namespace"
	)

	if ns.Name == "" {
		validationErrs = append(validationErrs, helga_errors.ErrValidation{StructName: structName, DerivedFromErr: fmt.Errorf("namespace name cannot be empty")})
	}

	if ns.SyncInterval < vars.SYNC_INTERVAL_DEFAULT_RETENTION {
		validationErrs = append(validationErrs, helga_errors.ErrValidation{StructName: structName, DerivedFromErr: fmt.Errorf("sync interval for namespace: %s, needs to be above: %d currently: %d", ns.Name, vars.SYNC_INTERVAL_DEFAULT_RETENTION, ns.SyncInterval)})
	}

	if errs := ns.Artifact.Validate(); len(errs) > 0 {
		validationErrs = append(validationErrs, helga_errors.ErrValidation{StructName: structName, DerivedFromErr: fmt.Errorf("error artifact with domain: %s, did not pass validation", ns.Artifact.Domain)})
	}

	helga_errors.HandleErrors(validationErrs)

	return validationErrs
}

func (ns *Namespace) addOrUpdateHelmRepos() {
	entries := ns.Artifact.GetArtifactReposAsEntries()

	for i, e := range entries {
		err := ns.helmClient.AddOrUpdateChartRepo(e)
		if err != nil {
			helga_errors.HandleError(fmt.Errorf(
				"error occured while trying to insert entry repo: %s, removing from repo targets, derived from err: %w", e.Name, err,
			))

			ns.Artifact.Repos = slices.Delete(ns.Artifact.Repos, i, i+1)
		}
	}
}

func (ns *Namespace) getDeployedReleases() ([]HelmChart, error) {
	releases, err := ns.helmClient.ListDeployedReleases()
	if err != nil {
		return nil, err
	}

	HelmReleaseInfoList := make([]HelmChart, len(releases))
	for i, rel := range releases {
		HelmReleaseInfoList[i] = HelmReleaseInfo{Release: *rel}
	}

	return HelmReleaseInfoList, nil
}

func (ns *Namespace) syncHelmPackages() (releasesToDelete []HelmChart, chartsToDeploy []HelmChart, err error) {
	releasesToDelete = nil
	chartsToDeploy = nil
	err = nil

	logger.GetLoggerInstance().Info(fmt.Sprintf("starting sync between releases and artifactory charts for namespace: %s", ns.String()))

	deployedReleases, err := ns.getDeployedReleases()
	if err != nil {
		err = fmt.Errorf("couldnt get deployed releases for namespace: %s", ns.String())
		return
	}

	artifactoryPkgsMap := ns.Artifact.GetChartPkgsInArtifact()
	if len(artifactoryPkgsMap) == 0 {
		err = fmt.Errorf("pkgs map recieved from artifactory for namespace: %s, is empty", ns.String())
		return
	}

	for _, rel := range deployedReleases {
		artifactoryHelmPkg, exists := artifactoryPkgsMap[rel.Name()]
		if exists {
			decideByVersion := ns.Artifact.GetRepoByName(artifactoryHelmPkg.(ArtifactHelmPackage).Repo).DecideByVersion

			pkg, err := DetermineNewerPkg(rel, artifactoryHelmPkg, decideByVersion)
			if err != nil {
				helga_errors.HandleError(fmt.Errorf("error occured while syncing pkgs for namespace: %s, err: %w", ns.Name, err))
				continue
			}

			pkgPtr, ok1 := pkg.(*ArtifactHelmPackage)
			ahpPtr, ok2 := artifactoryHelmPkg.(*ArtifactHelmPackage)

			if ok1 && ok2 && pkgPtr == ahpPtr {
				chartsToDeploy = append(chartsToDeploy, artifactoryHelmPkg)
			}
		} else {
			releasesToDelete = append(releasesToDelete, rel)
		}
	}

	return
}

func (ns *Namespace) SyncHelmPkgsWithCluster() {
	// it's important to notice that for now we dont delete releases from the cluster but it can easily be implemented in the code
	for {
		func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Println("panic occured: ", r)
				}

				time.Sleep(time.Duration(ns.SyncInterval) * time.Second)
			}()

			_, chartsToDeploy, err := ns.syncHelmPackages()
			if err != nil {
				panic(fmt.Errorf("couldnt sync pkgs on namespace: %s, because error occured in the sync pkgs. err: %w", ns.String(), err))
			}

			for _, pkg := range chartsToDeploy {
				ahp := pkg.(ArtifactHelmPackage)

				chartSpec := helmclient.ChartSpec{
					ReleaseName: ahp.Name(),
					ChartName:   ns.Artifact.Domain + "/" + ahp.Repo + "/" + ahp.Path + "/" + ahp.FullName,
					Namespace:   ns.Name,
					UpgradeCRDs: true,
					Wait:        true,
					Timeout:     30 * time.Second,
				}

				if _, err := ns.helmClient.InstallOrUpgradeChart(context.Background(), &chartSpec, nil); err != nil {
					helga_errors.HandleError(fmt.Errorf("error installing/upgrading chart: %s", chartSpec.ChartName))
				}
			}
		}()
	}
}
