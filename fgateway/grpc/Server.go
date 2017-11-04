package grpc

import (
	"io"
	"net"

	log "github.com/sirupsen/logrus"

	"github.com/trusch/btrfaas/fgateway/forwarder"

	"google.golang.org/grpc"
)

type Server struct {
	addr        string
	defaultPort uint16
	grpcOpts    []grpc.ServerOption
}

func NewServer(addr string, defaultPort uint16, opts ...grpc.ServerOption) *Server {
	return &Server{addr, defaultPort, opts}
}

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

func (s *Server) Run(stream FunctionRunner_RunServer) (err error) {
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
		log.Info("forward to function service ", firstPacket.FunctionID)
		err := forwarder.Forward(ctx, &forwarder.Options{
			Transport: forwarder.GRPC,
			Host:      firstPacket.FunctionID,
			Port:      s.defaultPort,
			Input:     input,
			Output:    outputWriter,
		})
		if err != nil {
			log.Errorf("error forwarding function call: %v", err)
		}
		close(done)
	}()
	go s.shovelInputData(stream, inputWriter)
	go s.shovelOutputData(stream, outputReader)
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

func (s *Server) shovelInputData(stream FunctionRunner_RunServer, input io.WriteCloser) error {
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
						input.Close()
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
					return stream.Send(&GWOutputData{Output: buf[:bs]})
				}
				if err != nil {
					return err
				}
				if err = stream.Send(&GWOutputData{Output: buf[:bs]}); err != nil {
					return err
				}
			}
		}
	}
}
