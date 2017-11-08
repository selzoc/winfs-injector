package injector

import (
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"
)

type Application struct {
	extractor      extractor
	releaseCreator releaseCreator
}

//go:generate counterfeiter -o ./fakes/extractor.go --fake-name Extractor . extractor
type extractor interface {
	ExtractWindowsFSRelease(inputTile string, outputDir string) (string, error)
}

//go:generate counterfeiter -o ./fakes/release_creator.go --fake-name ReleaseCreator . releaseCreator
type releaseCreator interface {
	CreateRelease(imageName, releaseDir, tarballPath, imageTagPath, versionDataPath, outputDir string) error
}

func NewApplication(extractor extractor, releaseCreator releaseCreator) Application {
	return Application{
		extractor:      extractor,
		releaseCreator: releaseCreator,
	}
}

func (a Application) Run(inputTile, workingDir string) error {
	destDir, err := a.extractor.ExtractWindowsFSRelease(inputTile, workingDir)
	if err != nil {
		return err
	}

	imageName := "cloudfoundry/windows2016fs"

	tileName := path.Base(inputTile)
	tileName = strings.Replace(tileName, path.Ext(tileName), "", 1)

	releaseDir := filepath.Join(destDir, tileName, "embed", "windows2016fs-release")

	imageTagPath := filepath.Join(releaseDir, "src", "code.cloudfoundry.org", "windows2016fs", "IMAGE_TAG")
	content, err := ioutil.ReadFile(imageTagPath)
	if err != nil {
		return err
	}

	versionDataPath := filepath.Join(releaseDir, "VERSION")

	winfsBlobsDir := filepath.Join(releaseDir, "blobs", "windows2016fs")

	tarballPath := filepath.Join(workingDir, fmt.Sprintf("windows2016fs-%s.tgz", strings.TrimSuffix(string(content), "\n")))

	err = a.releaseCreator.CreateRelease(imageName, releaseDir, tarballPath, imageTagPath, versionDataPath, winfsBlobsDir)
	if err != nil {
		return err
	}

	return nil
}
