package main

import (
	"errors"
	"log"
	"os"
	"strings"

	"github.com/trusch/frunner/config"
	"github.com/trusch/frunner/env"
	"github.com/trusch/frunner/framer"
	"github.com/trusch/frunner/http"
	"github.com/trusch/frunner/runnable"
	"github.com/trusch/frunner/runnable/afterburn"
	"github.com/trusch/frunner/runnable/exec"
)

var (
	binary     string
	binaryArgs []string
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}
	if err = getBinaryAndArgs(); err != nil {
		log.Fatal(err)
	}
	log.Printf("Listen Address: %v", *cfg.HTTPAddr)
	log.Printf("Read Limit: %v\n", *cfg.ReadLimit)
	log.Printf("Command: %v\n", binary)
	log.Printf("Arguments: %v\n", binaryArgs)
	log.Printf("Call timeout: %v\n", *cfg.CallTimeout)
	log.Printf("Buffer: %v\n", *cfg.Buffer)

	var cmd runnable.Runnable
	switch *cfg.Framer {
	case "":
		{
			cmd = exec.NewRunnable(binary, binaryArgs...)
			if *cfg.Buffer {
				cmd.(*exec.Runnable).EnableOutputBuffering()
			}
		}
	case "line":
		cmd = afterburn.NewRunnable(&framer.LineFramer{}, binary, binaryArgs...)
	case "json":
		cmd = afterburn.NewRunnable(&framer.JSONFramer{}, binary, binaryArgs...)
	case "http":
		cmd = afterburn.NewRunnable(&framer.HTTPFramer{}, binary, binaryArgs...)
	}
	server := http.NewServer(cmd, cfg)
	log.Print("start listening for requests...")
	log.Fatal(server.ListenAndServe())
}

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
		validProcessKeys := []string{
			"FRUNNER_PROCESS",
			"FRUNNER_CMD",
			"FAAS_CMD",
			"FPROCESS",
			"fprocess",
			"faas_cmd",
			"fwatchdog_cmd",
			"fwatch_cmd",
		}
		for _, key := range validProcessKeys {
			if val, ok := env[key]; ok {
				parts := strings.Split(val, " ")
				binary = parts[0]
				if len(parts) > 1 {
					binaryArgs = parts[1:]
				}
				break
			}
		}
	}

	if binary == "" {
		return errors.New("can not determine process to execute")
	}

	return nil
}
