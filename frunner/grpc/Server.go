package grpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"

	log "github.com/Sirupsen/logrus"

	btrfaasgrpc "github.com/trusch/btrfaas/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	"github.com/trusch/btrfaas/frunner/config"
	"github.com/trusch/btrfaas/frunner/runnable"
)

// Server is a gRPC server which serves function calls to a specific runnable
type Server struct {
	cmd      runnable.Runnable
	cfg      *config.Config
	grpcOpts []grpc.ServerOption
}

// NewServer returns a new server instance
func NewServer(cmd runnable.Runnable, cfg *config.Config, opts ...grpc.ServerOption) *Server {
	return &Server{cmd, cfg, opts}
}

// ListenAndServe start listening for requests
func (s *Server) ListenAndServe() error {
	lis, err := net.Listen("tcp", *s.cfg.GRPCAddr)
	if err != nil {
		return err
	}
	certificate, err := tls.LoadX509KeyPair("/run/secrets/btrfaas-function-cert.pem", "/run/secrets/btrfaas-function-key.pem")
	if err != nil {
		return fmt.Errorf("could not load server key pair: %s", err)
	}

	// Create a certificate pool from the certificate authority
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile("/run/secrets/btrfaas-ca-cert.pem")
	if err != nil {
		return fmt.Errorf("could not read ca certificate: %s", err)
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

// Run implements the server interface implied by the btrfaas protobuf service definition
func (s *Server) Run(stream btrfaasgrpc.FunctionRunner_RunServer) error {
	ctx := stream.Context()
	if *s.cfg.CallTimeout > 0 {
		log.Print("set timeout of ", *s.cfg.CallTimeout, " to context")
		c, cancel := context.WithTimeout(ctx, *s.cfg.CallTimeout)
		defer cancel()
		ctx = c
	}

	options := getOptionsFromStream(stream)

	inputReader, inputWriter := io.Pipe()
	outputReader, outputWriter := io.Pipe()

	var input io.Reader = inputReader
	if *s.cfg.ReadLimit > 0 {
		input = io.LimitReader(input, *s.cfg.ReadLimit)
	}

	done := make(chan error, 5)

	go func() {
		done <- s.cmd.Run(ctx, options, input, outputWriter)
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

func (s *Server) shovelInputData(stream btrfaasgrpc.FunctionRunner_RunServer, input io.WriteCloser) error {
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

func (s *Server) shovelOutputData(stream btrfaasgrpc.FunctionRunner_RunServer, output io.Reader) error {
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
					return stream.Send(&btrfaasgrpc.Data{Data: buf[:bs]})
				}
				if err != nil {
					return err
				}
				if err = stream.Send(&btrfaasgrpc.Data{Data: buf[:bs]}); err != nil {
					return err
				}
			}
		}
	}
}

func getOptionsFromStream(stream btrfaasgrpc.FunctionRunner_RunServer) []string {
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return nil
	}
	optionsList, ok := md["options"]
	if !ok {
		return nil
	}
	return optionsList
}
