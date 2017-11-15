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
	Describe("Run application", func() {
		var (
			fakeExtractor      *fakes.Extractor
			fakeReleaseCreator *fakes.ReleaseCreator
			fakeInjector       *fakes.Injector
			fakeZipper         *fakes.Zipper

			app injector.Application
		)

		BeforeEach(func() {
			fakeExtractor = new(fakes.Extractor)
			fakeReleaseCreator = new(fakes.ReleaseCreator)
			fakeInjector = new(fakes.Injector)
			fakeZipper = new(fakes.Zipper)

			fakeExtractor.ExtractWindowsFSReleaseReturns("fixtures", nil)

			app = injector.NewApplication(fakeExtractor, fakeReleaseCreator, fakeInjector, fakeZipper)
		})

		It("Extracts WindowsFSRelease", func() {
			err := app.Run("/path/to/input/tile", "/path/to/working/dir")
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeExtractor.ExtractWindowsFSReleaseCallCount()).To(Equal(1))
			inputTile, workingDir := fakeExtractor.ExtractWindowsFSReleaseArgsForCall(0)
			Expect(inputTile).To(Equal("/path/to/input/tile"))
			Expect(workingDir).To(Equal("/path/to/working/dir"))
		})

		It("Creates the release", func() {
			err := app.Run("/path/to/input/tile", "/path/to/working/dir")
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeReleaseCreator.CreateReleaseCallCount()).To(Equal(1))
			imageName, releaseDir, tarballPath, imageTagPath, versionDataPath, winfsBlobsDir := fakeReleaseCreator.CreateReleaseArgsForCall(0)
			Expect(imageName).To(Equal("cloudfoundry/windows2016fs"))
			Expect(releaseDir).To(Equal("fixtures/embed/windows2016fs-release"))
			Expect(tarballPath).To(Equal("/path/to/working/dir/windows2016fs-9.3.6.tgz"))
			Expect(imageTagPath).To(Equal("fixtures/embed/windows2016fs-release/src/code.cloudfoundry.org/windows2016fs/IMAGE_TAG"))
			Expect(versionDataPath).To(Equal("fixtures/embed/windows2016fs-release/VERSION"))
			Expect(winfsBlobsDir).To(Equal("fixtures/embed/windows2016fs-release/blobs/windows2016fs"))
		})

		It("extracts the tile", func() {
			err := app.Run("/path/to/input/tile", "/path/to/working/dir")
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeExtractor.ExtractWindowsFSReleaseCallCount()).To(Equal(1))
			Expect(fakeReleaseCreator.CreateReleaseCallCount()).To(Equal(1))

			Expect(fakeExtractor.ExtractTileCallCount()).To(Equal(1))
			inputTile, outputDir := fakeExtractor.ExtractTileArgsForCall(0)
			Expect(inputTile).To(Equal("/path/to/input/tile"))
			Expect(outputDir).To(Equal(filepath.Join("/path/to/working/dir", "extracted-tile")))
		})

		It("injects the build windows release into the extracted tile", func() {
			err := app.Run("/path/to/input/tile", "/path/to/working/dir")
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeExtractor.ExtractWindowsFSReleaseCallCount()).To(Equal(1))
			Expect(fakeReleaseCreator.CreateReleaseCallCount()).To(Equal(1))
			Expect(fakeExtractor.ExtractTileCallCount()).To(Equal(1))

			Expect(fakeInjector.AddReleaseToTileCallCount()).To(Equal(1))
			releasePath, releaseName, releaseVersion, tileDir := fakeInjector.AddReleaseToTileArgsForCall(0)
			Expect(releasePath).To(Equal("/path/to/working/dir/windows2016fs-9.3.6.tgz"))
			Expect(releaseName).To(Equal("windows2016fs"))
			Expect(releaseVersion).To(Equal("9.3.6"))
			Expect(tileDir).To(Equal(filepath.Join("/path/to/working/dir", "extracted-tile")))
		})

		It("zips up the injected tile dir", func() {
			err := app.Run("/path/to/input/tile", "/path/to/working/dir")
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeExtractor.ExtractWindowsFSReleaseCallCount()).To(Equal(1))
			Expect(fakeReleaseCreator.CreateReleaseCallCount()).To(Equal(1))
			Expect(fakeExtractor.ExtractTileCallCount()).To(Equal(1))
			Expect(fakeInjector.AddReleaseToTileCallCount()).To(Equal(1))

			Expect(fakeZipper.ZipCallCount()).To(Equal(1))
			zipDir, zipFile := fakeZipper.ZipArgsForCall(0)
			Expect(zipDir).To(Equal(filepath.Join("/path/to/working/dir", "extracted-tile")))
			Expect(zipFile).To(Equal("CHANGE_ME.zip"))
		})

		Context("Failure cases", func() {
			Context("when the extractor fails to extract the windows fs release", func() {
				It("returns the error", func() {
					fakeExtractor.ExtractWindowsFSReleaseReturns("", errors.New("some-error"))

					err := app.Run("/path/to/input/tile", "/path/to/working/dir")
					Expect(err).To(MatchError("some-error"))
				})
			})

			Context("when the extractor fails to extract the tile", func() {
				It("returns the error", func() {
					fakeExtractor.ExtractTileReturns(errors.New("some-error"))
					err := app.Run("/path/to/input/tile", "/path/to/working/dir")
					Expect(err).To(MatchError("some-error"))
				})
			})

			Context("when the injector fails to copy the release into the tile", func() {
				It("returns the error", func() {
					fakeInjector.AddReleaseToTileReturns(errors.New("some-error"))
					err := app.Run("/path/to/input/tile", "/path/to/working/dir")
					Expect(err).To(MatchError("some-error"))
				})
			})

			Context("when image tag cannot be read", func() {
				It("returns the error", func() {
					fakeExtractor.ExtractWindowsFSReleaseReturns("invalid-fixtures", nil)

					err := app.Run("/path/to/input/tile", "/path/to/working/dir")
					Expect(err.Error()).To(ContainSubstring("no such file or directory"))
				})
			})

			Context("when the release creator fails", func() {
				It("returns the error", func() {
					fakeReleaseCreator.CreateReleaseReturns(errors.New("some-error"))

					err := app.Run("/path/to/input/tile", "/path/to/working/dir")
					Expect(err).To(MatchError("some-error"))
				})
			})

			Context("when zipping the injected tile dir fails", func() {
				It("returns the error", func() {
					fakeZipper.ZipReturns(errors.New("some-error"))

					err := app.Run("/path/to/input/tile", "/path/to/working/dir")
					Expect(err).To(MatchError("some-error"))
				})
			})
		})
	})
})
