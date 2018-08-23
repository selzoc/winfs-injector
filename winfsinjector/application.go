package winfsinjector

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	yaml "gopkg.in/yaml.v2"
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
	embedDirectory := filepath.Join(extractedTileDir, "embed")
	files, err := readDir(embedDirectory)
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

	embeddedReleaseDir := filepath.Join(embedDirectory, e.Name())
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

	releaseName, err := a.extractReleaseName(embeddedReleaseDir)
	if err != nil {
		return err
	}

	imageName := "cloudfoundry/windows2016fs"
	imageTagPath, err := a.determineImageTagPath(releaseName, embeddedReleaseDir)
	if err != nil {
		return err
	}

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

func (a Application) extractReleaseName(releaseDir string) (string, error) {
	contents, err := readFile(filepath.Join(releaseDir, "config", "final.yml"))
	if err != nil {
		return "", err
	}

	type NameFile struct {
		Name string `yaml:"name"`
	}

	var f NameFile
	err = yaml.Unmarshal(contents, &f)
	if err != nil {
		return "", err
	}

	return f.Name, nil
}

func (a Application) determineImageTagPath(releaseName, releaseDir string) (string, error) {
	re, err := regexp.Compile(`windows(\d{4})fs`)
	if err != nil {
		return "", err
	}

	matches := re.FindStringSubmatch(releaseName)
	if len(matches) != 2 {
		return "", errors.New("could not match release name against `windows(\\d4)fs` to determine stemcell line")
	}
	stemcellLine := matches[1]

	var imageTagPath string
	imageTagDirectory := filepath.Join(releaseDir, "src", "code.cloudfoundry.org", "windows2016fs")
	switch stemcellLine {
	case "2016":
		// either windowsfs/IMAGE_TAG or windowsfs/1709/IMAGE_TAG

		files, err := readDir(imageTagDirectory)
		if err != nil {
			return "", err
		}

		for _, f := range files {
			if strings.Contains(f.Name(), "IMAGE_TAG") && f.IsDir() == false {
				return filepath.Join(imageTagDirectory, "IMAGE_TAG"), nil
			}
			if strings.Contains(f.Name(), "1709") && f.IsDir() == true {
				return filepath.Join(imageTagDirectory, "1709", "IMAGE_TAG"), nil
			}
		}

		return "", errors.New("unable to find IMAGE_TAG or 1709/IMAGE_TAG in windows2016fs; please contact tile authors to fix")
	default:
		return filepath.Join(imageTagDirectory, stemcellLine, "IMAGE_TAG"), nil
	}

	return imageTagPath, nil
}
