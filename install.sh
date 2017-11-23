#!/bin/bash
set -e

echo "###################################"
echo "##     DOWNLOADING BTRFAASCTL    ##"
echo "###################################"
curl -L https://github.com/trusch/btrfaas/releases/download/v0.2.2/btrfaasctl > /tmp/btrfaasctl
chmod +x /tmp/btrfaasctl
sudo mv /tmp/btrfaasctl /usr/bin/

echo "###################################"
echo "##    PULLING NEEDED IMAGES      ##"
echo "###################################"
docker pull btrfaas/fgateway:v0.2.2
docker pull btrfaas/frunner:v0.2.2
docker pull btrfaas/fui:v0.2.2
docker pull btrfaas/prometheus:v0.2.2
docker pull grafana/grafana
docker tag btrfaas/fgateway:v0.2.2 btrfaas/fgateway:latest
docker tag btrfaas/frunner:v0.2.2 btrfaas/frunner:latest
docker tag btrfaas/fui:v0.2.2 btrfaas/fui:latest
docker tag btrfaas/prometheus:v0.2.2 btrfaas/prometheus:latest

echo "###################################"
echo "##           READY!              ##"
echo "###################################"
echo ""
echo "You can now init a new local deployment with 'btrfaasctl init'."
echo ""

exit 0
