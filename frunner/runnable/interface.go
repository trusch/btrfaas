package runnable

import (
	"context"
	"io"
)

// Runnable is the interface for all runnable things (processes, functions...)
type Runnable interface {
	Run(ctx context.Context, options []string, input io.Reader, output io.Writer) error
}
