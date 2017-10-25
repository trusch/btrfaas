package callable

import (
	"bytes"
	"errors"
	"io"
	"log"
	"os/exec"
	"syscall"
	"time"

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
	return &AfterburnExecCallable{bin, args, nil, false, framer, nil, nil, nil}
}

// Call calls the executable, reading input and writing stdout to output
// stderr will be part of the error message
func (c *AfterburnExecCallable) Call(input io.Reader, env env.Env, output io.Writer) chan *CallError {
	if !c.running {
		log.Print("kick of process")
		errorBuffer := &bytes.Buffer{}
		c.cmd = exec.Command(c.bin, c.args...)
		inPipeReader, inPipeWriter := io.Pipe()
		outPipeReader, outPipeWriter := io.Pipe()
		c.cmd.Stderr = errorBuffer
		c.cmd.Stdin = inPipeReader
		c.cmd.Stdout = outPipeWriter
		c.inPipeWriter = inPipeWriter
		c.outPipeReader = outPipeReader
		if err := c.cmd.Start(); err != nil {
			ch := make(chan *CallError, 1)
			ch <- &CallError{err, nil}
			return ch
		}
		c.running = true
		go func() {
			for {
				time.Sleep(100 * time.Millisecond)
				if err := syscall.Kill(c.cmd.Process.Pid, syscall.Signal(0)); err != nil {
					log.Print("process died ", c.cmd.ProcessState)
					c.running = false
					c.outPipeReader.Close()
					break
				}
			}
		}()
		go func() {
			if err := c.cmd.Wait(); err != nil {
				log.Print("command exited with error: ", err)
			}
			log.Print("command exited")
			c.running = false
			c.inPipeWriter.Close()
			c.outPipeReader.Close()
		}()
	}
	c.feedEnv(env)
	errorChannel := c.runCommand(input, output)
	return errorChannel
}

// Stop stops the process
func (c *AfterburnExecCallable) Stop() error {
	if c.cmd == nil || c.cmd.Process == nil {
		return errors.New("process not running")
	}
	return c.cmd.Process.Kill()
}

// Copy returns a new copy of this callable
func (c *AfterburnExecCallable) Copy() Callable {
	return c
}

func (c *AfterburnExecCallable) runCommand(input io.Reader, output io.Writer) chan *CallError {
	errorChannel := make(chan *CallError)
	go func() {
		defer close(errorChannel)
		if _, err := io.Copy(c.inPipeWriter, input); err != nil {
			log.Print("error writing input: ", err)
			errorChannel <- &CallError{err, nil}
			return
		}
		if err := c.framer.Copy(output, c.outPipeReader); err != nil {
			log.Print("error parsing output: ", err)
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
