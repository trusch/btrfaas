#!/bin/bash
set -e
echo "###################################"
echo "##     DOWNLOADING BTRFAASCTL    ##"
echo "###################################"
curl -L https://github.com/trusch/btrfaas/releases/download/v0.2.0/btrfaasctl > /tmp/btrfaasctl
chmod +x /tmp/btrfaasctl
sudo mv /tmp/btrfaasctl /usr/bin/
echo "###################################"
echo "##  PULLING GATEWAY AND FRUNNER  ##"
echo "###################################"
docker pull btrfaas/fgateway:v0.2.0
docker pull btrfaas/frunner:v0.2.0
docker tag btrfaas/fgateway:v0.2.0 btrfaas/fgateway:latest
docker tag btrfaas/frunner:v0.2.0 btrfaas/frunner:latest
echo "###################################"
echo "##           READY!              ##"
echo "###################################"
echo ""
echo "You can now init a new local deployment with 'btrfaasctl init'."
exit 0
