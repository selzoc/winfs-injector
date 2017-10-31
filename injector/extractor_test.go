package injector_test

import (
	"archive/zip"
	"bytes"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/winfs-injector/injector"
	"github.com/pivotal-cf/winfs-injector/injector/injectorfakes"
)

var createFile = func(writer *zip.Writer, name string, contents string) {
	var fileWriter, _ = writer.Create(name)
	fileWriter.Write([]byte(contents))
}

var _ = Describe("injector", func() {
	Describe("ExtractWindowsFSRelease", func() {
		var fakeExtractContainer *injectorfakes.FakeExtractContainer
		var extractor injector.Extractor
		var openFileFakeReturns *os.File

		BeforeEach(func() {
			fakeExtractContainer = new(injectorfakes.FakeExtractContainer)
			extractor = injector.NewExtractor(fakeExtractContainer)

			fakeExtractContainer.MatchStub = regexp.MatchString

			var buffer = new(bytes.Buffer)
			var writer = zip.NewWriter(buffer)
			var err error

			createFile(writer, filepath.Join("some-tile", "embed", "windows2016fs-release", "foo", "bar.gif"), "hello")
			createFile(writer, filepath.Join("some-tile", "embed", "windows2016fs-release", "baz", "qux.gif"), "hello")

			writer.Flush()
			writer.Close()

			var readerAt = bytes.NewReader(buffer.Bytes())

			var reader, _ = zip.NewReader(readerAt, int64(buffer.Len()))

			var openReaderFakeReturns = &zip.ReadCloser{
				Reader: *reader,
			}

			openFileFakeReturns, err = ioutil.TempFile(".", "tmp-")
			if err != nil {
				log.Fatal(err)
			}
			fakeExtractContainer.OpenFileReturns(openFileFakeReturns, nil)

			fakeExtractContainer.OpenReaderReturns(openReaderFakeReturns, nil)
			fakeExtractContainer.TempDirReturns("tmp", nil)
		})

		AfterEach(func() {
			var err error
			err = os.Remove(openFileFakeReturns.Name())
			if err != nil {
				log.Fatal(err)
			}
		})

		Context("unzipping and extracting", func() {
			var resultPath string
			BeforeEach(func() {
				resultPath, _ = extractor.ExtractWindowsFSRelease("windows2016fs-release", "tmp")
			})

			It("opens a zip reader", func() {
				Expect(fakeExtractContainer.OpenReaderArgsForCall(0)).To(Equal("windows2016fs-release"))
				Expect(fakeExtractContainer.OpenReaderCallCount()).To(Equal(1))
			})

			It("creates a temp dir", func() {
				Expect(fakeExtractContainer.TempDirCallCount()).To(Equal(1))
			})

			It("extracts all the matching files", func() {
				var path, mode = fakeExtractContainer.MkdirAllArgsForCall(0)
				Expect(path).To(Equal(filepath.Join("tmp", "some-tile", "embed", "windows2016fs-release", "foo")))
				filePerm := int(0777)
				Expect(mode).To(Equal(os.FileMode(filePerm)))

				var name, _, _ = fakeExtractContainer.OpenFileArgsForCall(0)
				Expect(name).To(Equal(filepath.Join("tmp", "some-tile", "embed", "windows2016fs-release", "foo", "bar.gif")))

				name, _, _ = fakeExtractContainer.OpenFileArgsForCall(1)
				Expect(name).To(Equal(filepath.Join("tmp", "some-tile", "embed", "windows2016fs-release", "baz", "qux.gif")))

				Expect(fakeExtractContainer.CopyCallCount()).To(Equal(2))
			})

			It("returns the directory the windows2016fs-release was extracted to", func() {
				Expect(resultPath).To(Equal("tmp"))
			})
		})

		Context("failure cases", func() {
			Describe("when the open reader call fails", func() {
				It("should return the error", func() {
					fakeExtractContainer.OpenReaderReturns(nil, errors.New("some failure"))
					_, err := extractor.ExtractWindowsFSRelease("windows2016fs-release", "tmp")

					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError("some failure"))
				})
			})

			Describe("when getting a tempDir fails", func() {
				It("should return the error", func() {
					fakeExtractContainer.TempDirReturns("", errors.New("failed to get temp dir"))
					_, err := extractor.ExtractWindowsFSRelease("windows2016fs-release", "tmp")

					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError("failed to get temp dir"))
				})
			})

			Describe("when filepath matching fails", func() {
				It("should return the error", func() {
					fakeExtractContainer.MatchReturns(false, errors.New("some failure"))
					_, err := extractor.ExtractWindowsFSRelease("windows2016fs-release", "tmp")

					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError("some failure"))
				})
			})

			Describe("when copying the file fails", func() {
				It("should return the error", func() {
					fakeExtractContainer.CopyReturns(0, errors.New("some failure"))
					_, err := extractor.ExtractWindowsFSRelease("windows2016fs-release", "tmp")

					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError("some failure"))
				})
			})

			Describe("when mkdirAll fails", func() {
				It("should return the error", func() {
					fakeExtractContainer.MkdirAllReturns(errors.New("some failure"))
					_, err := extractor.ExtractWindowsFSRelease("windows2016fs-release", "tmp")

					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError("some failure"))
				})
			})
		})
	})
})
