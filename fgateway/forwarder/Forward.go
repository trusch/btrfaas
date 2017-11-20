package forwarder

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"io/ioutil"

	log "github.com/Sirupsen/logrus"

	"github.com/trusch/btrfaas/frunner/grpc"
	"github.com/trusch/btrfaas/frunner/runnable"
	"github.com/trusch/btrfaas/frunner/runnable/chain"
	g "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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
	CallOptions []string
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
	optSlice := make([][]string, len(options.Hosts))
	for i, host := range options.Hosts {
		switch host.Transport {
		case GRPC:
			{
				uri := fmt.Sprintf("%v:%v", host.Host, host.Port)
				creds, err := getTransportCredentials(host.Host)
				if err != nil {
					return err
				}
				fn, err := grpc.NewClient(uri, creds)
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

var certPool *x509.CertPool

func getTransportCredentials(target string) (g.DialOption, error) {
	if certPool == nil {
		ca, err := ioutil.ReadFile("/run/secrets/btrfaas-ca-cert.pem")
		if err != nil {
			return nil, fmt.Errorf("could not read ca certificate: %s", err)
		}
		certPool = x509.NewCertPool()
		// Append the certificates from the CA
		if ok := certPool.AppendCertsFromPEM(ca); !ok {
			return nil, errors.New("failed to append ca certs")
		}
	}

	creds := credentials.NewTLS(&tls.Config{
		ServerName: target,
		RootCAs:    certPool,
	})

	return g.WithTransportCredentials(creds), nil
}
