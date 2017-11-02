package exec

import (
	"bytes"
	"context"
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/trusch/btrfaas/frunner/env"
)

// Runnable implements the Runnable interface using exec.Cmd
type Runnable struct {
	bin          string
	args         []string
	cmd          *exec.Cmd
	bufferOutput bool
}

// NewRunnable creates a new Runnable instance
func NewRunnable(bin string, args ...string) *Runnable {
	return &Runnable{
		bin:  bin,
		args: args,
	}
}

// Run implements the Runnable interface
func (r *Runnable) Run(ctx context.Context, input io.Reader, output io.Writer) (err error) {
	r.cmd = exec.Command(r.bin, r.args...)
	r.cmd.Stdin = input
	r.cmd.Stdout = output
	r.cmd.Stderr = os.Stderr
	if r.bufferOutput {
		buf := &bytes.Buffer{}
		r.cmd.Stdout = buf
	}
	if env, err := env.FromContext(ctx); err == nil {
		r.cmd.Env = env.ToSlice()
	}
	done := make(chan struct{}, 1)
	go func() {
		err = r.cmd.Run()
		close(done)
	}()
	select {
	case <-done:
		{
			if err == nil && r.bufferOutput {
				_, err = io.Copy(output, r.cmd.Stdout.(*bytes.Buffer))
			}
			return
		}
	case <-ctx.Done():
		{
			if r.cmd.Process != nil {
				r.cmd.Process.Kill()
				log.Print("process got killed because of: ", ctx.Err())
			}
			return ctx.Err()
		}
	}
}

// EnableOutputBuffering ensures that nothing is written to the output in case of an error
func (r *Runnable) EnableOutputBuffering() {
	r.bufferOutput = true
}
