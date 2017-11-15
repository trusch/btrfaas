btrfaas
=======
**b**trfaas is **tr**usch's **f**unction **a**s **a** **s**ervice platform

## Disclaimer
This is heavily inspired by the architecture of [OpenFaaS](https://github.com/openfaas/openfaas) but has a focus on performance and maintainability.

## Features

* swappable deployment platforms (currently swarm, k8s in progress ;))
* simple command line client
* use secrets
* function can receive options
* native function chaining support
* communication is over gRPC
* no data buffering, true streaming
* easy to build functions
* can run every OpenFaaS function with minor modifications (swap watchdog with frunner)
* first level support for arbitary (non function) services

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
