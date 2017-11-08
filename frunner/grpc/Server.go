package grpc

import (
	"context"
	"io"
	"net"

	log "github.com/Sirupsen/logrus"

	"google.golang.org/grpc"

	"github.com/trusch/btrfaas/frunner/config"
	"github.com/trusch/btrfaas/frunner/runnable"
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
	done := make(chan error)

	var input io.Reader = inputReader
	if *s.cfg.ReadLimit > 0 {
		input = io.LimitReader(input, *s.cfg.ReadLimit)
	}

	firstPacket, err := stream.Recv()
	if err != nil {
		return err
	}
	options := firstPacket.GetOptions()
	go func() {
		defer outputWriter.Close()
		err = s.cmd.Run(ctx, options, input, outputWriter)
		if err != nil {
			log.Error("error executing runnable: ", err)
			done <- err
		}
	}()
	if len(firstPacket.Data) > 0 {
		if _, err = inputWriter.Write(firstPacket.Data); err != nil {
			return err
		}
	}

	go s.shovelInputData(stream, inputWriter)
	go func() {
		defer close(done)
		e := s.shovelOutputData(stream, outputReader)
		if e != nil {
			done <- e
			return
		}
	}()
	select {
	case err, ok := <-done:
		{
			if ok && err != nil {
				return stream.Send(&FrunnerOutputData{
					Ready:        true,
					Success:      false,
					ErrorMessage: err.Error(),
				})
			}
			return stream.Send(&FrunnerOutputData{
				Ready:   true,
				Success: true,
			})
		}
	case <-ctx.Done():
		{
			return ctx.Err()
		}
	}
}

func (s *Server) shovelInputData(stream FunctionRunner_RunServer, input io.WriteCloser) error {
	ctx := stream.Context()
	defer input.Close()
	defer log.Print("frunner finished shoveling input data")
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
	defer log.Print("frunner finished shoveling output data")
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
					return stream.Send(&FrunnerOutputData{Output: buf[:bs]})
				}
				if err != nil {
					return err
				}
				if err = stream.Send(&FrunnerOutputData{Output: buf[:bs]}); err != nil {
					return err
				}
			}
		}
	}
}
