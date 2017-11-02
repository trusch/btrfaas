btrfaas
=======
**b**trfaas is **tr**usch's **f**unction **a**s **a** **s**ervice platform

## Disclaimer
This is heavily inspired by the architecture of [OpenFaaS](https://github.com/openfaas/openfaas) but has a focus on performance and maintainability.

## Features

* swappable deployment platforms (currently swarm, k8s in progress ;))
* simple command line client
* use secrets
* function-gateway to function-service communication is over gRPC
* no data buffering, true streaming
* easy to build functions
* can run every OpenFaaS function with minor modifications (swap watchdog with frunner)
* first level support for arbitary (non function) services

## Walk Through
```bash
> make all        # build everything (frunner, fgateway, btrfaasctl + docker images)
> make install    # install btrfaasctl to $GOPATH/bin
> btrfaasctl init # init deployment

# deploy function gateway + example function
> btrfaasctl service deploy core-services/fgateway.yaml
> btrfaasctl service deploy service-examples/echo.yaml

# test it
> curl -d 'foobar' http://localhost:8080/api/v0/invoke/echo
foobar
```
