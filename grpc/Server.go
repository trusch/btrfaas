package grpc

import (
	"context"
	"io"
	"log"
	"net"

	"google.golang.org/grpc"

	"github.com/trusch/frunner/config"
	"github.com/trusch/frunner/runnable"
)

type Server struct {
	cmd      runnable.Runnable
	cfg      *config.Config
	grpcOpts []grpc.ServerOption
}

func NewServer(cmd runnable.Runnable, cfg *config.Config, opts ...grpc.ServerOption) *Server {
	return &Server{cmd, cfg, opts}
}

func (s *Server) ListenAndServe() error {
	lis, err := net.Listen("tcp", *s.cfg.GRPCAddr)
	if err != nil {
		return err
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	RegisterFunctionRunnerServer(grpcServer, s)
	return grpcServer.Serve(lis)
}

func (s *Server) Run(stream FunctionRunner_RunServer) (err error) {
	ctx := stream.Context()
	if *s.cfg.CallTimeout > 0 {
		log.Print("set timeout of ", *s.cfg.CallTimeout, " to context")
		c, cancel := context.WithTimeout(ctx, *s.cfg.CallTimeout)
		defer cancel()
		ctx = c
	}
	inputReader, inputWriter := io.Pipe()
	outputReader, outputWriter := io.Pipe()
	defer inputWriter.Close()
	defer outputWriter.Close()
	done := make(chan struct{})

	var input io.Reader = inputReader
	if *s.cfg.ReadLimit > 0 {
		input = io.LimitReader(input, *s.cfg.ReadLimit)
	}

	go func() {
		err = s.cmd.Run(ctx, input, outputWriter)
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
					return stream.Send(&OutputData{Output: buf[:bs]})
				}
				if err != nil {
					return err
				}
				if err = stream.Send(&OutputData{Output: buf[:bs]}); err != nil {
					return err
				}
			}
		}
	}
}
