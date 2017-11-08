package injector_test

import (
	"errors"

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
			app                injector.Application
		)

		BeforeEach(func() {
			fakeExtractor = new(fakes.Extractor)
			fakeReleaseCreator = new(fakes.ReleaseCreator)
			app = injector.NewApplication(fakeExtractor, fakeReleaseCreator)

			fakeExtractor.ExtractWindowsFSReleaseReturns("fixtures", nil)
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
			Expect(tarballPath).To(Equal("/path/to/working/dir/windows2016fs-some-version.tgz"))
			Expect(imageTagPath).To(Equal("fixtures/embed/windows2016fs-release/src/code.cloudfoundry.org/windows2016fs/IMAGE_TAG"))
			Expect(versionDataPath).To(Equal("fixtures/embed/windows2016fs-release/VERSION"))
			Expect(winfsBlobsDir).To(Equal("fixtures/embed/windows2016fs-release/blobs/windows2016fs"))
		})

		Context("Failure cases", func() {
			Context("When the extractor fails", func() {
				It("returns the error", func() {
					fakeExtractor.ExtractWindowsFSReleaseReturns("", errors.New("some-error"))

					err := app.Run("/path/to/input/tile", "/path/to/working/dir")
					Expect(err).To(MatchError("some-error"))
				})
			})

			Context("When image tag cannot be read", func() {
				It("returns the error", func() {
					fakeExtractor.ExtractWindowsFSReleaseReturns("invalid-fixtures", nil)

					err := app.Run("/path/to/input/tile", "/path/to/working/dir")
					Expect(err.Error()).To(ContainSubstring("no such file or directory"))
				})
			})

			Context("When the release creator fails", func() {
				It("returns the error", func() {
					fakeReleaseCreator.CreateReleaseReturns(errors.New("some-error"))

					err := app.Run("/path/to/input/tile", "/path/to/working/dir")
					Expect(err).To(MatchError("some-error"))
				})
			})

		})
	})
})
