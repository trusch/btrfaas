package template_test

import (
	"os"

	. "github.com/trusch/btrfaas/template"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Load", func() {
	It("should be possible to load a subfolder in a git repo", func() {
		Expect(Load("https://github.com/trusch/btrfaas.git/examples/btrfaas/native-functions/echo-go", "/tmp/btrfaas")).To(Succeed())
		os.RemoveAll("/tmp/btrfaas")
	})
})
