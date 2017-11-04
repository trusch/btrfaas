package forwarder

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/trusch/btrfaas/frunner/grpc"
	g "google.golang.org/grpc"
)

// Options are the options for the forwarding
type Options struct {
	Transport TransportProtocol
	Host      string
	Port      uint16
	Input     io.Reader
	Output    io.Writer
}

// TransportProtocol is the type of the transport, currently only GRPC is supported
type TransportProtocol int

const (
	// GRPC represents a gRPC transport layer
	GRPC TransportProtocol = iota
	// HTTP represents a http transport layer
	HTTP
)

// Forward forwards a function call
func Forward(ctx context.Context, options *Options) error {
	if options.Transport == GRPC {
		uri := fmt.Sprintf("%v:%v", options.Host, options.Port)
		function, err := grpc.NewClient(uri, g.WithInsecure())
		if err != nil {
			return err
		}
		return function.Run(ctx, options.Input, options.Output)
	}
	return errors.New("transport protocol not implemented")
}
