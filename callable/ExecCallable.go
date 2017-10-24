package callable

import (
	"bytes"
	"io"
	"os/exec"

	"github.com/trusch/frunner/env"
)

// ExecCallable implements the callable interface using os/exec
type ExecCallable struct {
	bin         string
	args        []string
	cmd         *exec.Cmd
	errorBuffer *bytes.Buffer
}

// NewExecCallable creates a new exec callable
func NewExecCallable(bin string, args ...string) *ExecCallable {
	errorBuffer := &bytes.Buffer{}
	cmd := exec.Command(bin, args...)
	cmd.Stderr = errorBuffer
	return &ExecCallable{bin, args, cmd, errorBuffer}
}

// Call calls the executable, reading input and writing stdout to output
// stderr will be part of the error message
func (c *ExecCallable) Call(input io.Reader, env env.Env, output io.Writer) chan *CallError {
	c.feedEnv(env)
	errorChannel := c.runCommand(input, env, output)
	return errorChannel
}

// Stop stops the process
func (c *ExecCallable) Stop() error {
	return c.cmd.Process.Kill()
}

// Copy returns a new copy of this callable
func (c *ExecCallable) Copy() Callable {
	return NewExecCallable(c.bin, c.args...)
}

func (c *ExecCallable) runCommand(input io.Reader, env env.Env, output io.Writer) chan *CallError {
	errorChannel := make(chan *CallError)
	c.cmd.Stdin = input
	c.cmd.Stdout = output
	err := c.cmd.Start()
	if err != nil {
		errorChannel <- &CallError{err, c.errorBuffer}
		close(errorChannel)
		return errorChannel
	}
	go func() {
		err = c.cmd.Wait()
		if err != nil {
			errorChannel <- &CallError{err, c.errorBuffer}
		}
		close(errorChannel)
	}()
	return errorChannel
}

func (c *ExecCallable) feedEnv(env env.Env) {
	for key, val := range env {
		c.cmd.Env = append(c.cmd.Env, key+"="+val)
	}
}
