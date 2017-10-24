package main

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/trusch/frunner/callable"
	"github.com/trusch/frunner/env"
	"github.com/trusch/frunner/http"
)

var (
	httpAddr         *string
	callTimeout      *time.Duration
	httpReadTimeout  *time.Duration
	httpWriteTimeout *time.Duration
	readLimit        *int64
	binary           string
	binaryArgs       []string
)

func initFlags() {
	flags := pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
	httpAddr = flags.StringP("http-addr", "l", ":8080", "http listen address")
	httpReadTimeout = flags.DurationP("http-read-timeout", "r", 5*time.Second, "http read timeout")
	httpWriteTimeout = flags.DurationP("http-write-timeout", "w", 5*time.Second, "http write timeout")
	callTimeout = flags.DurationP("call-timeout", "t", 5*time.Second, "call timeout")
	readLimit = flags.Int64("read-limit", 1048576, "read limit")
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
	if err := flags.Parse(args); err != nil {
		log.Fatal(err)
	}
	if len(rest) > 0 {
		binary = rest[0]
		binaryArgs = rest[1:]
	}
}

func applyEnvironmentConfig() {
	env := make(env.Env)
	if err := env.ReadOSEnvironment(); err != nil {
		log.Fatal(err)
	}
	if val, ok := env["fprocess"]; ok {
		parts := strings.Split(val, " ")
		binary = parts[0]
		if len(parts) > 1 {
			binaryArgs = parts[1:]
		}
	}
	if val, ok := env["read_timeout"]; ok {
		d, err := time.ParseDuration(val)
		if err != nil {
			log.Fatal(err)
		}
		httpReadTimeout = &d
	}
	if val, ok := env["write_timeout"]; ok {
		d, err := time.ParseDuration(val)
		if err != nil {
			log.Fatal(err)
		}
		httpWriteTimeout = &d
	}
	if val, ok := env["call_timeout"]; ok {
		d, err := time.ParseDuration(val)
		if err != nil {
			log.Fatal(err)
		}
		callTimeout = &d
	}
	if val, ok := env["read_limit"]; ok {
		d, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		readLimit = &d
	}
}

func main() {
	initFlags()
	applyEnvironmentConfig()
	cmd := callable.NewExecCallable(binary, binaryArgs...)
	tCmd := callable.NewTimeoutCallable(cmd, *callTimeout)
	env := make(env.Env)
	if err := env.ReadOSEnvironment(); err != nil {
		log.Fatal(err)
	}
	server := http.NewServer(tCmd, env, *httpAddr, *httpReadTimeout, *httpWriteTimeout, *readLimit)
	log.Fatal(server.ListenAndServe())
}
