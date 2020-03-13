package winfsinjector_test

import (
	"errors"
	"os"
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

			inputTile  string
			outputTile string
			registry   string
			workingDir string

			app winfsinjector.Application
		)

		BeforeEach(func() {
			fakeReleaseCreator = new(fakes.ReleaseCreator)
			fakeInjector = new(fakes.Injector)
			fakeZipper = new(fakes.Zipper)

			inputTile = "/path/to/input/tile"
			outputTile = "/path/to/output/tile"
			registry = "/path/to/docker/registry"
			workingDir = "/path/to/working/dir"

			winfsinjector.SetReadFile(func(path string) ([]byte, error) {
				switch filepath.Base(path) {
				case "VERSION":
					// reading VERSION file
					return []byte("9.3.6"), nil
				case "blobs.yml":
					// reading config/blobs.yml
					return []byte(`---
windows2019fs/windows2016fs-2019.0.43.tgz:
  size: 3333333333
  sha: abcdefg1234
`), nil
				case "final.yml":
					// reading config/final.yml
					return []byte(`name: windows2019fs`), nil
				default:
					return nil, errors.New("readFile called for unexpected input: " + path)
				}
			})

			fakeEmbeddedDirectory := new(fakes.FileInfo)
			fakeEmbeddedDirectory.IsDirReturns(true)
			fakeEmbeddedDirectory.NameReturns("windowsfs-release")

			winfsinjector.SetReadDir(func(string) ([]os.FileInfo, error) {
				return []os.FileInfo{
					fakeEmbeddedDirectory,
				}, nil
			})

			app = winfsinjector.NewApplication(fakeReleaseCreator, fakeInjector, fakeZipper)
		})

		AfterEach(func() {
			winfsinjector.ResetReadFile()
			winfsinjector.ResetRemoveAll()
			winfsinjector.ResetReadDir()
		})

		It("unzips the tile", func() {
			err := app.Run(inputTile, outputTile, registry, workingDir)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeZipper.UnzipCallCount()).To(Equal(1))

			inputTile, extractedTileDir := fakeZipper.UnzipArgsForCall(0)
			Expect(inputTile).To(Equal(filepath.Join("/", "path", "to", "input", "tile")))
			Expect(extractedTileDir).To(Equal(filepath.Join("/", "path", "to", "working", "dir", "extracted-tile")))
		})

		It("creates the release", func() {
			err := app.Run(inputTile, outputTile, registry, workingDir)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeReleaseCreator.CreateReleaseCallCount()).To(Equal(1))

			releaseName, imageName, releaseDir, tarballPath, imageTag, registry, version := fakeReleaseCreator.CreateReleaseArgsForCall(0)
			Expect(releaseName).To(Equal("windows2019fs"))
			Expect(imageName).To(Equal("cloudfoundry/windows2016fs"))
			Expect(releaseDir).To(Equal("/path/to/working/dir/extracted-tile/embed/windowsfs-release"))
			Expect(tarballPath).To(Equal("/path/to/working/dir/extracted-tile/releases/windows2019fs-9.3.6.tgz"))
			Expect(imageTag).To(Equal("2019.0.43"))
			Expect(registry).To(Equal("/path/to/docker/registry"))
			Expect(version).To(Equal("9.3.6"))
		})

		It("injects the build windows release into the extracted tile", func() {
			err := app.Run(inputTile, outputTile, registry, workingDir)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeReleaseCreator.CreateReleaseCallCount()).To(Equal(1))
			Expect(fakeZipper.UnzipCallCount()).To(Equal(1))

			Expect(fakeInjector.AddReleaseToMetadataCallCount()).To(Equal(1))
			releasePath, releaseName, releaseVersion, tileDir := fakeInjector.AddReleaseToMetadataArgsForCall(0)
			Expect(releasePath).To(Equal("/path/to/working/dir/extracted-tile/releases/windows2019fs-9.3.6.tgz"))
			Expect(releaseName).To(Equal("windows2019fs"))
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

			err := app.Run(inputTile, outputTile, registry, workingDir)
			Expect(err).NotTo(HaveOccurred())

			Expect(removeAllCallCount).To(Equal(1))
			Expect(removeAllPath).To(Equal(filepath.Join("/", "path", "to", "working", "dir", "extracted-tile", "embed", "windowsfs-release")))
		})

		It("zips up the injected tile dir", func() {
			err := app.Run(inputTile, outputTile, registry, workingDir)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeReleaseCreator.CreateReleaseCallCount()).To(Equal(1))
			Expect(fakeZipper.UnzipCallCount()).To(Equal(1))
			Expect(fakeInjector.AddReleaseToMetadataCallCount()).To(Equal(1))

			Expect(fakeZipper.ZipCallCount()).To(Equal(1))
			zipDir, zipFile := fakeZipper.ZipArgsForCall(0)
			Expect(zipDir).To(Equal(filepath.Join("/path/to/working/dir", "extracted-tile")))
			Expect(zipFile).To(Equal("/path/to/output/tile"))
		})

		Context("when the image tag of release dir is malformed", func() {
			BeforeEach(func() {
				winfsinjector.SetReadFile(func(path string) ([]byte, error) {
					switch filepath.Base(path) {
					case "VERSION":
						// reading VERSION file
						return []byte(""), nil
					case "blobs.yml":
						// reading config/blobs.yml
						return []byte(`---
windows2019fs/windows2016fs-MISSING-IMAGE-TAG.tgz:
  size: 3333333333
  sha: abcdefg1234
`), nil
					case "final.yml":
						// reading config/final.yml
						return []byte(``), nil
					default:
						return nil, errors.New("readFile called for unexpected input: " + path)
					}
				})
			})

			It("returns the error", func() {
				err := app.Run(inputTile, outputTile, registry, workingDir)
				Expect(err).To(MatchError(ContainSubstring("unable to parse tag from embedded rootfs:")))
			})
		})

		Context("when the zipper fails to unzip the tile", func() {
			BeforeEach(func() {
				fakeZipper.UnzipReturns(errors.New("some-error"))
			})
			It("returns the error", func() {
				err := app.Run(inputTile, outputTile, registry, workingDir)
				Expect(err).To(MatchError("some-error"))
			})
		})

		Context("when the injector fails to copy the release into the tile", func() {
			BeforeEach(func() {
				fakeInjector.AddReleaseToMetadataReturns(errors.New("some-error"))
			})

			It("returns the error", func() {
				err := app.Run(inputTile, outputTile, registry, workingDir)
				Expect(err).To(MatchError("some-error"))
			})
		})

		Context("when the release creator fails", func() {
			BeforeEach(func() {
				fakeReleaseCreator.CreateReleaseReturns(errors.New("some-error"))
			})

			It("returns the error", func() {
				err := app.Run(inputTile, outputTile, registry, workingDir)
				Expect(err).To(MatchError("some-error"))
			})
		})

		Context("when removing the windows2016fs-release dir from the embed directory fails", func() {
			BeforeEach(func() {
				winfsinjector.SetRemoveAll(func(path string) error {
					return errors.New("remove all failed")
				})
			})

			It("returns an error", func() {
				err := app.Run(inputTile, outputTile, registry, workingDir)
				Expect(err).To(MatchError("remove all failed"))
			})
		})

		Context("when zipping the injected tile dir fails", func() {
			BeforeEach(func() {
				fakeZipper.ZipReturns(errors.New("some-error"))
			})

			It("returns the error", func() {
				err := app.Run(inputTile, outputTile, registry, workingDir)
				Expect(err).To(MatchError("some-error"))
			})
		})

		Context("when input tile is not provided", func() {
			It("returns an error", func() {
				err := app.Run("", outputTile, registry, workingDir)
				Expect(err).To(MatchError("--input-tile is required"))
			})
		})

		Context("when output tile is not provided", func() {
			It("returns an error", func() {
				err := app.Run(inputTile, "", registry, workingDir)
				Expect(err).To(MatchError("--output-tile is required"))
			})
		})
	})
})
