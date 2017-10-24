package callable

import (
	"bytes"
	"io"

	"github.com/trusch/frunner/env"
)

// ExampleCallable is a reference implementation of a Callable
type ExampleCallable struct{}

// Call implements the Callable interface
func (c *ExampleCallable) Call(input io.Reader, env env.Env, output io.Writer) chan *CallError {
	errorChan := make(chan *CallError)
	go func() {
		if _, err := io.Copy(output, input); err != nil {
			errorChan <- &CallError{err, bytes.NewBufferString("failed to copy")}
			close(errorChan)
		}
		close(errorChan)
	}()
	return errorChan
}

// Stop implements the callable interface, this is a no op here but you should implement it!
func (c *ExampleCallable) Stop() error {
	return nil
}

// Copy copies the callable
func (c *ExampleCallable) Copy() Callable {
	return &ExampleCallable{}
}
