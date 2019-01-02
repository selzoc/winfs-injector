package tile_test

import (
	"archive/zip"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/winfs-injector/tile"
)

var _ = Describe("Zipper", func() {
	Describe("Zip", func() {
		var (
			zipper  tile.Zipper
			srcDir  string
			zipFile *os.File
		)

		BeforeEach(func() {
			zipper = tile.NewZipper()

			var err error
			srcDir, err = ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())

			zipFile, err = ioutil.TempFile("", "")
			Expect(err).NotTo(HaveOccurred())

			err = ioutil.WriteFile(filepath.Join(srcDir, "top-level-file"), []byte("foo"), os.FileMode(0644))
			Expect(err).NotTo(HaveOccurred())

			err = os.Mkdir(filepath.Join(srcDir, "second-level-dir"), os.FileMode(0755))
			Expect(err).NotTo(HaveOccurred())

			err = ioutil.WriteFile(filepath.Join(srcDir, "second-level-dir", "second-level-file"), []byte("bar"), os.FileMode(0644))
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			err := os.RemoveAll(srcDir)
			Expect(err).NotTo(HaveOccurred())

			err = zipFile.Close()
			Expect(err).NotTo(HaveOccurred())

			err = os.RemoveAll(zipFile.Name())
			Expect(err).NotTo(HaveOccurred())
		})

		It("zips the specified directory and creates a zip at the specified path", func() {
			err := zipper.Zip(srcDir, zipFile.Name())
			Expect(err).NotTo(HaveOccurred())

			actualZip, err := zip.OpenReader(zipFile.Name())
			Expect(err).NotTo(HaveOccurred())

			Expect(actualZip.File).To(HaveLen(3))

			fileAssertions := map[string]string{
				"top-level-file":   "foo",
				"second-level-dir": "",
				filepath.Join("second-level-dir", "second-level-file"): "bar",
			}

			for _, f := range actualZip.File {
				openedFile, err := f.Open()
				Expect(err).NotTo(HaveOccurred())

				defer openedFile.Close()

				fileContents, err := ioutil.ReadAll(openedFile)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(fileContents)).To(Equal(fileAssertions[f.Name]))
			}
		})

		Context("failure cases", func() {
			Context("when an intermediate dir in the destination path does not exist", func() {
				It("returns an error", func() {
					err := zipper.Zip(srcDir, "/path/to/non-existing/dir")
					Expect(err).To(MatchError(ContainSubstring("/path/to/non-existing/dir")))
				})
			})

			Context("when the source dir does not exist", func() {
				It("returns an error", func() {
					err := zipper.Zip("/path/to/non-existing/dir", zipFile.Name())
					Expect(err).To(MatchError(ContainSubstring("/path/to/non-existing/dir")))
				})
			})
		})
	})

	Describe("Unzip", func() {
		var (
			inputTile string
			destDir   string
			zipper    tile.Zipper
		)

		BeforeEach(func() {
			inputTile = filepath.Join("fixtures", "test.zip")

			var err error
			destDir, err = ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			err := os.RemoveAll(destDir)
			Expect(err).NotTo(HaveOccurred())
		})

		It("unzips the specified zip to a specified dir", func() {
			err := zipper.Unzip(inputTile, destDir)
			Expect(err).NotTo(HaveOccurred())

			fileList := make(map[string]os.FileInfo)
			err = filepath.Walk(destDir, func(path string, fileInfo os.FileInfo, err error) error {
				fileList[filepath.Base(path)] = fileInfo
				return nil
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(fileList).To(HaveLen(4))

			Expect(fileList).To(HaveKey("top-level-dir"))
			Expect(fileList["top-level-dir"].Mode().IsDir()).To(BeTrue())
			Expect(fileList["top-level-dir"].Mode().Perm()).To(Equal(os.FileMode(0700)))

			Expect(fileList).To(HaveKey("nested-file"))
			Expect(fileList["nested-file"].Mode().Perm()).To(Equal(os.FileMode(0644)))

			Expect(fileList).To(HaveKey("top-level-file"))
			Expect(fileList["top-level-file"].Mode().Perm()).To(Equal(os.FileMode(0664)))
		})

		Context("failure cases", func() {
			Context("when the zip open fails", func() {
				It("returns an error", func() {
					err := zipper.Unzip("/path/to/non-existing/dir", destDir)
					Expect(err).To(MatchError(ContainSubstring("/path/to/non-existing/dir")))
				})
			})
		})
	})
})
