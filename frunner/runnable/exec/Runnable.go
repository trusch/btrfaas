package exec

import (
	"bytes"
	"context"
	"io"
	"os/exec"

	log "github.com/sirupsen/logrus"

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
func (r *Runnable) Run(ctx context.Context, options map[string]string, input io.Reader, output io.Writer) error {
	args := append(r.args, constructArgsFromOptions(options)...)
	r.cmd = exec.Command(r.bin, args...)
	r.cmd.Stdin = input
	r.cmd.Stdout = output
	r.cmd.Stderr = output
	if r.bufferOutput {
		buf := &bytes.Buffer{}
		r.cmd.Stdout = buf
	}
	if env, err := env.FromContext(ctx); err == nil {
		r.cmd.Env = env.ToSlice()
	}
	done := make(chan error)
	go func() {
		done <- r.cmd.Run()
	}()
	select {
	case err := <-done:
		{
			if err == nil && r.bufferOutput {
				_, err = io.Copy(output, r.cmd.Stdout.(*bytes.Buffer))
			}
			log.Error(err)
			return err
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

func constructArgsFromOptions(options map[string]string) []string {
	res := make([]string, 0, 2*len(options))
	for k, v := range options {
		if len(k) == 1 {
			res = append(res, "-"+k, v)
		} else {
			res = append(res, "--"+k, v)
		}
	}
	return res
}
