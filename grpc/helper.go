package grpc

import (
	"context"
	"io"
)

// DataStream is either a FunctionRunner_RunClient or FunctionRunner_RunServer
type DataStream interface {
	Send(*Data) error
	Recv() (*Data, error)
}

// CopyToStream copies from a reader to a stream
func CopyToStream(ctx context.Context, source io.Reader, dest DataStream) error {
	buf := make([]byte, 4096)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			{
				bs, err := source.Read(buf[:])
				if bs > 0 {
					if sendError := dest.Send(&Data{Data: buf[:bs]}); sendError != nil {
						return sendError
					}
				}
				if err != nil {
					if err == io.EOF {
						return nil
					}
					return err
				}
			}
		}
	}
}

// CopyFromStream copies data from a stream to a writer
func CopyFromStream(ctx context.Context, source DataStream, dest io.Writer) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			{
				data, err := source.Recv()
				if data != nil && len(data.Data) > 0 {
					if _, writeErr := dest.Write(data.Data); writeErr != nil {
						return writeErr
					}
				}
				if err != nil {
					if err == io.EOF {
						return nil
					}
					return err
				}
			}
		}
	}
}
