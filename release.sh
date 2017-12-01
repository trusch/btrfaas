#!/bin/bash
set -e
rm -rf release || true
mkdir release

cp gopath/bin/btrfaasctl release/btrfaasctl.amd64
cp gopath/bin/fgateway release/fgateway.amd64
cp gopath/bin/frunner release/frunner.amd64
cp gopath/bin/fui release/fui.amd64

cp gopath/bin/linux_arm/btrfaasctl release/btrfaasctl.arm
cp gopath/bin/linux_arm/fgateway release/fgateway.arm
cp gopath/bin/linux_arm/frunner release/frunner.arm
cp gopath/bin/linux_arm/fui release/fui.arm

cp gopath/bin/linux_arm64/btrfaasctl release/btrfaasctl.arm64
cp gopath/bin/linux_arm64/fgateway release/fgateway.arm64
cp gopath/bin/linux_arm64/frunner release/frunner.arm64
cp gopath/bin/linux_arm64/fui release/fui.arm64

tar cfvj btrfaas-$(git describe).tar.bz2 -C release .

exit 0
