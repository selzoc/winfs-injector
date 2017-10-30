package injector

import (
	"archive/zip"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
)

type extractor struct {
	openReader func(string) (*zip.ReadCloser, error)
	tempDir    func(dir string, prefix string) (name string, err error)
	mkdirAll   func(path string, perm os.FileMode) error
	openFile   func(name string, flag int, perm os.FileMode) (*os.File, error)
	copy       func(dst io.Writer, src io.Reader) (written int64, err error)
	match      func(pattern, name string) (matched bool, err error)
}

type Extractor interface {
	ExtractWindowsFSRelease(inputTile string, outputDir string) (string, error)
}

func NewExtractor(container ExtractContainer) Extractor {
	return &extractor{
		openReader: container.OpenReader,
		tempDir:    container.TempDir,
		mkdirAll:   container.MkdirAll,
		openFile:   container.OpenFile,
		copy:       container.Copy,
		match:      container.Match,
	}
}

func (ex *extractor) ExtractWindowsFSRelease(inputTile, outputDir string) (string, error) {
	r, err := ex.openReader(inputTile)

	if err != nil {
		return "", err
	}

	destDir, err := ex.tempDir(outputDir, "windows2016fs")
	if err != nil {
		return "", err
	}

	for _, f := range r.File {
		fileMatch, err := ex.match(filepath.Join("*", "embed", "windows2016fs-release", "**", "*"), f.Name)

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

func (ex *extractor) extract(zipFile *zip.File, tempDir string) error {
	var err error
	var reader io.Reader

	reader, err = zipFile.Open()

	if err != nil {
		log.Fatal(err)
	}

	fullFilePath := path.Join(tempDir, zipFile.Name)

	err = ex.mkdirAll(path.Dir(fullFilePath), os.ModePerm)

	if err != nil {
		return err
	}

	fd, err := ex.openFile(fullFilePath, os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	_, err = ex.copy(fd, reader)

	if err != nil {
		return err
	}

	return nil
}
