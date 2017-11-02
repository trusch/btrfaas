package handler

import (
	"context"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/trusch/btrfaas/frunner/grpc"
	g "google.golang.org/grpc"
)

// FunctionDispatcher is an HTTP handler which dispatch function calls
// accepts something like this: /api/v0/invoke/<my-function-id>
type FunctionDispatcher struct{}

// NewFunctionDispatcher returns a new http handler
func NewFunctionDispatcher() http.Handler {
	return &FunctionDispatcher{}
}

func (d *FunctionDispatcher) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Info("request: %v %v (%v)", r.Method, r.URL, r.RemoteAddr)
	path := r.URL.Path
	if !strings.HasPrefix(path, "/api/v0") {
		log.Warn("unknown request path")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if strings.HasPrefix(path, "/api/v0/invoke") {
		parts := strings.Split(path, "/")
		if len(parts) != 5 {
			log.Warn("malformed invoke request: ", path)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		functionID := parts[4]
		function, err := grpc.NewClient(functionID+":2424", g.WithInsecure())
		if err != nil {
			log.Errorf("connecting function service %v: %v", functionID, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		ctx := context.Background()
		if timeout := r.URL.Query().Get("timeout"); timeout != "" {
			t, e := time.ParseDuration(timeout)
			if e != nil {
				log.Errorf("parsing timeout: %v", e)
			}
			c, cancel := context.WithTimeout(ctx, t)
			defer cancel()
			ctx = c
		}

		if err := function.Run(ctx, r.Body, w); err != nil {
			log.Errorf("error running function: %v", err)
		}
		log.Info("finished request")
		return
	}

	log.Warn("unknown request path")
	w.WriteHeader(http.StatusNotFound)
}
