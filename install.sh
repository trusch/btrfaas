#!/bin/bash
set -e

curl -L https://github.com/trusch/btrfaas/releases/download/v0.2.0/btrfaasctl > /tmp/btrfaasctl
chmod +x /tmp/btrfaasctl
sudo mv /tmp/btrfaasctl /usr/bin/
docker pull btrfaas/fgateway:v0.2.0
docker pull btrfaas/frunner:v0.2.0

exit 0
