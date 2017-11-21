btrfaas
=======

[![Go Report Card](https://goreportcard.com/badge/github.com/trusch/btrfaas)](https://goreportcard.com/report/github.com/trusch/btrfaas)
[![](https://godoc.org/github.com/trusch/btrfaas?status.svg)](http://godoc.org/github.com/trusch/btrfaas)

**b**trfaas is **tr**usch's **f**unction **a**s **a** **s**ervice platform

## Disclaimer
This is heavily inspired by the architecture of [OpenFaaS](https://github.com/openfaas/faas) but has a focus on performance and maintainability.

## Features

* swappable deployment platforms (currently docker and swarm, k8s in progress ;))
* encrypted gRPC communication
* simple command line client
* support for function secrets
* support for function options
* native function chaining support
* no data buffering, true streaming
* easy to build functions
* function can be native gRPC servers or openfaas-like stdin/stdout programs
* can run every OpenFaaS function with minor modifications natively (swap watchdog with frunner)
* can run every OpenFaaS function without modifications for backward compability
* first level support for arbitary (non function) services
* it is even possible to deploy the openfaas gateway

## Getting Started
```bash
> curl -sL https://github.com/trusch/btrfaas/releases/download/v0.2.0/btrfaasctl > /tmp/btrfaasctl
> chmod +x /tmp/btrfaasctl
> sudo mv /tmp/btrfaasctl /usr/bin/
> docker pull btrfaas/fgateway:v0.2.0
> docker pull btrfaas/frunner:v0.2.0
> btrfaasctl init     # init deployment

# deploy sample functions
> btrfaasctl function deploy https://raw.githubusercontent.com/trusch/btrfaas/v0.2.0/examples/sed.yaml
> btrfaasctl function deploy https://raw.githubusercontent.com/trusch/btrfaas/v0.2.0/examples/to-upper.yaml

# test it
> echo "I hate this" | btrfaasctl function invoke "sed -e s/hate/love/ | to-upper"
I LOVE THIS

# Teardown
> btrfaasctl teardown
```

## Build your own functions
```bash
# init deployment
> btrfaasctl init

# create and deploy function
> btrfaasctl function init my-echo --template go
# edit ./my-echo/ to fit your needs
> btrfaasctl function build my-echo
> btrfaasctl function deploy my-echo/function.yaml

# test it
> echo "Hello World" | btrfaasctl function invoke my-echo
Hello World

# Teardown
> btrfaasctl teardown
```

## How to Contribute
Contributions are welcome, please feel free to open a PR!
If you find a bug or have an idea on how to improve things, open an issue.
PR's are accepted if they follow the used coding standards, and the go-report keeps on 100%.
If you add end-user features, it would be great to see them integrated into the smoke tests.
