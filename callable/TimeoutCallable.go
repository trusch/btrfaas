package callable

import (
	"io"
	"time"

	"github.com/trusch/frunner/env"
)

// TimeoutCallable wraps another callable and stops after a specific timeout
type TimeoutCallable struct {
	base     Callable
	timeout  time.Duration
	stopChan chan struct{}
}

// NewTimeoutCallable creates a new timeout callable
func NewTimeoutCallable(base Callable, timeout time.Duration) *TimeoutCallable {
	return &TimeoutCallable{base, timeout, make(chan struct{})}
}

// Call calls the executable, reading input and writing stdout to output
// stderr will be part of the error message
func (c *TimeoutCallable) Call(input io.Reader, env env.Env, output io.Writer) chan *CallError {
	timeoutChannel := time.After(c.timeout)
	go func() {
		select {
		case <-timeoutChannel:
			{
				c.base.Stop()
			}
		case <-c.stopChan:
			{
				return
			}
		}
	}()
	return c.base.Call(input, env, output)
}

// Stop stops the process
func (c *TimeoutCallable) Stop() error {
	close(c.stopChan)
	return c.base.Stop()
}

// Copy copies the callable
func (c *TimeoutCallable) Copy() Callable {
	return NewTimeoutCallable(c.base.Copy(), c.timeout)
}
