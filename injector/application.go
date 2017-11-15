package injector

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
)

//go:generate counterfeiter -o ./fakes/extractor.go --fake-name Extractor . extractor

type extractor interface {
	ExtractWindowsFSRelease(inputTile, outputDir string) (string, error)
	ExtractTile(inputTile, outputDir string) error
}

//go:generate counterfeiter -o ./fakes/injector.go --fake-name Injector . injector

type injector interface {
	AddReleaseToTile(releasePath, releaseName, releaseVersion, tileDir string) error
}

//go:generate counterfeiter -o ./fakes/zipper.go --fake-name Zipper . zipper

type zipper interface {
	Zip(dir, zipFile string) error
}

//go:generate counterfeiter -o ./fakes/release_creator.go --fake-name ReleaseCreator . releaseCreator

type releaseCreator interface {
	CreateRelease(imageName, releaseDir, tarballPath, imageTagPath, versionDataPath, outputDir string) error
}

type Application struct {
	extractor      extractor
	injector       injector
	releaseCreator releaseCreator
	zipper         zipper
}

func NewApplication(extractor extractor, releaseCreator releaseCreator, injector injector, zipper zipper) Application {
	return Application{
		extractor:      extractor,
		injector:       injector,
		releaseCreator: releaseCreator,
		zipper:         zipper,
	}
}

func (a Application) Run(inputTile, workingDir string) error {
	destDir, err := a.extractor.ExtractWindowsFSRelease(inputTile, workingDir)
	if err != nil {
		return err
	}

	releaseName := "windows2016fs"
	imageName := "cloudfoundry/windows2016fs"
	releaseDir := filepath.Join(destDir, "embed", "windows2016fs-release")
	versionDataPath := filepath.Join(releaseDir, "VERSION")
	winfsBlobsDir := filepath.Join(releaseDir, "blobs", "windows2016fs")
	imageTagPath := filepath.Join(releaseDir, "src", "code.cloudfoundry.org", "windows2016fs", "IMAGE_TAG")

	rawReleaseVersion, err := ioutil.ReadFile(versionDataPath)
	if err != nil {
		return err
	}

	releaseVersion := strings.TrimSuffix(string(rawReleaseVersion), "\n")

	tarballPath := filepath.Join(
		workingDir,
		fmt.Sprintf("%s-%s.tgz", releaseName, releaseVersion),
	)

	err = a.releaseCreator.CreateRelease(imageName, releaseDir, tarballPath, imageTagPath, versionDataPath, winfsBlobsDir)
	if err != nil {
		return err
	}

	tileDir := filepath.Join(workingDir, "extracted-tile")
	err = a.extractor.ExtractTile(inputTile, tileDir)
	if err != nil {
		return err
	}

	err = a.injector.AddReleaseToTile(tarballPath, releaseName, releaseVersion, tileDir)
	if err != nil {
		return err
	}

	return a.zipper.Zip(tileDir, "CHANGE_ME.zip")
}
