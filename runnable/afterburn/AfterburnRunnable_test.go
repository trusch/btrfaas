package afterburn_test

import (
	"bytes"
	"context"

	"github.com/trusch/frunner/framer"
	. "github.com/trusch/frunner/runnable/afterburn"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AfterburnRunnable", func() {

	It("should be possible to create and use a AfterburnRunnable with afterburn", func() {
		cmd := NewRunnable(&framer.LineFramer{}, "cat", "-")
		input := bytes.NewBufferString("foobar\n")
		output := &bytes.Buffer{}
		Expect(cmd.Run(context.Background(), input, output)).To(Succeed())
		Expect(output.String()).To(Equal("foobar\n"))
	})

	It("should be possible to create and use a AfterburnRunnable multiple times", func() {
		cmd := NewRunnable(&framer.LineFramer{}, "cat", "-")
		input := bytes.NewBufferString("foobar\n")
		output := &bytes.Buffer{}
		Expect(cmd.Run(context.Background(), input, output)).To(Succeed())
		Expect(output.String()).To(Equal("foobar\n"))
		input = bytes.NewBufferString("foobar\n")
		output = &bytes.Buffer{}
		Expect(cmd.Run(context.Background(), input, output)).To(Succeed())
		Expect(output.String()).To(Equal("foobar\n"))
	})

	It("should be possible to cancel a runnable via context", func() {
		cmd := NewRunnable(&framer.LineFramer{}, "sleep", "5")
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		Expect(cmd.Run(ctx, nil, nil)).NotTo(Succeed())
	}, 0.5)

	It("should be possible to cancel a runnable via context and to reuse it", func() {
		cmd := NewRunnable(&framer.LineFramer{}, "cat", "-")
		input := bytes.NewBufferString("foobar\n")
		output := &bytes.Buffer{}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		Expect(cmd.Run(ctx, input, output)).NotTo(Succeed())
		input = bytes.NewBufferString("foobar\n")
		output = &bytes.Buffer{}
		Expect(cmd.Run(context.Background(), input, output)).To(Succeed())
		Expect(output.String()).To(Equal("foobar\n"))
	}, 0.5)

})
