package grpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/url"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/trusch/btrfaas/fgateway/forwarder"
	"github.com/trusch/btrfaas/fgateway/metrics"
	btrfaasgrpc "github.com/trusch/btrfaas/grpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

// Server represents a gRPC based function dispatcher
type Server struct {
	addr        string
	defaultPort uint16
	grpcOpts    []grpc.ServerOption
}

// NewServer creates a gRPC based function dispatcher
func NewServer(addr string, defaultPort uint16, opts ...grpc.ServerOption) *Server {
	return &Server{addr, defaultPort, opts}
}

// ListenAndServe starts listening for connections
func (s *Server) ListenAndServe() error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	certificate, err := tls.LoadX509KeyPair("/run/secrets/fgateway-cert.pem", "/run/secrets/fgateway-key.pem")
	if err != nil {
		certificate, err = tls.LoadX509KeyPair("/run/secrets/fgateway-cert.pem/value", "/run/secrets/fgateway-key.pem/value") // k8s specific -.-
		if err != nil {
			return fmt.Errorf("could not load server key pair: %s", err)
		}
	}

	// Create a certificate pool from the certificate authority
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile("/run/secrets/btrfaas-ca-cert.pem")
	if err != nil {
		ca, err = ioutil.ReadFile("/run/secrets/btrfaas-ca-cert.pem/value") // k8s specific -.-
		if err != nil {
			return fmt.Errorf("could not read ca certificate: %s", err)
		}
	}

	// Append the client certificates from the CA
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		return errors.New("failed to append client certs")
	}

	// Create the TLS credentials
	creds := credentials.NewTLS(&tls.Config{
		ClientAuth:   tls.VerifyClientCertIfGiven,
		Certificates: []tls.Certificate{certificate},
		ClientCAs:    certPool,
	})
	grpcServer := grpc.NewServer(grpc.Creds(creds))

	btrfaasgrpc.RegisterFunctionRunnerServer(grpcServer, s)
	return grpcServer.Serve(lis)
}

// Run implements the gRPC interface
func (s *Server) Run(stream btrfaasgrpc.FunctionRunner_RunServer) (err error) {
	log.Debug("new gRPC request")
	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()
	start := time.Now()

	chain, options, err := getOptionsFromStream(stream)
	if err != nil {
		return err
	}
	hosts, err := s.createHostConfigs(chain, options)
	if err != nil {
		return err
	}
	defer func() {
		end := time.Now()
		duration := end.Sub(start)
		log.Debugf("finished gRPC request in %v", duration)
		for _, host := range hosts {
			metrics.Observe(host.Host, err != nil, duration)
		}
	}()

	inputReader, inputWriter := io.Pipe()
	outputReader, outputWriter := io.Pipe()

	done := make(chan error, 5)

	go func() {
		log.Debug("forward to function services ", chain)
		done <- forwarder.Forward(stream.Context(), &forwarder.Options{
			Hosts:  hosts,
			Input:  inputReader,
			Output: outputWriter,
		})
		done <- outputWriter.Close()
	}()

	go func() {
		done <- btrfaasgrpc.CopyFromStream(ctx, stream, inputWriter)
		done <- inputWriter.Close()
	}()

	go func() {
		done <- btrfaasgrpc.CopyToStream(ctx, outputReader, stream)
	}()

	todo := 5
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-done:
			{
				if err != nil {
					log.Debugf("finished with error: %v", err)
					return fmt.Errorf("fgateway: %v", err)
				}
				todo--
				if todo == 0 {
					return nil
				}
			}
		}
	}
}

func getOptionsFromStream(stream btrfaasgrpc.FunctionRunner_RunServer) (chain []string, optionSlice [][]string, err error) {
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return nil, nil, errors.New("no metadata")
	}
	chain, ok = md["chain"]
	if !ok {
		return nil, nil, errors.New("no chain in metadata")
	}
	optionsList, ok := md["options"]
	if !ok {
		return chain, nil, nil
	}
	optionSlice = make([][]string, len(optionsList))
	for idx, objStr := range optionsList {
		var options []string
		if err := json.Unmarshal([]byte(objStr), &options); err != nil {
			return chain, nil, err
		}
		optionSlice[idx] = options
	}
	if len(chain) != len(optionSlice) {
		return nil, nil, errors.New("chain/option count mismatch")
	}
	return chain, optionSlice, nil
}

func (s *Server) createHostConfigs(functionIDs []string, opts [][]string) ([]*forwarder.HostConfig, error) {
	cfgs := make([]*forwarder.HostConfig, len(functionIDs))
	for i, id := range functionIDs {
		hostConfig := &forwarder.HostConfig{}
		uri, err := url.Parse(id)
		if err != nil {
			return nil, err
		}
		switch uri.Scheme {
		case "":
			{
				hostConfig.Transport = forwarder.GRPC
				hostConfig.Host = uri.Path
			}
		case "grpc":
			{
				hostConfig.Transport = forwarder.GRPC
				hostConfig.Host = uri.Host
			}
		case "http":
			{
				hostConfig.Transport = forwarder.HTTP
				hostConfig.Host = uri.Host
			}
		default:
			{
				return nil, fmt.Errorf("no such transport: %v uri: %v", uri.Scheme, id)
			}
		}
		if port := uri.Port(); port != "" {
			portNum, err := strconv.ParseUint(port, 10, 64)
			if err != nil {
				return nil, err
			}
			hostConfig.Port = uint16(portNum)
		} else {
			hostConfig.Port = s.defaultPort
			if hostConfig.Transport == forwarder.HTTP {
				hostConfig.Port = 8080
			}
		}
		hostConfig.CallOptions = opts[i]
		cfgs[i] = hostConfig
	}
	return cfgs, nil
}
