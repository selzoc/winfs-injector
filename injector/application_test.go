package injector_test

import (
	"errors"
	"path/filepath"

	"github.com/pivotal-cf/winfs-injector/injector"
	"github.com/pivotal-cf/winfs-injector/injector/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("application", func() {
	Describe("Run", func() {
		var (
			fakeReleaseCreator *fakes.ReleaseCreator
			fakeInjector       *fakes.Injector
			fakeZipper         *fakes.Zipper

			app injector.Application
		)

		BeforeEach(func() {
			fakeReleaseCreator = new(fakes.ReleaseCreator)
			fakeInjector = new(fakes.Injector)
			fakeZipper = new(fakes.Zipper)

			injector.SetReadFile(func(string) ([]byte, error) {
				return []byte("9.3.6"), nil
			})

			app = injector.NewApplication(fakeReleaseCreator, fakeInjector, fakeZipper)
		})

		AfterEach(func() {
			injector.ResetReadFile()
			injector.ResetRemoveAll()
		})

		It("unzips the tile", func() {
			err := app.Run("/path/to/input/tile", "/path/to/output/tile", "/path/to/working/dir")
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeZipper.UnzipCallCount()).To(Equal(1))

			inputTile, extractedTileDir := fakeZipper.UnzipArgsForCall(0)
			Expect(inputTile).To(Equal(filepath.Join("/", "path", "to", "input", "tile")))
			Expect(extractedTileDir).To(Equal(filepath.Join("/", "path", "to", "working", "dir", "extracted-tile")))
		})

		It("creates the release", func() {
			err := app.Run("/path/to/input/tile", "/path/to/output/tile", "/path/to/working/dir")
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeReleaseCreator.CreateReleaseCallCount()).To(Equal(1))
			imageName, releaseDir, tarballPath, imageTagPath, versionDataPath, winfsBlobsDir := fakeReleaseCreator.CreateReleaseArgsForCall(0)
			Expect(imageName).To(Equal("cloudfoundry/windows2016fs"))
			Expect(releaseDir).To(Equal("/path/to/working/dir/extracted-tile/embed/windows2016fs-release"))
			Expect(tarballPath).To(Equal("/path/to/working/dir/extracted-tile/releases/windows2016fs-9.3.6.tgz"))
			Expect(imageTagPath).To(Equal("/path/to/working/dir/extracted-tile/embed/windows2016fs-release/src/code.cloudfoundry.org/windows2016fs/IMAGE_TAG"))
			Expect(versionDataPath).To(Equal("/path/to/working/dir/extracted-tile/embed/windows2016fs-release/VERSION"))
			Expect(winfsBlobsDir).To(Equal("/path/to/working/dir/extracted-tile/embed/windows2016fs-release/blobs/windows2016fs"))
		})

		It("injects the build windows release into the extracted tile", func() {
			err := app.Run("/path/to/input/tile", "/path/to/output/tile", "/path/to/working/dir")
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeReleaseCreator.CreateReleaseCallCount()).To(Equal(1))
			Expect(fakeZipper.UnzipCallCount()).To(Equal(1))

			Expect(fakeInjector.AddReleaseToMetadataCallCount()).To(Equal(1))
			releasePath, releaseName, releaseVersion, tileDir := fakeInjector.AddReleaseToMetadataArgsForCall(0)
			Expect(releasePath).To(Equal("/path/to/working/dir/extracted-tile/releases/windows2016fs-9.3.6.tgz"))
			Expect(releaseName).To(Equal("windows2016fs"))
			Expect(releaseVersion).To(Equal("9.3.6"))
			Expect(tileDir).To(Equal(filepath.Join("/path/to/working/dir", "extracted-tile")))
		})

		It("removes the windows2016fs-release from the embed directory", func() {
			var (
				removeAllCallCount int
				removeAllPath      string
			)

			injector.SetRemoveAll(func(path string) error {
				removeAllCallCount++
				removeAllPath = path
				return nil
			})

			err := app.Run("/path/to/input/tile", "/path/to/output/tile", "/path/to/working/dir")
			Expect(err).NotTo(HaveOccurred())

			Expect(removeAllCallCount).To(Equal(1))
			Expect(removeAllPath).To(Equal(filepath.Join("/", "path", "to", "working", "dir", "extracted-tile", "embed", "windows2016fs-release")))
		})

		It("zips up the injected tile dir", func() {
			err := app.Run("/path/to/input/tile", "/path/to/output/tile", "/path/to/working/dir")
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeReleaseCreator.CreateReleaseCallCount()).To(Equal(1))
			Expect(fakeZipper.UnzipCallCount()).To(Equal(1))
			Expect(fakeInjector.AddReleaseToMetadataCallCount()).To(Equal(1))

			Expect(fakeZipper.ZipCallCount()).To(Equal(1))
			zipDir, zipFile := fakeZipper.ZipArgsForCall(0)
			Expect(zipDir).To(Equal(filepath.Join("/path/to/working/dir", "extracted-tile")))
			Expect(zipFile).To(Equal("/path/to/output/tile"))
		})

		Context("failure cases", func() {
			Context("when the zipper fails to unzip the tile", func() {
				It("returns the error", func() {
					fakeZipper.UnzipReturns(errors.New("some-error"))
					err := app.Run("/path/to/input/tile", "/path/to/output/tile", "/path/to/working/dir")
					Expect(err).To(MatchError("some-error"))
				})
			})

			Context("when the injector fails to copy the release into the tile", func() {
				It("returns the error", func() {
					fakeInjector.AddReleaseToMetadataReturns(errors.New("some-error"))
					err := app.Run("/path/to/input/tile", "/path/to/output/tile", "/path/to/working/dir")
					Expect(err).To(MatchError("some-error"))
				})
			})

			Context("when the release creator fails", func() {
				It("returns the error", func() {
					fakeReleaseCreator.CreateReleaseReturns(errors.New("some-error"))

					err := app.Run("/path/to/input/tile", "/path/to/output/tile", "/path/to/working/dir")
					Expect(err).To(MatchError("some-error"))
				})
			})

			Context("when removing the windows2016fs-release dir from the embed directory fails", func() {
				It("returns an error", func() {
					injector.SetRemoveAll(func(path string) error {
						return errors.New("remove all failed")
					})

					err := app.Run("/path/to/input/tile", "/path/to/output/tile", "/path/to/working/dir")
					Expect(err).To(MatchError("remove all failed"))
				})
			})

			Context("when zipping the injected tile dir fails", func() {
				It("returns the error", func() {
					fakeZipper.ZipReturns(errors.New("some-error"))

					err := app.Run("/path/to/input/tile", "/path/to/output/tile", "/path/to/working/dir")
					Expect(err).To(MatchError("some-error"))
				})
			})
		})
	})
})
