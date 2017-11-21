Run the FaaS comparision demo
=============================

This will start an increasing load against the four btrfaas echo implementations and the openfaas echo implementation. You can inspect the average call durations on the commandline or look at a grafana dashboard showing you requests per second and request latencies.
```bash
> bash dev/btrfaas-openfaas-comparision/demo-setup.sh
> www-browser http://127.0.0.1:3000/dashboard/db/echo?refresh=10s&orgId=1
> go run dev/btrfaas-openfaas-comparision/bench.go
```
