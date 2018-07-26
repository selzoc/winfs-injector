package winfsinjector_test

import (
	"errors"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/winfs-injector/winfsinjector"
	"github.com/pivotal-cf/winfs-injector/winfsinjector/fakes"
)

var _ = Describe("application", func() {
	Describe("Run", func() {
		var (
			fakeReleaseCreator *fakes.ReleaseCreator
			fakeInjector       *fakes.Injector
			fakeZipper         *fakes.Zipper

			app winfsinjector.Application
		)

		BeforeEach(func() {
			fakeReleaseCreator = new(fakes.ReleaseCreator)
			fakeInjector = new(fakes.Injector)
			fakeZipper = new(fakes.Zipper)

			winfsinjector.SetReadFile(func(string) ([]byte, error) {
				return []byte("9.3.6"), nil
			})

			app = winfsinjector.NewApplication(fakeReleaseCreator, fakeInjector, fakeZipper)
		})

		AfterEach(func() {
			winfsinjector.ResetReadFile()
			winfsinjector.ResetRemoveAll()
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
			imageName, releaseDir, tarballPath, imageTagPath, versionDataPath := fakeReleaseCreator.CreateReleaseArgsForCall(0)
			Expect(imageName).To(Equal("cloudfoundry/windows2016fs"))
			Expect(releaseDir).To(Equal("/path/to/working/dir/extracted-tile/embed/windows2016fs-release"))
			Expect(tarballPath).To(Equal("/path/to/working/dir/extracted-tile/releases/windows2016fs-9.3.6.tgz"))
			Expect(imageTagPath).To(Equal("/path/to/working/dir/extracted-tile/embed/windows2016fs-release/src/code.cloudfoundry.org/windows2016fs/1709/IMAGE_TAG"))
			Expect(versionDataPath).To(Equal("/path/to/working/dir/extracted-tile/embed/windows2016fs-release/VERSION"))
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

			winfsinjector.SetRemoveAll(func(path string) error {
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
					winfsinjector.SetRemoveAll(func(path string) error {
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

			Context("when input tile is not provided", func() {
				It("returns an error", func() {
					err := app.Run("", "/path/to/output/tile", "/path/to/working/dir")
					Expect(err).To(MatchError("--input-tile is required"))
				})
			})

			Context("when output tile is not provided", func() {
				It("returns an error", func() {
					err := app.Run("/path/to/input/tile", "", "/path/to/working/dir")
					Expect(err).To(MatchError("--output-tile is required"))
				})
			})
		})
	})
})
