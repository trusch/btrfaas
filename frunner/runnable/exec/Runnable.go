package exec

import (
	"bytes"
	"context"
	"io"
	"os/exec"

	log "github.com/Sirupsen/logrus"

	"github.com/trusch/btrfaas/frunner/env"
)

// Runnable implements the Runnable interface using exec.Cmd
type Runnable struct {
	bin          string
	args         []string
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
func (r *Runnable) Run(ctx context.Context, options []string, input io.Reader, output io.Writer) error {
	args := append(r.args, options...)
	cmd := exec.Command(r.bin, args...)
	cmd.Stdin = input
	cmd.Stdout = output
	cmd.Stderr = output
	if r.bufferOutput {
		buf := &bytes.Buffer{}
		cmd.Stdout = buf
	}
	if env, err := env.FromContext(ctx); err == nil {
		cmd.Env = env.ToSlice()
	}
	done := make(chan error)
	go func() {
		done <- cmd.Run()
	}()
	select {
	case err := <-done:
		{
			if err == nil && r.bufferOutput {
				_, err = io.Copy(output, cmd.Stdout.(*bytes.Buffer))
			}
			return err
		}
	case <-ctx.Done():
		{
			if cmd.Process != nil {
				cmd.Process.Kill()
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
