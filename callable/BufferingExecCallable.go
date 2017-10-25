package callable

import (
	"bytes"
	"errors"
	"io"
	"os/exec"

	"github.com/trusch/frunner/env"
)

// BufferingExecCallable implements the callable interface using os/exec and buffers the output.
// -> no writes to the output writer are done if the process finished with an error
type BufferingExecCallable struct {
	bin         string
	args        []string
	cmd         *exec.Cmd
	errorBuffer *bytes.Buffer
}

// NewBufferingExecCallable creates a new exec callable
func NewBufferingExecCallable(bin string, args ...string) *BufferingExecCallable {
	errorBuffer := &bytes.Buffer{}
	cmd := exec.Command(bin, args...)
	cmd.Stderr = errorBuffer
	return &BufferingExecCallable{bin, args, cmd, errorBuffer}
}

// Call calls the executable, reading input and writing stdout to output
// stderr will be part of the error message
func (c *BufferingExecCallable) Call(input io.Reader, env env.Env, output io.Writer) chan *CallError {
	c.feedEnv(env)
	errorChannel := c.runCommand(input, env, output)
	return errorChannel
}

// Stop stops the process
func (c *BufferingExecCallable) Stop() error {
	if c.cmd == nil || c.cmd.Process == nil {
		return errors.New("process not running")
	}
	return c.cmd.Process.Kill()
}

// Copy returns a new copy of this callable
func (c *BufferingExecCallable) Copy() Callable {
	return NewBufferingExecCallable(c.bin, c.args...)
}

func (c *BufferingExecCallable) runCommand(input io.Reader, env env.Env, output io.Writer) chan *CallError {
	errorChannel := make(chan *CallError)
	c.cmd.Stdin = input
	c.cmd.Stdout = &bytes.Buffer{}
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
			close(errorChannel)
			return
		}
		if _, err := io.Copy(output, c.cmd.Stdout.(io.Reader)); err != nil {
			errorChannel <- &CallError{err, nil}
		}
		close(errorChannel)
	}()
	return errorChannel
}

func (c *BufferingExecCallable) feedEnv(env env.Env) {
	for key, val := range env {
		c.cmd.Env = append(c.cmd.Env, key+"="+val)
	}
}
