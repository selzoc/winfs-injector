package winfsinjector

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestWinFSInjector(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "WinFS Injector Suite")
}
