package grpc

import (
	"context"
	"errors"
	"io"

	"google.golang.org/grpc"
)

type Client struct {
	conn   *grpc.ClientConn
	client FunctionRunnerClient
}

func NewClient(target string, opts ...grpc.DialOption) (*Client, error) {
	conn, err := grpc.Dial(target, opts...)
	if err != nil {
		return nil, err
	}
	client := NewFunctionRunnerClient(conn)
	return &Client{conn, client}, nil
}

func (c *Client) Run(ctx context.Context, input io.Reader, output io.Writer) error {
	cli, err := c.client.Run(ctx)
	if err != nil {
		return err
	}

	var (
		readDone  = make(chan struct{})
		sendError error
		readError error
	)

	go func() {
		sendError = c.shovelInputData(cli, input)
	}()

	go func() {
		defer close(readDone)
		readError = c.shovelOutputData(cli, output)
	}()

	for {
		select {
		case <-ctx.Done():
			{
				return ctx.Err()
			}
		case <-readDone:
			{
				return readError
			}
		}
	}
}

func (c *Client) shovelInputData(cli FunctionRunner_RunClient, input io.Reader) error {
	inputBuffer := make([]byte, 4096)
	defer cli.CloseSend()
	ctx := cli.Context()
	for {
		select {
		case <-ctx.Done():
			{
				return ctx.Err()
			}
		default:
			{
				bs, err := input.Read(inputBuffer[:])
				if err != nil && err != io.EOF {
					return err
				}
				if bs > 0 {
					e := cli.Send(&InputData{inputBuffer[:bs]})
					if e != nil {
						return err
					}
				}
				if err == io.EOF {
					return nil
				}
			}
		}
	}
}

func (c *Client) shovelOutputData(cli FunctionRunner_RunClient, output io.Writer) error {
	ctx := cli.Context()
	for {
		select {
		case <-ctx.Done():
			{
				return ctx.Err()
			}
		default:
			{
				data, err := cli.Recv()
				if err != nil && err != io.EOF {
					return err
				}
				if data != nil {
					if len(data.Output) > 0 {
						if _, e := output.Write(data.Output); e != nil {
							return e
						}
					}
					if data.Ready {
						if data.Success {
							return nil
						} else {
							return errors.New(data.ErrorMessage)
						}
					}
					// @TODO: handle data.Errors stream
				}
				if err == io.EOF {
					return nil
				}
			}
		}
	}
}
