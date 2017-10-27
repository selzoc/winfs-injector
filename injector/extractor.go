package injector

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

type extractor struct {
	openReader func(string) (*zip.ReadCloser, error)
	writeFile  func(filename string, data []byte, perm os.FileMode) error
	tempDir    func(dir string, prefix string) (name string, err error)
	mkdirAll   func(path string, perm os.FileMode) error
}

type Extractor interface {
	ExtractWindowsFSRelease(inputTile string) error
}

func NewExtractor(container ExtractContainer) Extractor {
	return &extractor{
		openReader: container.OpenReader,
		writeFile:  container.WriteFile,
		tempDir:    container.TempDir,
		mkdirAll:   container.MkdirAll,
	}
}

func (ex *extractor) extract(zipFile *zip.File, tempDir string) error {
	var bytes []byte
	var err error
	var reader io.Reader

	reader, err = zipFile.Open()

	if err != nil {
		panic(err)
	}

	bytes, err = ioutil.ReadAll(reader)

	if err != nil {
		panic(err)
	}

	fullFilePath := path.Join(tempDir, zipFile.Name)

	err = ex.mkdirAll(path.Dir(fullFilePath), os.ModePerm)

	if err != nil {
		panic(err)
	}

	err = ex.writeFile(fullFilePath, bytes, 0644)
	if err != nil {
		panic(err)
	}

	return nil
}

func (ex *extractor) ExtractWindowsFSRelease(inputTile string) error {
	r, err := ex.openReader(inputTile)

	if err != nil {
		return err
	}

	destDir, err := ex.tempDir("/tmp/", "windows2016fs")
	if err != nil {
		panic(err)
	}

	fmt.Printf(destDir)

	for _, f := range r.File {
		// impossible to fail since pattern is hard-coded
		fileMatch, _ := filepath.Match("*/embed/windows2016fs-release/**/*", f.Name)
		if fileMatch && f.Mode().IsRegular() {
			err = ex.extract(f, destDir)

			if err != nil {
				return err
			}
		}
	}

	return nil
}
