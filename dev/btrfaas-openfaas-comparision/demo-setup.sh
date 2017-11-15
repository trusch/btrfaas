#!/bin/bash

btrfaasctl --platform swarm --faas-provider btrfaas init
btrfaasctl --platform swarm --faas-provider btrfaas function deploy $(find examples/btrfaas -name "echo-*.yaml")
btrfaasctl --platform swarm --faas-provider openfaas init
btrfaasctl --platform swarm --faas-provider openfaas function deploy examples/openfaas/echo.yaml

pushd dev/btrfaas-openfaas-comparision/prometheus
bash setup.sh
popd

# wait for grafana and add example dashboard
while ! curl -s -H "Content-Type: application/json" \
    -XPOST http://admin:admin@127.0.0.1:3000/api/dashboards/db \
    -d @- < dev/btrfaas-openfaas-comparision/echo-dashboard.json
do sleep 1; done

echo ""
echo "everything deployed, visit grafana on http://127.0.0.1:3000/dashboard/db/echo?refresh=10s&orgId=1"

exit 0
