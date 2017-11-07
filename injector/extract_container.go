package injector

import (
	"archive/zip"
	"io"
	"io/ioutil"
	"os"
	"regexp"
)

//go:generate counterfeiter -o ./fakes/extract_container.go --fake-name ExtractContainer . ExtractContainer
type ExtractContainer interface {
	OpenReader(string) (*zip.ReadCloser, error)
	TempDir(string, string) (string, error)
	MkdirAll(path string, perm os.FileMode) error
	OpenFile(name string, flag int, perm os.FileMode) (*os.File, error)
	Copy(dst io.Writer, src io.Reader) (written int64, err error)
	Match(pattern, name string) (matched bool, err error)
}

type ProductionExtractContainer struct{}

func (p *ProductionExtractContainer) OpenReader(name string) (*zip.ReadCloser, error) {
	return zip.OpenReader(name)
}

func (p *ProductionExtractContainer) TempDir(name string, prefix string) (string, error) {
	return ioutil.TempDir(name, prefix)
}

func (p *ProductionExtractContainer) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (p *ProductionExtractContainer) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(name, flag, perm)
}

func (p *ProductionExtractContainer) Copy(dst io.Writer, src io.Reader) (written int64, err error) {
	return io.Copy(dst, src)
}

func (p *ProductionExtractContainer) Match(pattern, name string) (matched bool, err error) {
	return regexp.MatchString(pattern, name)
}
