package main

import (
	"log"

	"github.com/trusch/frunner/callable"
	"github.com/trusch/frunner/config"
	"github.com/trusch/frunner/http"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	cmd := &callable.ExampleCallable{}

	server := http.NewServer(cmd, cfg)
	log.Fatal(server.ListenAndServe())
}
