Deploy BtrFaaS and OpenFaaS side by side
========================================

```bash
# init faas's
> btrfaasctl --platform swarm --faas-provider btrfaas init
> btrfaasctl --platform swarm --faas-provider openfaas init

# deploy sample functions
> btrfaasctl --platform swarm --faas-provider btrfaas function deploy examples/echo-shell.yaml
> btrfaasctl --platform swarm --faas-provider openfaas function deploy examples/echo-openfaas.yaml

# call sample functions
> echo "hello btrfaas" | btrfaasctl --platform swarm --faas-provider btrfaas function invoke echo-shell
> echo "hello openfaas" | btrfaasctl --platform swarm --faas-provider openfaas function invoke echo-openfaas

# Teardown
> btrfaasctl --platform swarm teardown
```
