package http

import (
	"context"
	"io"
	"log"
	"net/http"

	"github.com/trusch/btrfaas/frunner/config"
	"github.com/trusch/btrfaas/frunner/env"
	"github.com/trusch/btrfaas/frunner/runnable"
)

// Server serves HTTP requests and calls the given callable
type Server struct {
	srv *http.Server
	cmd runnable.Runnable
	env env.Env
	cfg *config.Config
}

// NewServer creates a new HTTP server for a given Callable
func NewServer(cmd runnable.Runnable, cfg *config.Config) *Server {
	srv := &http.Server{
		Addr:              *cfg.HTTPAddr,
		ReadHeaderTimeout: *cfg.HTTPReadHeaderTimeout,
		MaxHeaderBytes:    1 << 20, // Max header of 1MB
	}
	server := &Server{srv, cmd, make(env.Env), cfg}
	server.srv.Handler = server
	if err := server.env.ReadOSEnvironment(); err != nil {
		log.Fatal(err)
	}
	return server
}

func (server *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// prepare environment
	environment := server.env.Copy()
	environment.AddFromHTTPRequest(r)

	// prepare input data
	var input io.Reader = r.Body
	if *server.cfg.ReadLimit > 0 {
		input = io.LimitReader(input, *server.cfg.ReadLimit)
	}

	// create context
	ctx := context.Background()
	ctx = env.NewContext(ctx, environment)
	if *server.cfg.CallTimeout > 0 {
		c, cancel := context.WithTimeout(ctx, *server.cfg.CallTimeout)
		ctx = c
		defer cancel()
	}

	// call the function
	err := server.cmd.Run(ctx, nil, input, w)
	if err != nil {
		log.Print("error while calling: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
}

// ListenAndServe starts the HTTP server
func (server *Server) ListenAndServe() error {
	return server.srv.ListenAndServe()
}
