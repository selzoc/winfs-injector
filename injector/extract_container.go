package injector

import (
	"archive/zip"
	"io"
	"io/ioutil"
	"os"
	"regexp"
)

type ExtractContainer struct{}

func NewExtractContainer() ExtractContainer {
	return ExtractContainer{}
}

func (ec ExtractContainer) OpenReader(name string) (*zip.ReadCloser, error) {
	return zip.OpenReader(name)
}

func (ec ExtractContainer) TempDir(name string, prefix string) (string, error) {
	return ioutil.TempDir(name, prefix)
}

func (ec ExtractContainer) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (ec ExtractContainer) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(name, flag, perm)
}

func (ec ExtractContainer) Copy(dst io.Writer, src io.Reader) (int64, error) {
	return io.Copy(dst, src)
}

func (ec ExtractContainer) Match(pattern, name string) (bool, error) {
	return regexp.MatchString(pattern, name)
}
