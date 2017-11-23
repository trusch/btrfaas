#!/bin/bash

C=10
N=1000

pushd dev/echo-demo
go build -o bench bench.go
./bench -function echo-go -c ${C} -n ${N}
./bench -function echo-node -c ${C} -n ${N}
./bench -function echo-python -c ${C} -n ${N}
./bench -function echo-shell -c ${C} -n ${N}
./bench -function http://echo-openfaas -c ${C} -n ${N}
popd

exit 0
