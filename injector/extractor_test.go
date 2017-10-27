package injector_test

import (
	"archive/zip"
	"bytes"
	"errors"
	"os"

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
		// var inputTile string

		var fakeExtractContainer *injectorfakes.FakeExtractContainer
		var extractor injector.Extractor

		BeforeEach(func() {
			fakeExtractContainer = new(injectorfakes.FakeExtractContainer)
			extractor = injector.NewExtractor(fakeExtractContainer)

			var buffer = new(bytes.Buffer)
			var writer = zip.NewWriter(buffer)

			createFile(writer, "some-tile/embed/windows2016fs-release/foo/bar.gif", "hello")
			createFile(writer, "some-tile/embed/windows2016fs-release/baz/qux.gif", "hello")

			writer.Flush()
			writer.Close()

			var readerAt = bytes.NewReader(buffer.Bytes())

			var reader, _ = zip.NewReader(readerAt, int64(buffer.Len()))

			var openReaderFakeReturns = &zip.ReadCloser{
				Reader: *reader,
			}

			fakeExtractContainer.OpenReaderReturns(openReaderFakeReturns, nil)
			fakeExtractContainer.TempDirReturns("/tmp", nil)
		})

		It("Unzips and extracts windows2016fs", func() {
			extractor.ExtractWindowsFSRelease("windows2016fs-release")

			Expect(fakeExtractContainer.OpenReaderArgsForCall(0)).To(Equal("windows2016fs-release"))
			Expect(fakeExtractContainer.OpenReaderCallCount()).To(Equal(1))

			var path, mode = fakeExtractContainer.MkdirAllArgsForCall(0)
			Expect(path).To(Equal("/tmp/some-tile/embed/windows2016fs-release/foo"))
			filePerm := int(0777)
			Expect(mode).To(Equal(os.FileMode(filePerm)))

			var name, data, _ = fakeExtractContainer.WriteFileArgsForCall(0)
			Expect(name).To(Equal("/tmp/some-tile/embed/windows2016fs-release/foo/bar.gif"))
			Expect(string(data)).To(Equal("hello"))

			name, data, _ = fakeExtractContainer.WriteFileArgsForCall(1)
			Expect(name).To(Equal("/tmp/some-tile/embed/windows2016fs-release/baz/qux.gif"))
			Expect(string(data)).To(Equal("hello"))

			Expect(fakeExtractContainer.TempDirCallCount()).To(Equal(1))
		})

		Context("failure cases", func() {
			Describe("when the open reader call fails", func() {
				It("should return the error", func() {
					fakeExtractContainer.OpenReaderReturns(nil, errors.New("some failure"))
					err := extractor.ExtractWindowsFSRelease("windows2016fs-release")

					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError("some failure"))
				})
			})
		})
	})
})
