package injector

import (
	"archive/zip"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
)

type Extractor struct {
	container extractContainer
}

//go:generate counterfeiter -o ./fakes/extract_container.go --fake-name ExtractContainer . extractContainer
type extractContainer interface {
	OpenReader(string) (*zip.ReadCloser, error)
	TempDir(string, string) (string, error)
	MkdirAll(path string, perm os.FileMode) error
	OpenFile(name string, flag int, perm os.FileMode) (*os.File, error)
	Copy(dst io.Writer, src io.Reader) (int64, error)
	Match(pattern, name string) (bool, error)
}

func NewExtractor(container extractContainer) Extractor {
	return Extractor{
		container: container,
	}
}

func (ex Extractor) ExtractWindowsFSRelease(inputTile, outputDir string) (string, error) {
	r, err := ex.container.OpenReader(inputTile)

	if err != nil {
		return "", err
	}

	destDir, err := ex.container.TempDir(outputDir, "windows2016fs")
	if err != nil {
		return "", err
	}

	for _, f := range r.File {
		fileMatch, err := ex.container.Match(filepath.Join("embed", "windows2016fs-release"), f.Name)

		if err != nil {
			return "", err
		}

		if fileMatch && f.Mode().IsRegular() {
			err = ex.extract(f, destDir)

			if err != nil {
				return "", err
			}
		}
	}

	return destDir, err
}

func (ex Extractor) extract(zipFile *zip.File, tempDir string) error {
	var err error
	var reader io.Reader

	reader, err = zipFile.Open()

	if err != nil {
		log.Fatal(err)
	}

	fullFilePath := path.Join(tempDir, zipFile.Name)

	err = ex.container.MkdirAll(path.Dir(fullFilePath), os.ModePerm)

	if err != nil {
		return err
	}

	fd, err := ex.container.OpenFile(fullFilePath, os.O_CREATE|os.O_WRONLY, zipFile.FileHeader.Mode())
	if err != nil {
		return err
	}

	_, err = ex.container.Copy(fd, reader)

	if err != nil {
		return err
	}

	return nil
}
