btrfaas
=======

[![Go Report Card](https://goreportcard.com/badge/github.com/trusch/btrfaas)](https://goreportcard.com/report/github.com/trusch/btrfaas)
[![](https://godoc.org/github.com/trusch/btrfaas?status.svg)](http://godoc.org/github.com/trusch/btrfaas)

**b**trfaas is **tr**usch's **f**unction **a**s **a** **s**ervice platform

## Disclaimer
This is heavily inspired by the architecture of [OpenFaaS](https://github.com/openfaas/faas) but has a focus on performance, security and maintainability.

## Features

* swappable deployment platforms (plain docker, swarm and k8s)
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
# install
> curl -sL https://raw.githubusercontent.com/trusch/btrfaas/v0.3.2/install.sh | sh

# init deployment
> btrfaasctl init

# deploy sample functions
> btrfaasctl function deploy https://raw.githubusercontent.com/trusch/btrfaas/v0.3.2/examples/sed.yaml
> btrfaasctl function deploy https://raw.githubusercontent.com/trusch/btrfaas/v0.3.2/examples/to-upper.yaml

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

## Full Setup
This will setup the complete btrfaas stack.
This includes:

* fgateway
* fui (simple web based UI)
* prometheus
* grafana
* two sample functions: "sed" and "to-upper"

```bash
# install
> curl -sL https://raw.githubusercontent.com/trusch/btrfaas/v0.3.2/install.sh | sh

# init deployment
> btrfaasctl init

# deploy fui, prometheus and grafana
> btrfaasctl service deploy https://raw.githubusercontent.com/trusch/btrfaas/v0.3.2/core-services/fui/fui.yaml
> btrfaasctl service deploy https://raw.githubusercontent.com/trusch/btrfaas/v0.3.2/core-services/prometheus/prometheus.yaml
> btrfaasctl service deploy https://raw.githubusercontent.com/trusch/btrfaas/v0.3.2/core-services/prometheus/grafana.yaml

# configure grafana:
> while ! curl -s -H "Content-Type: application/json" \
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
> btrfaasctl function deploy https://raw.githubusercontent.com/trusch/btrfaas/v0.3.2/examples/sed.yaml
> btrfaasctl function deploy https://raw.githubusercontent.com/trusch/btrfaas/v0.3.2/examples/to-upper.yaml
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
