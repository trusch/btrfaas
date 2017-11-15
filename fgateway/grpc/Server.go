package grpc

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/trusch/btrfaas/fgateway/forwarder"
	"github.com/trusch/btrfaas/fgateway/metrics"
	btrfaasgrpc "github.com/trusch/btrfaas/grpc"

	"google.golang.org/grpc"
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
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
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
	hosts := s.createHostConfigs(chain, options)

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

func getOptionsFromStream(stream btrfaasgrpc.FunctionRunner_RunServer) (chain []string, optionSlice []map[string]string, err error) {
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return nil, nil, errors.New("no metadata")
	}
	log.Print(md)
	chain, ok = md["chain"]
	if !ok {
		return nil, nil, errors.New("no chain in metadata")
	}
	optionsList, ok := md["options"]
	if !ok {
		return chain, nil, nil
	}
	optionSlice = make([]map[string]string, len(optionsList))
	for idx, objStr := range optionsList {
		options := make(map[string]string)
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

func (s *Server) createHostConfigs(functionIDs []string, opts []map[string]string) []*forwarder.HostConfig {
	cfgs := make([]*forwarder.HostConfig, len(functionIDs))
	for i, id := range functionIDs {
		cfgs[i] = &forwarder.HostConfig{
			Transport:   forwarder.GRPC,
			Host:        strings.Trim(id, " \t"),
			Port:        s.defaultPort,
			CallOptions: opts[i],
		}
	}
	return cfgs
}
