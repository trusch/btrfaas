package callable_test

import (
	"bytes"
	"time"

	. "github.com/trusch/frunner/callable"
	"github.com/trusch/frunner/env"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TimeoutCallable", func() {
	It("should be possible to specify a timeout in a TimeoutCallable", func() {
		c := NewTimeoutCallable(NewExecCallable("tail", "-f", "/dev/null"), 100*time.Millisecond)
		input := &bytes.Buffer{}
		output := &bytes.Buffer{}
		env := make(env.Env)
		errorChannel := c.Call(input, env, output)
		Expect(errorChannel).NotTo(BeNil())
		Expect(<-errorChannel).NotTo(BeNil())
	}, 0.2)

	It("should be possible to stop an TimeoutCallable", func() {
		c := NewTimeoutCallable(NewExecCallable("tail", "-f", "/dev/null"), 100*time.Millisecond)
		input := &bytes.Buffer{}
		output := &bytes.Buffer{}
		env := make(env.Env)
		errorChannel := c.Call(input, env, output)
		Expect(errorChannel).NotTo(BeNil())
		Expect(c.Stop()).To(Succeed())
		Expect(<-errorChannel).NotTo(BeNil())
	}, 0.2)

	It("should be possible to copy an TimeoutCallable", func() {
		c := NewTimeoutCallable(NewExecCallable("cat", "-"), 100*time.Millisecond)
		input := bytes.NewBufferString("foobar")
		output := &bytes.Buffer{}
		env := make(env.Env)
		errorChannel := c.Call(input, env, output)
		Expect(errorChannel).NotTo(BeNil())
		Expect(<-errorChannel).To(BeNil())
		Expect(output.String()).To(Equal("foobar"))
		c2 := c.Copy()
		input = bytes.NewBufferString("foobar")
		output = &bytes.Buffer{}
		errorChannel = c2.Call(input, env, output)
		Expect(errorChannel).NotTo(BeNil())
		Expect(<-errorChannel).To(BeNil())
		Expect(output.String()).To(Equal("foobar"))
	})
})
