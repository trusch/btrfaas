package exec_test

import (
	"bytes"
	"context"

	"github.com/trusch/frunner/env"
	. "github.com/trusch/frunner/runnable/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Runnable", func() {

	It("should be possible to create and use a Runnable", func() {
		cmd := NewRunnable("cat", "-")
		input := bytes.NewBufferString("foobar")
		output := &bytes.Buffer{}
		Expect(cmd.Run(context.Background(), input, output)).To(Succeed())
		Expect(output.String()).To(Equal("foobar"))
	})

	It("should be possible to create and use a Runnable multiple times", func() {
		cmd := NewRunnable("cat", "-")
		for i := 0; i < 10; i++ {
			input := bytes.NewBufferString("foobar")
			output := &bytes.Buffer{}
			Expect(cmd.Run(context.Background(), input, output)).To(Succeed())
			Expect(output.String()).To(Equal("foobar"))
		}
	})

	It("should be possible to cancel a runnable via context", func() {
		cmd := NewRunnable("sleep", "5")
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		Expect(cmd.Run(ctx, nil, nil)).NotTo(Succeed())
	}, 0.5)

	It("should be possible to pass environment variables in context", func() {
		environment := make(env.Env)
		environment["FOO"] = "bar"
		ctx := env.NewContext(context.Background(), environment)
		cmd := NewRunnable("sh", "-c", "echo -n $FOO")
		output := &bytes.Buffer{}
		Expect(cmd.Run(ctx, nil, output)).To(Succeed())
		Expect(output.String()).To(Equal("bar"))
	})

})
