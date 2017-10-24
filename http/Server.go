package http

import (
	"io"
	"net/http"
	"time"

	"github.com/trusch/frunner/callable"
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
func NewServer(cmd callable.Callable, env env.Env, addr string, readTimeout, writeTimeout time.Duration, readLimit int64) *Server {
	srv := &http.Server{
		Addr:           addr,
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: 1 << 20, // Max header of 1MB
	}
	server := &Server{srv, cmd, env, readLimit}
	server.srv.Handler = server
	return server
}

func (server *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// clean copy of callable
	cmd := server.cmd.Copy()

	// prepare environment
	env := server.env.Copy()
	env.AddFromHTTPRequest(r)

	// call the callable
	errorChannel := cmd.Call(io.LimitReader(r.Body, server.readLimit), env, w)
	if err := <-errorChannel; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
}

// ListenAndServe starts the HTTP server
func (server *Server) ListenAndServe() error {
	return server.srv.ListenAndServe()
}
