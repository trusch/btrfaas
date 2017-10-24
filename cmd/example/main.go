package main

import (
	"log"
	"os"
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
)

func initFlags() {
	flags := pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
	httpAddr = flags.StringP("http-addr", "l", ":8080", "http listen address")
	httpReadTimeout = flags.DurationP("http-read-timeout", "r", 5*time.Second, "http read timeout")
	httpWriteTimeout = flags.DurationP("http-write-timeout", "w", 5*time.Second, "http write timeout")
	callTimeout = flags.DurationP("call-timeout", "t", 5*time.Second, "call timeout")
}

func applyEnvironmentConfig() {
	env := make(env.Env)
	if err := env.ReadOSEnvironment(); err != nil {
		log.Fatal(err)
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
}

func main() {
	initFlags()
	applyEnvironmentConfig()

	cmd := &callable.ExampleCallable{}
	tCmd := callable.NewTimeoutCallable(cmd, *callTimeout)
	env := make(env.Env)
	if err := env.ReadOSEnvironment(); err != nil {
		log.Fatal(err)
	}
	server := http.NewServer(tCmd, env, *httpAddr, *httpReadTimeout, *httpWriteTimeout)
	log.Fatal(server.ListenAndServe())
}
