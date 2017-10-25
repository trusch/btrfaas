package config

import (
	"os"
	"strconv"
	"time"

	"github.com/spf13/pflag"
	"github.com/trusch/frunner/env"
)

// Config contains the common config for frunner
type Config struct {
	Flags            *pflag.FlagSet
	HTTPAddr         *string
	CallTimeout      *time.Duration
	ReadLimit        *int64
	HTTPReadTimeout  *time.Duration
	HTTPWriteTimeout *time.Duration
	Mode             *string
	Framer           *string
}

// New creates a new config object
func New() (*Config, error) {
	flags := pflag.NewFlagSet(os.Args[0], pflag.ExitOnError)
	cfg := &Config{
		Flags:            flags,
		HTTPAddr:         flags.StringP("http-addr", "l", ":8080", "http listen address"),
		HTTPReadTimeout:  flags.DurationP("http-read-timeout", "r", 5*time.Second, "http read timeout"),
		HTTPWriteTimeout: flags.DurationP("http-write-timeout", "w", 5*time.Second, "http write timeout"),
		CallTimeout:      flags.DurationP("call-timeout", "t", 5*time.Second, "call timeout"),
		ReadLimit:        flags.Int64("read-limit", 1048576, "read limit"),
		Mode:             flags.StringP("mode", "m", "buffer", "operation mode: buffer, pipe or afterburn"),
		Framer:           flags.StringP("framer", "f", "http", "framer to use: line, json or http"),
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
	return cfg.Flags.Parse(stripEverythingAfterDoubleDash(os.Args))
}

// parseEnvironment parses the environment for config entries
func (cfg *Config) parseEnvironment() error {
	env := make(env.Env)
	if err := env.ReadOSEnvironment(); err != nil {
		return err
	}
	if val, ok := env["read_timeout"]; ok {
		d, err := time.ParseDuration(val)
		if err != nil {
			return err
		}
		cfg.HTTPReadTimeout = &d
	}
	if val, ok := env["write_timeout"]; ok {
		d, err := time.ParseDuration(val)
		if err != nil {
			return err
		}
		cfg.HTTPWriteTimeout = &d
	}
	if val, ok := env["call_timeout"]; ok {
		d, err := time.ParseDuration(val)
		if err != nil {
			return err
		}
		cfg.CallTimeout = &d
	}
	if val, ok := env["read_limit"]; ok {
		d, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		cfg.ReadLimit = &d
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
