package injector_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/winfs-injector/injector"
)

var _ = Describe("TileInjector", func() {
	var (
		tileInjector injector.TileInjector

		baseTmpDir       string
		tileDir          string
		releasePath      string
		releaseName      string
		releaseVersion   string
		metadataPath     string
		expectedMetadata injector.Metadata
	)

	BeforeEach(func() {
		releaseName = "some-release"
		releaseVersion = "1.2.3"

		var err error
		baseTmpDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		releasePath = filepath.Join(baseTmpDir, "some-release.tgz")
		err = ioutil.WriteFile(releasePath, []byte("something"), 0644)
		Expect(err).NotTo(HaveOccurred())

		tileDir = filepath.Join(baseTmpDir, "some-tile")
		err = os.Mkdir(tileDir, 0755)
		Expect(err).NotTo(HaveOccurred())

		initialMetadataPath := filepath.Join("fixtures", "initial_metadata.yml")
		initialMetadataContents, err := ioutil.ReadFile(initialMetadataPath)
		Expect(err).NotTo(HaveOccurred())

		err = os.Mkdir(filepath.Join(tileDir, "metadata"), 0755)
		Expect(err).NotTo(HaveOccurred())

		metadataPath = filepath.Join(tileDir, "metadata", "some-product-metadata.yml")
		err = ioutil.WriteFile(metadataPath, initialMetadataContents, 0644)
		Expect(err).NotTo(HaveOccurred())

		expectedMetadataPath := filepath.Join("fixtures", "expected_metadata.yml")
		expectedMetadataContents, err := ioutil.ReadFile(expectedMetadataPath)
		Expect(err).NotTo(HaveOccurred())

		err = yaml.Unmarshal(expectedMetadataContents, &expectedMetadata)
		Expect(err).NotTo(HaveOccurred())

		tileInjector = injector.NewTileInjector()
	})

	AfterEach(func() {
		Expect(os.RemoveAll(baseTmpDir)).To(Succeed())
	})

	Describe("AddReleaseToMetadata", func() {
		It("adds the release to the tile metadata", func() {
			err := tileInjector.AddReleaseToMetadata(releasePath, releaseName, releaseVersion, tileDir)
			Expect(err).NotTo(HaveOccurred())

			rawMetadata, err := ioutil.ReadFile(metadataPath)
			Expect(err).NotTo(HaveOccurred())

			var actualMetadata injector.Metadata
			Expect(yaml.Unmarshal(rawMetadata, &actualMetadata)).To(Succeed())

			Expect(actualMetadata).To(Equal(expectedMetadata))
		})

		Context("failure cases", func() {
			It("returns an error when opening the metadata file fails", func() {
				Expect(os.RemoveAll(metadataPath)).To(Succeed())

				err := tileInjector.AddReleaseToMetadata(releasePath, releaseName, releaseVersion, tileDir)
				Expect(err).To(MatchError(ContainSubstring("expected to find a product metadata file")))
			})

			It("returns an error when reading the metadata file fails", func() {
				Expect(os.RemoveAll(metadataPath)).To(Succeed())
				Expect(os.MkdirAll(metadataPath, 0777)).To(Succeed())

				err := tileInjector.AddReleaseToMetadata(releasePath, releaseName, releaseVersion, tileDir)
				Expect(err).To(MatchError(ContainSubstring("is a directory")))
			})

			It("returns an error when metadata contains malformed yaml", func() {
				err := ioutil.WriteFile(metadataPath, []byte("%%%%"), 0644)
				Expect(err).NotTo(HaveOccurred())

				err = tileInjector.AddReleaseToMetadata(releasePath, releaseName, releaseVersion, tileDir)
				Expect(err).To(MatchError(ContainSubstring("yaml: ")))
			})

			It("returns an error when multiple yaml files exist in the metadata directory", func() {
				secondMetadataPath := filepath.Join(filepath.Dir(metadataPath), "second.yml")
				err := ioutil.WriteFile(secondMetadataPath, []byte("{}"), 0644)
				Expect(err).NotTo(HaveOccurred())

				err = tileInjector.AddReleaseToMetadata(releasePath, releaseName, releaseVersion, tileDir)
				Expect(err).To(MatchError(ContainSubstring("expected to find a single metadata file")))
			})
		})
	})
})
