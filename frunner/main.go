package main

import (
	"errors"
	"log"
	"os"
	"strings"
	"time"

	"github.com/trusch/btrfaas/frunner/config"
	"github.com/trusch/btrfaas/frunner/env"
	"github.com/trusch/btrfaas/frunner/grpc"
	"github.com/trusch/btrfaas/frunner/http"
	"github.com/trusch/btrfaas/frunner/runnable/exec"
	g "google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
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

	cfg.Print()

	cmd := exec.NewRunnable(binary, binaryArgs...)
	if *cfg.Buffer {
		cmd.EnableOutputBuffering()
	}

	httpServer := http.NewServer(cmd, cfg)
	log.Print("start listening for requests via http on ", *cfg.HTTPAddr)
	go func() {
		log.Fatal(httpServer.ListenAndServe())
	}()

	grpcServer := grpc.NewServer(cmd, cfg, g.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionAge:      2 * time.Minute,
		MaxConnectionAgeGrace: 10 * time.Second,
	}))
	log.Print("start listening for requests via grpc on ", *cfg.GRPCAddr)
	go func() {
		log.Fatal(grpcServer.ListenAndServe())
	}()
	select {}
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
