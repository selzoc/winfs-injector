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

		baseTmpDir,
		tileDir,
		releasePath,
		releaseVersion,
		releaseName string

		expectedInitialMetatdata injector.Metadata
	)

	BeforeEach(func() {
		var err error
		baseTmpDir, err = ioutil.TempDir("", "injector-test")
		Expect(err).NotTo(HaveOccurred())

		tileDir, err = ioutil.TempDir(baseTmpDir, "")
		Expect(err).NotTo(HaveOccurred())

		rawExampleMetadata, err := ioutil.ReadFile(filepath.Join("fixtures", "example_metadata.yml"))
		Expect(err).NotTo(HaveOccurred())

		Expect(yaml.Unmarshal(rawExampleMetadata, &expectedInitialMetatdata)).To(Succeed())

		err = ioutil.WriteFile(filepath.Join(tileDir, "metadata.yml"), rawExampleMetadata, 0644)
		Expect(err).NotTo(HaveOccurred())

		releaseName = "some-release"
		releaseVersion = "9.3.6"

		releaseFile, err := ioutil.TempFile(baseTmpDir, "")
		Expect(err).NotTo(HaveOccurred())
		defer releaseFile.Close()

		_, err = releaseFile.Write([]byte("some-release-contents"))
		Expect(err).NotTo(HaveOccurred())

		releasePath = releaseFile.Name()

		tileInjector = injector.NewTileInjector()
	})

	AfterEach(func() {
		Expect(os.RemoveAll(baseTmpDir)).To(Succeed())
	})

	Describe("AddReleaseToTile", func() {
		It("moves the release into the tile's releases dir", func() {
			err := tileInjector.AddReleaseToTile(releasePath, releaseName, releaseVersion, tileDir)
			Expect(err).NotTo(HaveOccurred())

			fi, err := os.Stat(filepath.Join(tileDir, "releases"))
			Expect(err).NotTo(HaveOccurred())

			Expect(fi.IsDir()).To(BeTrue())
			Expect(fi.Mode() & os.ModePerm).To(Equal(os.FileMode(0755)))

			fi, err = os.Stat(filepath.Join(tileDir, "releases", filepath.Base(releasePath)))
			Expect(err).NotTo(HaveOccurred())
			Expect(fi.IsDir()).NotTo(BeTrue())
		})

		It("adds the release to the tile metadata", func() {
			err := tileInjector.AddReleaseToTile(releasePath, releaseName, releaseVersion, tileDir)
			Expect(err).NotTo(HaveOccurred())

			rawMetadata, err := ioutil.ReadFile(filepath.Join(tileDir, "metadata.yml"))
			Expect(err).NotTo(HaveOccurred())

			var actualMetadata injector.Metadata
			Expect(yaml.Unmarshal(rawMetadata, &actualMetadata)).To(Succeed())

			expectedMetadata := injector.Metadata{
				Releases: append(expectedInitialMetatdata.Releases, injector.Release{
					Name:    releaseName,
					Version: releaseVersion,
					File:    filepath.Base(releasePath),
				}),
				Other: expectedInitialMetatdata.Other,
			}
			Expect(actualMetadata).To(Equal(expectedMetadata))
		})

		Context("failure cases", func() {
			It("returns an error when it fails to create the releases dir", func() {
				fileCannotBeDir, err := ioutil.TempFile(baseTmpDir, "")
				defer fileCannotBeDir.Close()
				Expect(err).NotTo(HaveOccurred())

				err = tileInjector.AddReleaseToTile(releasePath, releaseName, releaseVersion, fileCannotBeDir.Name())
				Expect(err).To(MatchError(ContainSubstring("not a directory")))
			})

			It("returns an error when moving the release into the tile release dir fails", func() {
				err := tileInjector.AddReleaseToTile("/does/not/exist", releaseName, releaseVersion, tileDir)
				Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
			})

			It("returns an error when opening the metadata file fails", func() {
				Expect(os.RemoveAll(filepath.Join(tileDir, "metadata.yml"))).To(Succeed())

				err := tileInjector.AddReleaseToTile(releasePath, releaseName, releaseVersion, tileDir)
				Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
			})

			It("returns an error when reading the metadata file fails", func() {
				Expect(os.RemoveAll(filepath.Join(tileDir, "metadata.yml"))).To(Succeed())
				Expect(os.MkdirAll(filepath.Join(tileDir, "metadata.yml"), 0777)).To(Succeed())

				err := tileInjector.AddReleaseToTile(releasePath, releaseName, releaseVersion, tileDir)
				Expect(err).To(MatchError(ContainSubstring("is a directory")))
			})

			It("returns an error when metadata contains malformed yaml", func() {
				err := ioutil.WriteFile(filepath.Join(tileDir, "metadata.yml"), []byte("%%%%"), 0644)
				Expect(err).NotTo(HaveOccurred())

				err = tileInjector.AddReleaseToTile(releasePath, releaseName, releaseVersion, tileDir)
				Expect(err).To(MatchError(ContainSubstring("yaml: ")))
			})
		})
	})
})
