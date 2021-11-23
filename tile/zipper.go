package tile

import (
	"fmt"
	"os"

	"github.com/jhoonb/archivex"
	"github.com/mholt/archiver"
)

type Zipper struct{}

func NewZipper() Zipper {
	return Zipper{}
}

func (z Zipper) Zip(zipDir, outputFile string) error {
	zipFile := fmt.Sprintf("%s.zip", outputFile)
	zf := archivex.ZipFile{}

	err := zf.Create(zipFile)
	if err != nil {
		return err
	}

	err = zf.AddAll(zipDir, false)
	if err != nil {
		return err
	}

	err = zf.Close()
	if err != nil {
		return err
	}

	err = os.Rename(zipFile, outputFile)
	if err != nil {
		return err
	}

	return nil
}

func (z Zipper) Unzip(zipFile, outputDir string) error {
	return archiver.DefaultZip.Unarchive(zipFile, outputDir)
}
