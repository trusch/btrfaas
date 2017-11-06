package example

import (
	"context"
	"io"
)

// Runnable is an example runnable. It implements a Echo-Service which is cancelable via context
type Runnable struct{}

// Run implements the runnable.Runnable interface
func (r *Runnable) Run(ctx context.Context, options map[string]string, input io.Reader, output io.Writer) (err error) {
	buf := make([]byte, 1<<10)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			{
				bs, err := input.Read(buf[:])
				if err != nil {
					if err == io.EOF {
						_, err = output.Write(buf[:bs])
						if err != nil {
							return err
						}
						return nil
					}
					return err
				}
				_, err = output.Write(buf[:bs])
				if err != nil {
					return err
				}
			}
		}
	}
}
