btrfaas
=======

[![Go Report Card](https://goreportcard.com/badge/github.com/trusch/btrfaas)](https://goreportcard.com/report/github.com/trusch/btrfaas)
[![](https://godoc.org/github.com/trusch/btrfaas?status.svg)](http://godoc.org/github.com/trusch/btrfaas)

**b**trfaas is **tr**usch's **f**unction **a**s **a** **s**ervice platform

## What is this?
Btrfaas is a framework for developing and deploying serverless applications. It provides ways to bootstrap, build and deploy functions as a basic building blocks of your next project.

You can deploy your btrfaas clusters natively on either kubernetes, docker swarm, or even locally on plain docker for development use cases. So developing and testing your functions and services locally and deploying them at scale becomes trivial.

## Features

* encrypted, authenticated, high performant gRPC communication
* easy function bootstrapping
* native templates in go, nodejs, python and bash
* deploy anywhere: plain docker, swarm or k8s
* simple command line client
* function can consume options and secrets
* native function chaining support
* no data buffering, true streaming
* function can be native gRPC servers or openfaas-like stdin/stdout programs
* can run every openfaas function with minor modifications natively (swap watchdog with frunner)
* can run every openfaas function without modifications for backward compability
* prometheus metrics built-in
* first level support for arbitary services like prometheus, grafana...
* ARM and ARM64 support

## Why another FaaS platform?
I started working on this after playing a while with [OpenFaaS](https://github.com/openfaas/faas). I love this project. Its focus on developer/user experience and its simplicity are a big plus. Unfortunately I encounterd some serious problems when evaluating if openfaas would be usable in a range of edge cases coming from a big data background. Some of theses problems are:

* process more data in a function call than available memory
* efficently chain multiple functions (for memory reasons like above)
* interrupt long running tasks cleanly
* apply per-call timeouts
* error handling after the first output byte has been sent (no chance to catch that in openfaas)

Additionally here are some general problems I found regarding production readyness:

* Security:
  * the openfaas gateway is not only a gateway, it also concentrates all deployment logic.
  * to do so, it needs access to the docker socket.
  * If you are able to hijack the gateway, you hijacked the complete system and can do whatever you want.
  * All communication is unencrypted and unauthenticated.
* Performance:
  * Every function call gets dispatched by the gateway using a gorilla mux.
  * HTTP/1.1 is used so the known limitations in terms of connection recycling etc. occurs.
  * On function level, each function call leads to a full fork and exec cycle of the binary which does the work.
* Maintainability
  * although being relatively simple the codebase is distributed over many repositories
  * there are no clear layers in the codebase. For example the create-function http handler directly includes a call to the docker swarm api. I don't know how this works together with k8s ;-)
  * the watchdog (the binary which does the fork/exec for every function call) has a cyclomatic complexity > 15, just for example

All I wanted was a version of openfaas which solves the problems above, and I hope that some of the ideas from this project will find its way back into the original openfaas codebase.

## Getting Started
```bash
# install
curl -sL https://raw.githubusercontent.com/trusch/btrfaas/v0.3.3/install.sh | sh

# init deployment
btrfaasctl init

# deploy sample functions
btrfaasctl function deploy https://raw.githubusercontent.com/trusch/btrfaas/v0.3.3/examples/sed.yaml
btrfaasctl function deploy https://raw.githubusercontent.com/trusch/btrfaas/v0.3.3/examples/to-upper.yaml

# test it
echo "I hate this" | btrfaasctl function invoke "sed -e s/hate/love/ | to-upper"
I LOVE THIS

# Teardown
btrfaasctl teardown
```

## Build your own functions
```bash
# bootstrap function
btrfaasctl function init my-echo --template go

# edit ./my-echo/ to fit your needs

# build and deploy
btrfaasctl function build my-echo
btrfaasctl function deploy my-echo/function.yaml

# test it
echo "Hello World" | btrfaasctl function invoke my-echo
Hello World
```

## Full Setup
This will setup the complete btrfaas stack.
This includes:

* fgateway
* fui (simple web based UI)
* prometheus
* grafana
* two sample functions: "sed" and "to-upper"

```bash
# init deployment
btrfaasctl init

# deploy fui, prometheus and grafana
btrfaasctl service deploy https://raw.githubusercontent.com/trusch/btrfaas/v0.3.3/core-services/fui/fui.yaml
btrfaasctl service deploy https://raw.githubusercontent.com/trusch/btrfaas/v0.3.3/core-services/prometheus/prometheus.yaml
btrfaasctl service deploy https://raw.githubusercontent.com/trusch/btrfaas/v0.3.3/core-services/prometheus/grafana.yaml

# configure grafana:
while ! curl -s -H "Content-Type: application/json" \
    -XPOST http://admin:admin@localhost:3000/api/datasources \
    -d @- <<EOF
{
    "name": "prometheus",
    "type": "prometheus",
    "access": "proxy",
    "isDefault": true,
    "url": "http://prometheus:9090"
}
EOF
do sleep 1; done

# deploy sample functions
btrfaasctl function deploy https://raw.githubusercontent.com/trusch/btrfaas/v0.3.3/examples/sed.yaml
btrfaasctl function deploy https://raw.githubusercontent.com/trusch/btrfaas/v0.3.3/examples/to-upper.yaml
```

You can now visit:

* fui on `http://localhost:8000`
* prometheus on `http://localhost:9000`
* grafana on `http://localhost:3000`

## How to Contribute
Contributions are welcome, please feel free to open a PR!
If you find a bug or have an idea on how to improve things, open an issue.
PR's are accepted if they follow the used coding standards, and the go-report keeps on 100%.
If you add end-user features, it would be great to see them integrated into the smoke tests.
