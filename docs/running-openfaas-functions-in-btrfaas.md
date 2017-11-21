Run OpenFaaS functions
======================

It is possible to run openfaas functions without modifications by explicit specifying the transport layer.
You can even mix the functions with other functions in a pipeline.

Anyway, this brings some problems:

* openfaas functions will block your pipeline until EOF is sent
* the watchdog will copy the entire input in memory before start working
* if an error occurs after the first output byte is send, we can not catch it
* if you cancel your request, or a timeout occurs, the watchdog will not be informed.

```bash
# init btrfaas
> btrfaasctl init

# deploy sample functions
> btrfaasctl function deploy examples/echo-shell.yaml
> btrfaasctl function deploy examples/echo-openfaas.yaml

# call sample functions
> echo "hello world" | btrfaasctl function invoke "http://echo-openfaas | echo-shell"

# Teardown
> btrfaasctl teardown
```
