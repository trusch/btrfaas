#!/bin/bash

BTRFAASCTL=${BTRFAASCTL:-"btrfaasctl"}

# build prometheus with config for scraping http://fgateway:8080/metrics
pushd core-services/prometheus
docker build -t btrfaas/prometheus .

# deploy prometheus and grafana
${BTRFAASCTL} service deploy prometheus.yaml grafana.yaml

# wait for grafana and configure prometheus data source
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

popd

exit $?
