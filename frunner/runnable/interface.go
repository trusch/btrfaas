package runnable

import (
	"context"
	"io"
)

// Runnable is the interface for all runnable things (processes, functions...)
type Runnable interface {
	Run(ctx context.Context, options map[string]string, input io.Reader, output io.Writer) error
}
