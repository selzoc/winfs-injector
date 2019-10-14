package acceptance_test

import (
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/p-windows-runtime-2016/windowsfs-release/src/github.com/onsi/gomega/gexec"
)

var _ = Describe("acceptance", func() {
	Describe("main", func() {
		var (
			winfsInjector string
			cmd           *exec.Cmd
			inputTile     string
			outputTile    string
		)
		BeforeEach(func() {
			var err error
			winfsInjector, err = gexec.Build("github.com/pivotal-cf/winfs-injector")
			Expect(err).ToNot(HaveOccurred())

			inputTile = "input-tile-path"
			outputTile = "output-tile-path"
		})

		AfterEach(func() {
			Expect(os.Remove(winfsInjector)).NotTo(HaveOccurred())
		})

		It("requires an input tile path", func() {
			cmd = exec.Command(winfsInjector, "-o", outputTile)
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(1))
			Expect(string(session.Err.Contents())).To(ContainSubstring("--input-tile is required"))
		})

		It("requires an output tile path", func() {
			cmd = exec.Command(winfsInjector, "-i", inputTile)
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(1))
			Expect(string(session.Err.Contents())).To(ContainSubstring("--output-tile is required"))
		})

		It("prints usage when the help flag is provided", func() {
			cmd = exec.Command(winfsInjector, "--help")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))
			Expect(string(session.Out.Contents())).To(ContainSubstring(`
  --input-tile, -i   path to input tile (example: /path/to/input.pivotal)
  --output-tile, -o  path to output tile (example: /path/to/output.pivotal)
  --registry, -r     path to docker registry (example: /path/to/registry, default: "https://registry.hub.docker.com")
  --help, -h         prints this usage information`))
		})
	})
})
