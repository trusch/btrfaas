package callable

import (
	"bytes"
	"errors"
	"io"
	"os/exec"

	"github.com/trusch/frunner/env"
)

// PipingExecCallable implements the callable interface using os/exec
type PipingExecCallable struct {
	bin         string
	args        []string
	cmd         *exec.Cmd
	errorBuffer *bytes.Buffer
}

// NewPipingExecCallable creates a new exec callable
func NewPipingExecCallable(bin string, args ...string) *PipingExecCallable {
	errorBuffer := &bytes.Buffer{}
	cmd := exec.Command(bin, args...)
	cmd.Stderr = errorBuffer
	return &PipingExecCallable{bin, args, cmd, errorBuffer}
}

// Call calls the executable, reading input and writing stdout to output
// stderr will be part of the error message
func (c *PipingExecCallable) Call(input io.Reader, env env.Env, output io.Writer) chan *CallError {
	c.feedEnv(env)
	errorChannel := c.runCommand(input, env, output)
	return errorChannel
}

// Stop stops the process
func (c *PipingExecCallable) Stop() error {
	if c.cmd == nil || c.cmd.Process == nil {
		return errors.New("process not running")
	}
	return c.cmd.Process.Kill()
}

// Copy returns a new copy of this callable
func (c *PipingExecCallable) Copy() Callable {
	return NewPipingExecCallable(c.bin, c.args...)
}

func (c *PipingExecCallable) runCommand(input io.Reader, env env.Env, output io.Writer) chan *CallError {
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

func (c *PipingExecCallable) feedEnv(env env.Env) {
	for key, val := range env {
		c.cmd.Env = append(c.cmd.Env, key+"="+val)
	}
}
