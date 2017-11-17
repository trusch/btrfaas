#!/bin/bash
set -e
echo "running release script"
echo "first do a clean followed by integration tests"
make clean
make all

echo "Current Version: " $(git describe)
read -e -p "New Version Tag:  " TAG
if [[ $(git describe) != "${TAG}" ]]; then
  echo "Tagging current HEAD as ${TAG}"
  git tag -a -m "version ${TAG}" ${TAG}
fi
echo "doing docker pushes"
docker tag btrfaas/fgateway:latest btrfaas/fgateway:${TAG}
docker tag btrfaas/frunner:latest btrfaas/frunner:${TAG}
docker tag btrfaas/fui:latest btrfaas/fui:${TAG}
docker push btrfaas/fgateway:latest
docker push btrfaas/frunner:latest
docker push btrfaas/fui:latest
docker push btrfaas/fgateway:${TAG}
docker push btrfaas/frunner:${TAG}
docker push btrfaas/fui:${TAG}

rm -rf release || true
mkdir release
cp btrfaasctl/btrfaasctl release/btrfaasctl
cp fgateway/fgateway release/fgateway
cp frunner/cmd/frunner/frunner release/frunner
cp fui/fui release/fui
tar cfvj btrfaas-${TAG}.tar.bz2 -C release .

exit 0
