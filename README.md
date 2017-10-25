frunner
=======

## OpenFaaS

This is intended to be used with openfaas and can hopefully be used as a replacement for the fwatchdog.
This is heavily inspired by openfaas/faas/watchdog and all credits for this are going to @alexellis. Thank you for your fantastic work :)

## Scope

This provides an HTTP server which executes the configured binary with the input from the request and returns the result.
Data is passed around via io.Reader/io.Writer so that data sized > available memory is possible.

Additionally a interface is provided to enable developers to create customized versions of the frunner to prevent forking a process for every call.

## Configuration
```bash
frunner --help
Usage of frunner:
  -t, --call-timeout duration         call timeout (default 5s)
  -f, --framer string                 framer to use: line, json or http (default "http")
  -l, --http-addr string              http listen address (default ":8080")
  -r, --http-read-timeout duration    http read timeout (default 5s)
  -w, --http-write-timeout duration   http write timeout (default 5s)
  -m, --mode string                   operation mode: buffer, pipe or afterburn (default "buffer")
      --read-limit int                read limit (default 1048576)
```

A typical call would look like this:
`frunner --http-addr :8080 -- cat -`

This would result in an HTTP echo server ("cat -" just rewrites stdin to stdout)
Everything after the "--" is interpreted as the executable and its arguments

You can also configure the `frunner` via environment variables:
```bash
export fprocess="sha512sum"
export call_timeout="5s"
export read_timeout="10s"
export write_timeout="10s"
frunner
```

## Install
```bash
go get -u -v github.com/trusch/frunner/cmd/frunner
```
