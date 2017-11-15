#!/bin/bash

btrfaasctl init
btrfaasctl function deploy $(find examples/btrfaas -name "echo-*.yaml")

bash core-services/prometheus/setup.sh

# wait for grafana and add example dashboard
while ! curl -s -H "Content-Type: application/json" \
    -XPOST http://admin:admin@localhost:3000/api/dashboards/db \
    -d @- < dev/echo-dashboard.json
do sleep 1; done

echo ""
echo "everything deployed, visit grafana on http://localhost:3000/dashboard/db/echo?refresh=10s&orgId=1"

exit 0