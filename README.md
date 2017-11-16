btrfaas
=======
**b**trfaas is **tr**usch's **f**unction **a**s **a** **s**ervice platform

## Disclaimer
This is heavily inspired by the architecture of [OpenFaaS](https://github.com/openfaas/faas) but has a focus on performance and maintainability.

## Features

* swappable deployment platforms (currently docker and swarm, k8s in progress ;))
* simple command line client
* use secrets
* function can receive options
* native function chaining support
* communication is over gRPC
* no data buffering, true streaming
* easy to build functions
* can run every OpenFaaS function with minor modifications (swap watchdog with frunner)
* first level support for arbitary (non function) services
* it is possible to deploy the openfaas gateway and functions

## Walk Through
```bash
# use GOOS=darwin when on mac to build a mac compatible version of btrfaasctl
> make all GOOS=linux # build everything (frunner, fgateway, btrfaasctl + docker images)
> btrfaasctl init     # init deployment

# deploy functions
> btrfaasctl function deploy examples/btrfaas/sed.yaml
> btrfaasctl function deploy examples/btrfaas/to-upper.yaml
> btrfaasctl function deploy examples/btrfaas/**/echo-*.yaml

# test it
> echo "I hate this" | btrfaasctl function invoke "sed e=s/hate/love/ | to-upper"
I LOVE THIS
> echo "foobar" | btrfaasctl function invoke "echo-go | echo-node | echo-python | echo-shell"
foobar
```

## Deploy BtrFaaS and OpenFaaS side by side
```bash
# init faas's
> btrfaasctl --platform swarm --faas-provider btrfaas init
> btrfaasctl --platform swarm --faas-provider openfaas init
# deploy sample functions
> btrfaasctl --platform swarm --faas-provider btrfaas function deploy examples/btrfaas/echo-shell.yaml
> btrfaasctl --platform swarm --faas-provider openfaas function deploy examples/openfaas/echo.yaml
# call sample functions
> echo "hello btrfaas" | btrfaasctl --platform swarm --faas-provider btrfaas function invoke echo-shell
> echo "hello openfaas" | btrfaasctl --platform swarm --faas-provider openfaas function invoke echo
```

## Run BtrFaaS and OpenFaaS side by side without openfaas-gateway
It is possible to run openfaas functions without modifications by explicit specifying the transport layer.
You can even mix the functions with other functions in a pipeline.

Anyway, this brings some problems:

* openfaas functions will block your pipeline until EOF is sent
* the watchdog will copy the entire input in memory before start working
* if an error occurs after the first output byte is send, we can not catch it
* if you cancel your request, or a timeout occurs, the watchdog will not be informed.

```bash
> btrfaasctl init
# deploy sample functions
> btrfaasctl function deploy examples/btrfaas/echo-shell.yaml
> btrfaasctl function deploy examples/openfaas/echo.yaml
# call sample functions
> echo "hello world" | btrfaasctl function invoke "http://echo | echo-shell"
```

## Run the FaaS comparision demo
This will start an increasing load against the four btrfaas echo implementations and the openfaas echo implementation. You can inspect the average call durations on the commandline or look at a grafana dashboard showing you requests per second and request latencies.
```bash
> bash dev/btrfaas-openfaas-comparision/demo-setup.sh
> www-browser http://127.0.0.1:3000/dashboard/db/echo?refresh=10s&orgId=1
> go run dev/btrfaas-openfaas-comparision/bench.go
```
