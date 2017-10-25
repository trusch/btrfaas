package callable

import (
	"bytes"
	"errors"
	"io"
	"log"
	"os/exec"

	"github.com/trusch/frunner/env"
	"github.com/trusch/frunner/framer"
)

// AfterburnExecCallable implements the callable interface using os/exec and buffers the output.
// -> no writes to the output writer are done if the process finished with an error
type AfterburnExecCallable struct {
	bin           string
	args          []string
	cmd           *exec.Cmd
	running       bool
	framer        framer.Framer
	errorBuffer   *bytes.Buffer
	inPipeWriter  io.WriteCloser
	outPipeReader io.ReadCloser
}

// NewAfterburnExecCallable creates a new exec callable
func NewAfterburnExecCallable(framer framer.Framer, bin string, args ...string) *AfterburnExecCallable {
	errorBuffer := &bytes.Buffer{}
	cmd := exec.Command(bin, args...)
	inPipeReader, inPipeWriter := io.Pipe()
	outPipeReader, outPipeWriter := io.Pipe()
	cmd.Stderr = errorBuffer
	cmd.Stdin = inPipeReader
	cmd.Stdout = outPipeWriter
	return &AfterburnExecCallable{bin, args, cmd, false, framer, errorBuffer, inPipeWriter, outPipeReader}
}

// Call calls the executable, reading input and writing stdout to output
// stderr will be part of the error message
func (c *AfterburnExecCallable) Call(input io.Reader, env env.Env, output io.Writer) chan *CallError {
	c.feedEnv(env)
	errorChannel := c.runCommand(input, env, output)
	if !c.running {
		if err := c.cmd.Start(); err != nil {
			errorChannel <- &CallError{err, nil}
			return errorChannel
		}
		go func() {
			if err := c.cmd.Wait(); err != nil {
				log.Print("command exited")
			}
			c.running = false
		}()
	}
	return errorChannel
}

// Stop stops the process
func (c *AfterburnExecCallable) Stop() error {
	if c.cmd == nil || c.cmd.Process == nil {
		return errors.New("process not running")
	}
	if err := c.inPipeWriter.Close(); err != nil {
		return err
	}
	if err := c.outPipeReader.Close(); err != nil {
		return err
	}
	return c.cmd.Process.Kill()
}

// Copy returns a new copy of this callable
func (c *AfterburnExecCallable) Copy() Callable {
	return NewAfterburnExecCallable(c.framer, c.bin, c.args...)
}

func (c *AfterburnExecCallable) runCommand(input io.Reader, env env.Env, output io.Writer) chan *CallError {
	errorChannel := make(chan *CallError)
	go func() {
		defer close(errorChannel)
		if _, err := io.Copy(c.inPipeWriter, input); err != nil {
			errorChannel <- &CallError{err, nil}
			return
		}
		if err := c.framer.Copy(output, c.outPipeReader); err != nil {
			errorChannel <- &CallError{err, nil}
		}
	}()
	return errorChannel
}

func (c *AfterburnExecCallable) feedEnv(env env.Env) {
	for key, val := range env {
		c.cmd.Env = append(c.cmd.Env, key+"="+val)
	}
}
