package acceptance

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("winfs-injector", func() {
	var (
		inputTile string
		err       error
		outputDir string
	)

	BeforeEach(func() {
		outputDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		inputTile, err = filepath.Abs(filepath.Join(".", "fixtures", "some-tile.pivotal"))
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err = os.RemoveAll(outputDir)
		Expect(err).NotTo(HaveOccurred())
	})

	It("extracts the embedded windows2016fs-release", func() {
		command := exec.Command(pathToMain,
			"--input-tile", inputTile,
			"--output", outputDir,
		)
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))
		sessionContents := strings.TrimSpace(string(session.Out.Contents()))
		Expect(sessionContents).To(ContainSubstring("windows2016fs"))

		path := filepath.Join(sessionContents, "some-tile", "embed", "windows2016fs-release", "some-folder", "some-file.txt")

		path = filepath.Join(sessionContents, "some-tile", "embed")
		_, err = ioutil.ReadDir(path)
		Expect(err).NotTo(HaveOccurred())
	})
})
