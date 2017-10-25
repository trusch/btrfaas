package callable_test

import (
	"bytes"

	. "github.com/trusch/frunner/callable"
	"github.com/trusch/frunner/env"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BufferingExecCallable", func() {
	It("should be possible to create and exec an BufferingExecCallable", func() {
		c := NewBufferingExecCallable("cat", "-")
		input := bytes.NewBufferString("foobar")
		output := &bytes.Buffer{}
		env := make(env.Env)
		errorChannel := c.Call(input, env, output)
		Expect(errorChannel).NotTo(BeNil())
		Expect(<-errorChannel).To(BeNil())
		Expect(output.String()).To(Equal("foobar"))
	})

	It("should be possible to stop an BufferingExecCallable", func() {
		c := NewBufferingExecCallable("tail", "-f", "/dev/null")
		input := &bytes.Buffer{}
		output := &bytes.Buffer{}
		env := make(env.Env)
		errorChannel := c.Call(input, env, output)
		Expect(errorChannel).NotTo(BeNil())
		Expect(c.Stop()).To(Succeed())
		Expect(<-errorChannel).NotTo(BeNil())
	})

	It("should be possible to copy an BufferingExecCallable", func() {
		c := NewBufferingExecCallable("cat", "-")
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
