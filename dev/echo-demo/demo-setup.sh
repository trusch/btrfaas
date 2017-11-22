#!/bin/bash

btrfaasctl init

btrfaasctl service deploy \
  core-services/prometheus/prometheus.yaml \
  core-services/prometheus/grafana.yaml

btrfaasctl function deploy \
  examples/echo-go/function.yaml \
  examples/echo-python/function.yaml \
  examples/echo-node/function.yaml \
  examples/echo-shell.yaml \
  examples/echo-openfaas.yaml

# wait for grafana and configure prometheus data source
while ! curl -s -H "Content-Type: application/json" \
    -XPOST http://admin:admin@localhost:3000/api/datasources \
    -d @- 2>&1 >/dev/null <<EOF
{
    "name": "prometheus",
    "type": "prometheus",
    "access": "proxy",
    "isDefault": true,
    "url": "http://prometheus:9090"
}
EOF
do sleep 1; done

# add example dashboard
curl -s -H "Content-Type: application/json" \
    -XPOST http://admin:admin@localhost:3000/api/dashboards/db \
    -d @- < dev/echo-demo/echo-dashboard.json 2>&1 >/dev/null

echo ""
echo "everything deployed, visit grafana on http://localhost:3000/dashboard/db/echo?refresh=5s&orgId=1"
echo ""

exit 0
