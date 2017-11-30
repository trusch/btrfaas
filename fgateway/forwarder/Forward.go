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
	"google.golang.org/grpc/balancer"
	_ "google.golang.org/grpc/balancer/roundrobin"
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

var clients = make(map[string]*grpc.Client)

// Forward forwards a function call
func Forward(ctx context.Context, options *Options) (err error) {
	log.Debug("construct forwarding pipeline")
	defer func() {
		if err != nil {
			cleanupClients(options)
		}
	}()
	runnables := make([]runnable.Runnable, len(options.Hosts))
	optSlice := make([][]string, len(options.Hosts))
	for i, host := range options.Hosts {
		switch host.Transport {
		case GRPC:
			{
				uri := fmt.Sprintf("dns:///%v:%v", host.Host, host.Port)
				var fn *grpc.Client
				if cli, ok := clients[uri]; ok {
					fn = cli
				} else {
					creds, err := getTransportCredentials(host.Host)
					if err != nil {
						log.Errorf("failed to get credentials for %v: %v", host.Host, err)
						return err
					}
					rr := balancer.Get("round_robin")
					fn, err = grpc.NewClient(uri, creds, g.WithBalancerBuilder(rr))
					if err != nil {
						log.Errorf("failed to get gRPC client for %v: %v", host.Host, err)
						return err
					}
					clients[uri] = fn
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

var (
	creds = make(map[string]credentials.TransportCredentials)
)

func getTransportCredentials(target string) (g.DialOption, error) {
	if _, ok := creds[target]; !ok {
		ca, err := ioutil.ReadFile("/run/secrets/btrfaas-ca-cert.pem")
		if err != nil {
			ca, err = ioutil.ReadFile("/run/secrets/btrfaas-ca-cert.pem/value")
			if err != nil {
				return nil, fmt.Errorf("could not read ca certificate: %s", err)
			}
		}

		certPool := x509.NewCertPool()
		if ok := certPool.AppendCertsFromPEM(ca); !ok {
			return nil, errors.New("failed to append ca certs")
		}

		cert, err := tls.LoadX509KeyPair("/run/secrets/client-cert.pem", "/run/secrets/client-key.pem")
		if err != nil {
			cert, err = tls.LoadX509KeyPair("/run/secrets/client-cert.pem/value", "/run/secrets/client-key.pem/value")
			if err != nil {
				return nil, err
			}
		}
		cfg := &tls.Config{
			ServerName:   target,
			RootCAs:      certPool,
			Certificates: []tls.Certificate{cert},
		}
		cfg.BuildNameToCertificate()
		creds[target] = credentials.NewTLS(cfg)
	}

	return g.WithTransportCredentials(creds[target]), nil
}

func cleanupClients(options *Options) {
	for _, host := range options.Hosts {
		if host.Transport == GRPC {
			key := fmt.Sprintf("dns:///%v:%v", host.Host, host.Port)
			if _, ok := clients[key]; ok {
				delete(clients, key)
			}
		}
	}
}
