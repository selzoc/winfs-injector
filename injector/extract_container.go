package injector

import (
	"archive/zip"
	"io/ioutil"
	"os"
)

type ExtractContainer interface {
	OpenReader(string) (*zip.ReadCloser, error)
	WriteFile(filename string, data []byte, perm os.FileMode) error
	TempDir(string, string) (string, error)
	MkdirAll(path string, perm os.FileMode) error
}

type ProductionExtractContainer struct{}

func (p *ProductionExtractContainer) OpenReader(name string) (*zip.ReadCloser, error) {
	return zip.OpenReader(name)
}

func (p *ProductionExtractContainer) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return ioutil.WriteFile(filename, data, perm)
}

func (p *ProductionExtractContainer) TempDir(name string, prefix string) (string, error) {
	return ioutil.TempDir(name, prefix)
}

func (p *ProductionExtractContainer) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}
