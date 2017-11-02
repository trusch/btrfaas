frunner
=======

## Scope

This provides an HTTP server which executes the configured binary with the input from the request and returns the result.
Data is passed around via io.Reader/io.Writer so that data sized > available memory is possible.

Additionally a interface is provided to enable developers to create customized versions of the frunner to prevent forking a process for every call.

## OpenFaaS

This is intended to be used with openfaas and can hopefully be used as a replacement for the fwatchdog.
This is heavily inspired by openfaas/faas/watchdog and all credits for this are going to @alexellis. Thank you for your fantastic work :)

## Install
```bash
go get -d github.com/trusch/btrfaas/frunner/cmd/frunner
cd ${GOPATH}/github.com/trusch/btrfaas/frunner
make all
make install
```

## Configuration
```bash
frunner --help
Usage of frunner:
  -b, --buffer                  buffer output before writing
  -t, --call-timeout duration   function call timeout
  -f, --framer string           afterburn framer to use: line, json or http
  -l, --http-addr string        http listen address (default ":8080")
  -h, --http-timeout duration   http timeout for reading request headers (default 1s)
      --read-limit int          limit the amount of data which can be contained in a requests body (default -1)
```

A typical call would look like this:
`frunner --http-addr :8080 -- cat -`

This would result in an HTTP echo server ("cat -" just rewrites stdin to stdout)
Everything after the "--" is interpreted as the executable and its arguments

You can also configure the `frunner` via environment variables:
```bash
# export FRUNNER_CALL_TIMEOUT="5s"
# export FRUNNER_HTTP_TIMEOUT="1s"
# export FRUNNER_HTTP_ADDRESS=":8080"
# export FRUNNER_READ_LIMIT=1024
# export FRUNNER_BUFFER=false
# export FRUNNER_FRAMER=http
export FRUNNER_CMD="sha512sum"
frunner
```


## Afterburn

If you specify `--framer` or `FRUNNER_FRAMER` you execute your binary in afterburn mode.
This means that the executable is called once the `frunner` is started and requests are feeded into the process as they arive.
When the process finished processing a request, it writes the result to stdout. The response is written by the http handler and forwarded to the client.
This is where the framer is needed. It is needed to decide when an anwser package from the binary is complete. There are three framers currently available:

* line: every output line is one message
* json: every complete JSON document is one message
* http: every complete HTTP Response is one message
