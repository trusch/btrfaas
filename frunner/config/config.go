package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	yaml "gopkg.in/yaml.v2"

	"github.com/spf13/pflag"
	"github.com/trusch/btrfaas/frunner/env"
)

// Config contains the common config for frunner
type Config struct {
	glags                 *pflag.FlagSet
	HTTPAddr              *string
	GRPCAddr              *string
	HTTPReadHeaderTimeout *time.Duration
	CallTimeout           *time.Duration
	ReadLimit             *int64
	Framer                *string
	Buffer                *bool
}

// New creates a new config object
func New() (*Config, error) {
	flags := pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
	cfg := &Config{
		glags:                 flags,
		HTTPAddr:              flags.StringP("http-addr", "l", ":8080", "http listen address"),
		GRPCAddr:              flags.StringP("grpc-addr", "g", ":2424", "grpc listen address"),
		HTTPReadHeaderTimeout: flags.DurationP("http-timeout", "h", 1*time.Second, "http timeout for reading request headers"),
		CallTimeout:           flags.DurationP("call-timeout", "t", 0*time.Second, "function call timeout"),
		ReadLimit:             flags.Int64("read-limit", -1, "limit the amount of data which can be contained in a requests body"),
		Framer:                flags.StringP("framer", "f", "", "afterburn framer to use: line, json or http"),
		Buffer:                flags.BoolP("buffer", "b", false, "buffer output before writing"),
	}
	if err := cfg.parseCommandline(); err != nil {
		return nil, err
	}
	if err := cfg.parseEnvironment(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// parseCommandline parses os.Args and fills the config entries
func (cfg *Config) parseCommandline() error {
	return cfg.glags.Parse(stripEverythingAfterDoubleDash(os.Args))
}

// parseEnvironment parses the environment for config entries
func (cfg *Config) parseEnvironment() error {
	env := make(env.Env)
	if err := env.ReadOSEnvironment(); err != nil {
		return err
	}
	if val, ok := env["FRUNNER_CALL_TIMEOUT"]; ok {
		d, err := time.ParseDuration(val)
		if err != nil {
			return err
		}
		cfg.CallTimeout = &d
	}
	if val, ok := env["FRUNNER_HTTP_TIMEOUT"]; ok {
		d, err := time.ParseDuration(val)
		if err != nil {
			return err
		}
		cfg.HTTPReadHeaderTimeout = &d
	}
	if val, ok := env["FRUNNER_READ_LIMIT"]; ok {
		d, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		cfg.ReadLimit = &d
	}
	if val, ok := env["FRUNNER_FRAMER"]; ok {
		cfg.Framer = &val
	}
	if val, ok := env["FRUNNER_HTTP_ADDRESS"]; ok {
		cfg.HTTPAddr = &val
	}
	if val, ok := env["FRUNNER_GRPC_ADDRESS"]; ok {
		cfg.GRPCAddr = &val
	}
	if _, ok := env["FRUNNER_BUFFER"]; ok {
		v := true
		cfg.Buffer = &v
	}
	return nil
}

func stripEverythingAfterDoubleDash(args []string) []string {
	idx := -1
	for i, val := range args {
		if val == "--" {
			idx = i
			break
		}
	}
	if idx != -1 {
		return args[:idx]
	}
	return args
}

func (cfg *Config) Print() {
	bs, _ := yaml.Marshal(cfg)
	fmt.Println(string(bs))
}
