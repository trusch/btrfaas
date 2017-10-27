package main

import (
	"log"

	"github.com/trusch/frunner/config"
	"github.com/trusch/frunner/http"
	"github.com/trusch/frunner/runnable/example"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	cmd := &example.Runnable{}

	server := http.NewServer(cmd, cfg)
	log.Fatal(server.ListenAndServe())
}
