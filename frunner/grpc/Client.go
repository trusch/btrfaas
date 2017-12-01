package grpc

import (
	"context"
	"io"

	btrfaasgrpc "github.com/trusch/btrfaas/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// Client is a gRPC client for the function runner interface implementing Runnable
type Client struct {
	conn   *grpc.ClientConn
	client btrfaasgrpc.FunctionRunnerClient
}

// NewClient returns a new client instance
func NewClient(target string, opts ...grpc.DialOption) (*Client, error) {
	conn, err := grpc.Dial(target, opts...)
	if err != nil {
		return nil, err
	}
	client := btrfaasgrpc.NewFunctionRunnerClient(conn)
	return &Client{conn, client}, nil
}

// NewClientWithContext returns a new client instance
func NewClientWithContext(ctx context.Context, target string, opts ...grpc.DialOption) (*Client, error) {
	conn, err := grpc.DialContext(ctx, target, opts...)
	if err != nil {
		return nil, err
	}
	client := btrfaasgrpc.NewFunctionRunnerClient(conn)
	return &Client{conn, client}, nil
}

// Close closes the client connection
func (c *Client) Close() error {
	return c.conn.Close()
}

// Run implements the Runnable interface
func (c *Client) Run(ctx context.Context, options []string, input io.Reader, output io.Writer) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.MD{
		"options": options,
	})
	cli, err := c.client.Run(ctx)
	if err != nil {
		return err
	}

	done := make(chan error, 3)
	go func() {
		done <- btrfaasgrpc.CopyToStream(ctx, input, cli)
		done <- cli.CloseSend()
	}()
	go func() {
		done <- btrfaasgrpc.CopyFromStream(ctx, cli, output)
	}()

	todo := 3 // send done, close-send done, read done
	for {
		select {
		case <-ctx.Done():
			{
				return ctx.Err()
			}
		case err := <-done:
			{
				if err != nil {
					return err
				}
				todo--
				if todo == 0 {
					return nil
				}
			}
		}
	}
}
