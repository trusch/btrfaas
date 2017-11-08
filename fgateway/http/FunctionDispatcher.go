package handler

import (
	"context"
	"net/http"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/trusch/btrfaas/fgateway/forwarder"
)

// FunctionDispatcher is an HTTP handler which dispatch function calls
// accepts something like this: /api/v0/invoke/<my-function-id>
type FunctionDispatcher struct {
	DefaultPort uint16
}

// NewFunctionDispatcher returns a new http handler
func NewFunctionDispatcher(defaultPort uint16) http.Handler {
	return &FunctionDispatcher{defaultPort}
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
		err := forwarder.Forward(ctx, &forwarder.Options{
			Hosts: []*forwarder.HostConfig{
				&forwarder.HostConfig{
					Transport: forwarder.GRPC,
					Host:      functionID,
					Port:      d.DefaultPort,
				},
			},
			Input:  r.Body,
			Output: w,
		})
		if err != nil {
			log.Errorf("error forwarding function call: %v", err)
		}
		log.Info("finished request")
		return
	}

	log.Warn("unknown request path")
	w.WriteHeader(http.StatusNotFound)
}
