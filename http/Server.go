package http

import (
	"io"
	"log"
	"net/http"

	"github.com/trusch/frunner/callable"
	"github.com/trusch/frunner/config"
	"github.com/trusch/frunner/env"
)

// Server serves HTTP requests and calls the given callable
type Server struct {
	srv       *http.Server
	cmd       callable.Callable
	env       env.Env
	readLimit int64
}

// NewServer creates a new HTTP server for a given Callable
func NewServer(cmd callable.Callable, cfg *config.Config) *Server {
	srv := &http.Server{
		Addr:           *cfg.HTTPAddr,
		ReadTimeout:    *cfg.HTTPReadTimeout,
		WriteTimeout:   *cfg.HTTPWriteTimeout,
		MaxHeaderBytes: 1 << 20, // Max header of 1MB
	}
	server := &Server{srv, cmd, make(env.Env), *cfg.ReadLimit}
	server.srv.Handler = server
	if *cfg.CallTimeout > 0 {
		server.cmd = callable.NewTimeoutCallable(cmd, *cfg.CallTimeout)
	}
	if err := server.env.ReadOSEnvironment(); err != nil {
		log.Fatal(err)
	}
	return server
}

func (server *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// clean copy of callable
	cmd := server.cmd.Copy()

	// prepare environment
	env := server.env.Copy()
	env.AddFromHTTPRequest(r)

	// call the callable
	var reader io.Reader
	reader = r.Body
	if server.readLimit > 0 {
		reader = io.LimitReader(reader, server.readLimit)
	}
	errorChannel := cmd.Call(reader, env, w)
	if err := <-errorChannel; err != nil {
		log.Print("error while calling: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		_, e := w.Write([]byte(err.Error()))
		if e != nil {
			log.Print("failed to write error to client: ", e)
		}
	}
}

// ListenAndServe starts the HTTP server
func (server *Server) ListenAndServe() error {
	return server.srv.ListenAndServe()
}
