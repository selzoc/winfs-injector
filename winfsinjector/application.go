package winfsinjector

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var (
	readFile  = ioutil.ReadFile
	removeAll = os.RemoveAll
)

type Application struct {
	injector       injector
	releaseCreator releaseCreator
	zipper         zipper
}

//go:generate counterfeiter -o ./fakes/injector.go --fake-name Injector . injector

type injector interface {
	AddReleaseToMetadata(releasePath, releaseName, releaseVersion, extractedTileDir string) error
}

//go:generate counterfeiter -o ./fakes/zipper.go --fake-name Zipper . zipper

type zipper interface {
	Zip(dir, zipFile string) error
	Unzip(zipFile, dest string) error
}

//go:generate counterfeiter -o ./fakes/release_creator.go --fake-name ReleaseCreator . releaseCreator

type releaseCreator interface {
	CreateRelease(imageName, releaseDir, tarballPath, imageTagPath, versionDataPath, outputDir string) error
}

func NewApplication(releaseCreator releaseCreator, injector injector, zipper zipper) Application {
	return Application{
		injector:       injector,
		releaseCreator: releaseCreator,
		zipper:         zipper,
	}
}

func (a Application) Run(inputTile, outputTile, workingDir string) error {
	if inputTile == "" {
		return errors.New("--input-tile is required")
	}

	if outputTile == "" {
		return errors.New("--output-tile is required")
	}

	extractedTileDir := filepath.Join(workingDir, "extracted-tile")
	err := a.zipper.Unzip(inputTile, extractedTileDir)
	if err != nil {
		return err
	}

	releaseDir := filepath.Join(extractedTileDir, "embed", "windows2016fs-release")
	releaseVersion, err := a.extractReleaseVersion(releaseDir)
	if err != nil {
		return err
	}

	releaseName := "windows2016fs"
	imageName := "cloudfoundry/windows2016fs"
	winfsBlobsDir := filepath.Join(releaseDir, "blobs", "windows2016fs")
	imageTagPath := filepath.Join(releaseDir, "src", "code.cloudfoundry.org", "windows2016fs", "IMAGE_TAG")
	tarballPath := filepath.Join(extractedTileDir, "releases", fmt.Sprintf("%s-%s.tgz", releaseName, releaseVersion))

	err = a.releaseCreator.CreateRelease(imageName, releaseDir, tarballPath, imageTagPath, filepath.Join(releaseDir, "VERSION"), winfsBlobsDir)
	if err != nil {
		return err
	}

	err = a.injector.AddReleaseToMetadata(tarballPath, releaseName, releaseVersion, extractedTileDir)
	if err != nil {
		return err
	}

	err = removeAll(filepath.Join(extractedTileDir, "embed", "windows2016fs-release"))
	if err != nil {
		return err
	}

	return a.zipper.Zip(extractedTileDir, outputTile)
}

func (a Application) extractReleaseVersion(releaseDir string) (string, error) {
	rawReleaseVersion, err := readFile(filepath.Join(releaseDir, "VERSION"))
	if err != nil {
		return "", err
	}

	return strings.TrimSuffix(string(rawReleaseVersion), "\n"), nil
}
