frunner
=======

## Scope

This provides an HTTP server which executes the configured binary with the input from the request and returns the result.
Data is passed around via io.Reader/io.Writer so that data sized > available memory is possible.

Additionally a interface is provided to enable developers to create customized versions of the frunner to prevent forking a process for every call.

## Configuration
```bash
frunner --help
Usage of frunner:
  -t, --call-timeout duration         call timeout (default 5s)
  -l, --http-addr string              http listen address (default ":8080")
  -r, --http-read-timeout duration    http read timeout (default 5s)
  -w, --http-write-timeout duration   http write timeout (default 5s)
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
