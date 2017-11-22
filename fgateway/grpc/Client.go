package grpc

import (
	"context"
	"encoding/json"
	"io"

	btrfaasgrpc "github.com/trusch/btrfaas/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// Client is a gRPC client to the fgateway
type Client struct {
	conn   *grpc.ClientConn
	client btrfaasgrpc.FunctionRunnerClient
}

// NewClient creates a new client instance
func NewClient(gateway string, opts ...grpc.DialOption) (*Client, error) {
	conn, err := grpc.Dial(gateway, opts...)
	if err != nil {
		return nil, err
	}
	client := btrfaasgrpc.NewFunctionRunnerClient(conn)
	return &Client{conn, client}, nil
}

// Run nearly implements the runnable interface, except that it supports specifying chains of functions instead of a single function
func (c *Client) Run(ctx context.Context, chain []string, options [][]string, input io.Reader, output io.Writer) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, metadata.MD{
		"chain":   chain,
		"options": buildOptionsForMetadata(options),
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

// Close closes the client connection
func (c *Client) Close() error {
	return c.conn.Close()
}

func buildOptionsForMetadata(options [][]string) (res []string) {
	for _, v := range options {
		bs, _ := json.Marshal(v)
		res = append(res, string(bs))
	}
	return
}
