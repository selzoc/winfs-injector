package injector

import (
	"io/ioutil"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"
)

type TileInjector struct {
}

func NewTileInjector() TileInjector {
	return TileInjector{}
}

func (i TileInjector) AddReleaseToTile(releasePath, releaseName, releaseVersion, tileDir string) error {
	releaseDir := filepath.Join(tileDir, "releases")
	err := os.MkdirAll(releaseDir, 0755)
	if err != nil {
		return err
	}

	releaseFileName := filepath.Base(releasePath)

	targetReleasePath := filepath.Join(releaseDir, releaseFileName)
	err = os.Rename(releasePath, targetReleasePath)
	if err != nil {
		return err
	}

	metadataFilePath := filepath.Join(tileDir, "metadata.yml")

	f, err := os.Open(metadataFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	var originalMetadata Metadata
	err = yaml.Unmarshal(data, &originalMetadata)
	if err != nil {
		return err
	}

	originalMetadata.Releases = append(originalMetadata.Releases, Release{
		Name:    releaseName,
		Version: releaseVersion,
		File:    releaseFileName,
	})

	contents, err := yaml.Marshal(&originalMetadata)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(metadataFilePath, contents, 0644)
}
