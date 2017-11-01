package main

import (
	"log"

	"github.com/trusch/frunner/config"
	"github.com/trusch/frunner/grpc"
	"github.com/trusch/frunner/http"
	"github.com/trusch/frunner/runnable/example"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	cmd := &example.Runnable{}

	go func() {
		httpServer := http.NewServer(cmd, cfg)
		log.Fatal(httpServer.ListenAndServe())
	}()

	go func() {
		grpcServer := grpc.NewServer(cmd, cfg)
		log.Fatal(grpcServer.ListenAndServe())
	}()

	select {}
}
