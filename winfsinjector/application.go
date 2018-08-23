package winfsinjector

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	readFile  = ioutil.ReadFile
	removeAll = os.RemoveAll
	readDir   = ioutil.ReadDir
)

type Application struct {
	injector       injector
	releaseCreator releaseCreator
	zipper         zipper
}

//go:generate counterfeiter -o ./fakes/file_info.go --fake-name FileInfo os.FileInfo

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
	CreateRelease(releaseName, imageName, releaseDir, tarballPath, imageTagPath, version string) error
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

	// find what the embedded directory is
	files, err := readDir(filepath.Join(extractedTileDir, "embed"))
	if err != nil {
		return err
	}

	if len(files) > 1 {
		return errors.New("there is more than one file system embedded in the tile; please contact the tile authors to fix")
	} else if len(files) == 0 {
		return errors.New("there is no file system embedded in the tile; please contact the tile authors to fix")
	}

	e := files[0]
	if !e.IsDir() {
		return errors.New("the embedded file system is not a directory; please contact the tile authors to fix")
	}

	embeddedReleaseDir := e.Name()
	releaseVersion, err := a.extractReleaseVersion(embeddedReleaseDir)
	if err != nil {
		return err
	}

	if runtime.GOOS == "windows" {
		cmd := exec.Command("git", "config", "core.filemode", "false")
		cmd.Dir = embeddedReleaseDir
		stdoutStderr, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("unable to fix file permissions for windows: %s, %s", stdoutStderr, err)
		}

		cmd = exec.Command("git", "submodule", "foreach", "git", "config", "core.filemode", "false")
		cmd.Dir = embeddedReleaseDir
		stdoutStderr, err = cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("unable to fix file permissions for windows: %s, %s", stdoutStderr, err)
		}
	}

	// Dependent on what the tile metadata expects, p-windows-runtime-2016/jobs/windows1803fs.yml
	releaseName := "windows1803fs"
	imageName := "cloudfoundry/windows2016fs"
	imageTagPath := filepath.Join(embeddedReleaseDir, "src", "code.cloudfoundry.org", "windows2016fs", "1803", "IMAGE_TAG")
	tarballPath := filepath.Join(extractedTileDir, "releases", fmt.Sprintf("%s-%s.tgz", releaseName, releaseVersion))

	err = a.releaseCreator.CreateRelease(releaseName, imageName, embeddedReleaseDir, tarballPath, imageTagPath, releaseVersion)
	if err != nil {
		return err
	}

	err = a.injector.AddReleaseToMetadata(tarballPath, releaseName, releaseVersion, extractedTileDir)
	if err != nil {
		return err
	}

	err = removeAll(embeddedReleaseDir)
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
