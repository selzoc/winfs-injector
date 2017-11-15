package tile

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"
)

type TileInjector struct{}

func NewTileInjector() TileInjector {
	return TileInjector{}
}

func (i TileInjector) AddReleaseToMetadata(releasePath, releaseName, releaseVersion, tileDir string) error {
	releaseFileName := filepath.Base(releasePath)

	metadataGlob := filepath.Join(tileDir, "metadata", "*.yml")
	yamlFiles, err := filepath.Glob(metadataGlob)
	if err != nil {
		return err
	}
	if yamlFiles == nil {
		return fmt.Errorf("expected to find a product metadata file matching path '%s', but found none", metadataGlob)
	}
	if len(yamlFiles) > 1 {
		return fmt.Errorf("expected to find a single metadata file matching path '%s', but found multiple", metadataGlob)
	}
	metadataFilePath := yamlFiles[0]

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
