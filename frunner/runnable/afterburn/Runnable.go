package afterburn

import (
	"context"
	"io"
	"log"

	"github.com/trusch/btrfaas/frunner/framer"
	"github.com/trusch/btrfaas/frunner/runnable"
	"github.com/trusch/btrfaas/frunner/runnable/exec"
)

// Runnable implements the Runnable interface using exec.Cmd
type Runnable struct {
	cmd    runnable.Runnable
	ctx    context.Context
	cancel context.CancelFunc
	input  io.Writer
	output io.Reader
	framer framer.Framer
}

// NewRunnable creates a new Runnable instance
func NewRunnable(framer framer.Framer, bin string, args ...string) *Runnable {
	r := &Runnable{framer: framer}
	ir, iw := io.Pipe()
	or, ow := io.Pipe()
	r.input = iw
	r.output = or
	initReady := false
	initReadyChan := make(chan struct{}, 1)
	go func() {
		for { // recover from errors / timeouts
			cmd := exec.NewRunnable(bin, args...)
			ctx, cancel := context.WithCancel(context.Background())
			r.ctx = ctx
			r.cancel = cancel
			r.cmd = cmd
			if !initReady {
				initReady = true
				close(initReadyChan)
			}
			if err := cmd.Run(ctx, ir, ow); err != nil {
				log.Print(err)
			}
		}
	}()
	<-initReadyChan
	return r
}

// Run implements the Runnable interface
func (r *Runnable) Run(ctx context.Context, input io.Reader, output io.Writer) (err error) {
	done := make(chan struct{}, 1)
	go func() {
		if input != nil {
			go func() { _, err = io.Copy(r.input, input); log.Print("finished with input") }()
		}
		if output != nil {
			err = r.framer.Copy(output, r.output)
		}
		close(done)
	}()
	select {
	case <-done:
		{
			return
		}
	case <-ctx.Done():
		{
			r.cancel()
			<-done
			return ctx.Err()
		}
	}
}
