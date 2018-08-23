package winfsinjector_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestWinfsinjector(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Winfsinjector Suite")
}
