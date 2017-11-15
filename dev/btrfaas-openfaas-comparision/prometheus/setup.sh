#!/bin/bash


# build prometheus with config for scraping http://fgateway:8080/metrics
docker build -t btrfaas/prometheus .

# deploy prometheus and grafana
btrfaasctl --platform swarm service deploy prometheus.yaml grafana.yaml

# wait for grafana and configure prometheus data source
while ! curl -s -H "Content-Type: application/json" \
    -XPOST http://admin:admin@127.0.0.1:3000/api/datasources \
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

exit $?
