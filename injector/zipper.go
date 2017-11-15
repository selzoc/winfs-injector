package injector

type Zipper struct{}

func NewZipper() Zipper {
	return Zipper{}
}

func (z Zipper) Zip(zipDir, zipFile string) error {
	panic("not implemented")
}
