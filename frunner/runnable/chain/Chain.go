package chain

import (
	"context"
	"io"

	"github.com/trusch/btrfaas/frunner/runnable"
)

// Chain is a chain of runnables
type Chain struct {
	runnables []runnable.Runnable
}

// New creates a new chain from a number of runnables
func New(cmd ...runnable.Runnable) *Chain {
	return &Chain{cmd}
}

// Run implements the runnable.Runnable interface
func (c *Chain) Run(ctx context.Context, options [][]string, input io.Reader, output io.Writer) error {
	var currentReader = input

	// cancel everything when something goes wrong
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	done := make(chan error, len(c.runnables))

	// kick of runnables and chain via io.Pipe
	for i := 0; i < len(c.runnables); i++ {
		var opts []string
		if i < len(options) {
			opts = options[i]
		}
		currentReader = runChained(ctx, c.runnables[i], opts, currentReader, done)
	}

	// shovel last output to the output of this runnable
	go func() {
		_, err := io.Copy(output, currentReader)
		done <- err
	}()

	// wait for everything to finish
	for i := 0; i < len(c.runnables)+1; i++ {
		err := <-done
		if err != nil {
			return err
		}
	}

	return nil
}

func runChained(ctx context.Context, cmd runnable.Runnable, options []string, input io.Reader, done chan error) io.Reader {
	pipeReader, pipeWriter := io.Pipe()
	go func() {
		// IMPORTANT: pipes need to be closed to send EOF
		defer pipeWriter.Close()
		done <- cmd.Run(ctx, options, input, pipeWriter)
	}()
	return pipeReader
}
