package forwarder

import (
	"context"
	"errors"
	"fmt"
	"io"

	log "github.com/Sirupsen/logrus"

	"github.com/trusch/btrfaas/frunner/grpc"
	"github.com/trusch/btrfaas/frunner/runnable"
	"github.com/trusch/btrfaas/frunner/runnable/chain"
	g "google.golang.org/grpc"
)

// Options are the options for the forwarding
type Options struct {
	Hosts  []*HostConfig
	Input  io.Reader
	Output io.Writer
}

// HostConfig specifies one function service
type HostConfig struct {
	Transport   TransportProtocol
	Host        string
	Port        uint16
	CallOptions map[string]string
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
	log.Debug("construct forwarding pipeline")
	runnables := make([]runnable.Runnable, len(options.Hosts))
	optSlice := make([]map[string]string, len(options.Hosts))
	for i, host := range options.Hosts {
		switch host.Transport {
		case GRPC:
			{
				uri := fmt.Sprintf("%v:%v", host.Host, host.Port)
				fn, err := grpc.NewClient(uri, g.WithInsecure())
				if err != nil {
					return err
				}
				runnables[i] = fn
				optSlice[i] = host.CallOptions
				log.Debugf("added grpc://%v to the pipeline", uri)
			}
		case HTTP:
			{
				fn := NewHTTPRunnable(fmt.Sprintf("http://%v:%v", host.Host, host.Port))
				runnables[i] = fn
				optSlice[i] = host.CallOptions
			}
		default:
			{
				return errors.New("transport not implemented")
			}
		}
	}
	cmd := chain.New(runnables...)
	log.Debug("finished constructing pipeline, kickoff...")
	return cmd.Run(ctx, optSlice, options.Input, options.Output)
}
