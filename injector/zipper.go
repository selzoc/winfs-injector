package injector

import "archive/zip"

type Zipper struct{}

func NewZipper() Zipper {
	return Zipper{}
}

func (z *Zipper) OpenReader(name string) (*zip.ReadCloser, error) {
	return zip.OpenReader(name)
}
