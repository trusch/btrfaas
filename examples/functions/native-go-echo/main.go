package main

import (
	"log"

	"github.com/trusch/btrfaas/frunner/config"
	"github.com/trusch/btrfaas/frunner/grpc"
	"github.com/trusch/btrfaas/frunner/http"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	cmd := &Runnable{}

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
