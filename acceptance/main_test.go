package acceptance

import (
	"io/ioutil"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("winfs-injector", func() {
	var (
		inputTile string
		err       error
	)

	BeforeEach(func() {
		inputTile, err = filepath.Abs("./fixtures/some-tile.pivotal")
		Expect(err).NotTo(HaveOccurred())
	})

	It("extracts the embedded windows2016fs-release", func() {
		command := exec.Command(pathToMain,
			"--input-tile", inputTile,
		)
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		//TODO: make this better
		Eventually(session).Should(gexec.Exit(0))
		sessionContents := session.Out.Contents()
		Expect(string(sessionContents)).To(ContainSubstring("/tmp/"))

		_, err = ioutil.ReadDir(string(sessionContents))
		Expect(err).NotTo(HaveOccurred())
	})
})
