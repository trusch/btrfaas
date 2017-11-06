package chain_test

import (
	"bytes"
	"context"

	. "github.com/trusch/btrfaas/frunner/runnable/chain"
	"github.com/trusch/btrfaas/frunner/runnable/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Chain", func() {
	It("should be possible to declare and use a chain", func() {
		echo := exec.NewRunnable("cat", "-")
		toUpper := exec.NewRunnable("tr", "[:lower:]", "[:upper:]")
		chainedRunnable := New(echo, toUpper)
		input := bytes.NewBufferString("foobar")
		output := &bytes.Buffer{}
		Expect(chainedRunnable.Run(context.Background(), nil, input, output)).To(Succeed())
		Expect(output.String()).To(Equal("FOOBAR"))
	})
})
