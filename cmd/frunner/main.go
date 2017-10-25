package main

import (
	"errors"
	"log"
	"os"
	"strings"

	"github.com/trusch/frunner/callable"
	"github.com/trusch/frunner/config"
	"github.com/trusch/frunner/env"
	"github.com/trusch/frunner/framer"
	"github.com/trusch/frunner/http"
)

var (
	binary     string
	binaryArgs []string
)

func getBinaryAndArgs() error {
	// check if "--" is in argument list -> everything after that is interpreted as command
	dashDashIndex := -1
	for idx, val := range os.Args {
		if val == "--" {
			dashDashIndex = idx
			break
		}
	}
	args := os.Args
	rest := []string{}
	if dashDashIndex != -1 {
		rest = args[dashDashIndex+1:]
		args = args[:dashDashIndex]
	}
	if len(rest) > 0 {
		binary = rest[0]
		binaryArgs = rest[1:]
	}

	if binary == "" {
		env := make(env.Env)
		if err := env.ReadOSEnvironment(); err != nil {
			return err
		}
		log.Print(env)
		if val, ok := env["fprocess"]; ok {
			parts := strings.Split(val, " ")
			binary = parts[0]
			if len(parts) > 1 {
				binaryArgs = parts[1:]
			}
		}
	}

	if binary == "" {
		return errors.New("can not determine process to execute")
	}

	return nil
}

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}
	if err := getBinaryAndArgs(); err != nil {
		log.Fatal(err)
	}
	log.Printf("Listen Address: %v", *cfg.HTTPAddr)
	log.Printf("Mode: %v\n", *cfg.Mode)
	log.Printf("Timeouts:\n\tCall: %v\n\tRead: %v\n\tWrite: %v\n", *cfg.CallTimeout, *cfg.HTTPReadTimeout, *cfg.HTTPWriteTimeout)
	log.Printf("Read Limit: %v\n", *cfg.ReadLimit)
	log.Printf("Command: %v\n", binary)
	log.Printf("Arguments: %v\n", binaryArgs)
	var cmd callable.Callable
	switch *cfg.Mode {
	case "pipe":
		cmd = callable.NewPipingExecCallable(binary, binaryArgs...)
	case "buffer":
		cmd = callable.NewBufferingExecCallable(binary, binaryArgs...)
	case "afterburn":
		{
			switch *cfg.Framer {
			case "line":
				cmd = callable.NewAfterburnExecCallable(&framer.LineFramer{}, binary, binaryArgs...)
			case "json":
				cmd = callable.NewAfterburnExecCallable(&framer.JSONFramer{}, binary, binaryArgs...)
			case "http":
				cmd = callable.NewAfterburnExecCallable(&framer.HTTPFramer{}, binary, binaryArgs...)
			}
		}
	}
	if *cfg.Mode == "pipe" {
	} else if *cfg.Mode == "buffer" {
	}
	server := http.NewServer(cmd, cfg)
	log.Print("start listening for requests...")
	log.Fatal(server.ListenAndServe())
}
