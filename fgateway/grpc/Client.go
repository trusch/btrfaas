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

func NewClient(gateway string, opts ...grpc.DialOption) (*Client, error) {
	conn, err := grpc.Dial(gateway, opts...)
	if err != nil {
		return nil, err
	}
	client := NewFunctionRunnerClient(conn)
	return &Client{conn, client}, nil
}

func (c *Client) Run(ctx context.Context, functionID string, input io.Reader, output io.Writer) error {
	cli, err := c.client.Run(ctx)
	if err != nil {
		return err
	}

	var (
		readDone  = make(chan struct{})
		sendError error
		readError error
	)

	if err := cli.Send(&GWInputData{FunctionID: functionID}); err != nil {
		return err
	}

	go func() {
		sendError = c.shovelGWInputData(cli, input)
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

func (c *Client) shovelGWInputData(cli FunctionRunner_RunClient, input io.Reader) error {
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
					e := cli.Send(&GWInputData{Data: inputBuffer[:bs]})
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
						}
						return errors.New(data.ErrorMessage)
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
