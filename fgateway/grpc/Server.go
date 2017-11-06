package grpc

import (
	"io"
	"net"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/trusch/btrfaas/fgateway/forwarder"

	"google.golang.org/grpc"
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
	RegisterFunctionRunnerServer(grpcServer, s)
	return grpcServer.Serve(lis)
}

// Run implements the gRPC interface
func (s *Server) Run(stream FunctionRunner_RunServer) (err error) {
	defer func() {
		if err != nil {
			log.Error(err)
			return
		}
		log.Info("finished gRPC request successfully")
	}()
	log.Info("new gRPC request")
	ctx := stream.Context()
	inputReader, inputWriter := io.Pipe()
	outputReader, outputWriter := io.Pipe()
	defer inputWriter.Close()
	defer outputWriter.Close()
	done := make(chan struct{})

	var input io.Reader = inputReader

	firstPacket, err := stream.Recv()
	if err != nil {
		return err
	}
	go func() {
		log.Info("forward to function service ", firstPacket.Chain)
		err = forwarder.Forward(ctx, &forwarder.Options{
			Hosts:  s.createHostConfigs(strings.Split(firstPacket.Chain, "|"), firstPacket.Options),
			Input:  input,
			Output: outputWriter,
		})
		if err != nil {
			log.Errorf("error forwarding function call: %v", err)
		}
		close(done)
	}()
	go func() { err = s.shovelInputData(stream, inputWriter) }()
	go func() { err = s.shovelOutputData(stream, outputReader) }()
	select {
	case <-done:
		{
			return
		}
	case <-ctx.Done():
		{
			return ctx.Err()
		}
	}
}

func (s *Server) createHostConfigs(functionIDs []string, opts map[string]*FunctionOptions) []*forwarder.HostConfig {
	cfgs := make([]*forwarder.HostConfig, len(functionIDs))
	for i, id := range functionIDs {
		cfgs[i] = &forwarder.HostConfig{
			Transport:   forwarder.GRPC,
			Host:        strings.Trim(id, " \t"),
			Port:        s.defaultPort,
			CallOptions: opts[id].Options,
		}
	}
	return cfgs
}

func (s *Server) shovelInputData(stream FunctionRunner_RunServer, input io.WriteCloser) error {
	defer input.Close()
	ctx := stream.Context()
	for {
		select {
		case <-ctx.Done():
			{
				return ctx.Err()
			}
		default:
			{
				data, err := stream.Recv()
				if err != nil {
					if err == io.EOF {
						return nil
					}
					return err
				}
				if _, err = input.Write(data.Data); err != nil {
					return err
				}
			}
		}
	}
}

func (s *Server) shovelOutputData(stream FunctionRunner_RunServer, output io.Reader) error {
	ctx := stream.Context()
	buf := make([]byte, 4096)
	for {
		select {
		case <-ctx.Done():
			{
				return ctx.Err()
			}
		default:
			{
				bs, err := output.Read(buf[:])
				if err == io.EOF {
					return stream.Send(&FgatewayOutputData{Output: buf[:bs]})
				}
				if err != nil {
					return err
				}
				if err = stream.Send(&FgatewayOutputData{Output: buf[:bs]}); err != nil {
					return err
				}
			}
		}
	}
}
