package acceptance

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = XDescribe("winfs-injector", func() {
	It("compiles", func() {
		Expect(true).To(BeTrue())
	})
})
