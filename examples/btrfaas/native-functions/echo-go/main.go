package main

import (
	"log"

	"github.com/trusch/btrfaas/frunner/config"
	"github.com/trusch/btrfaas/frunner/grpc"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	cmd := &Runnable{}

	go func() {
		grpcServer := grpc.NewServer(cmd, cfg)
		log.Fatal(grpcServer.ListenAndServe())
	}()

	select {}
}
