GOOS=linux
GOARCH=amd64

SRC=$(shell find ./btrfaasctl ./deployment ./fgateway ./frunner ./faas ./fui ./grpc ./template ./pki -type f -name "*.go")

all: vendor fmt vet unit-tests frunner btrfaasctl fgateway fui docker install prepare-templates echo-examples integration-tests

docker: docker/frunner docker/fgateway docker/fui echo-examples

install: btrfaasctl
	cp btrfaasctl/btrfaasctl $(GOPATH)/bin/

clean:
	rm -rf vendor btrfaasctl/btrfaasctl fgateway/fgateway frunner/cmd/frunner/frunner

frunner: frunner/cmd/frunner/frunner

btrfaasctl: btrfaasctl/btrfaasctl

fgateway: fgateway/fgateway

fui: fui/fui

docker/frunner: frunner
	cd frunner/cmd/frunner && docker build --no-cache -t btrfaas/frunner .

docker/fgateway: fgateway
	cd fgateway && docker build --no-cache -t btrfaas/fgateway .

docker/fui: fui
	cd fui && docker build --no-cache -t btrfaas/fui .

btrfaasctl/btrfaasctl: vendor $(SRC)
	docker run \
		-v $(shell pwd):/go/src/github.com/trusch/btrfaas \
		-w /go/src/github.com/trusch/btrfaas/btrfaasctl \
		-u $(shell ls -n .|tail -1|tr -s ' '|awk '{print $$3 ":" $$4}') \
		-e CGO_ENABLED=0 \
		-e GOOS=$(GOOS) \
		-e GOARCH=$(GOARCH) \
		golang:1.9 \
			go build -v -a -ldflags '-extldflags "-static"' .

fgateway/fgateway: vendor $(SRC)
	docker run \
		-v $(shell pwd):/go/src/github.com/trusch/btrfaas \
		-w /go/src/github.com/trusch/btrfaas/fgateway \
		-u $(shell ls -n .|tail -1|tr -s ' '|awk '{print $$3 ":" $$4}') \
		-e CGO_ENABLED=0 \
		golang:1.9 \
			go build -v -a -ldflags '-extldflags "-static"' .


frunner/cmd/frunner/frunner: vendor $(SRC)
	docker run \
		-v $(shell pwd):/go/src/github.com/trusch/btrfaas \
		-w /go/src/github.com/trusch/btrfaas/frunner/cmd/frunner \
		-u $(shell ls -n .|tail -1|tr -s ' '|awk '{print $$3 ":" $$4}') \
		-e CGO_ENABLED=0 \
		golang:1.9 \
			go build -v -a -ldflags '-extldflags "-static"' .

fui/fui: vendor $(SRC)
	docker run \
		-v $(shell pwd):/go/src/github.com/trusch/btrfaas \
		-w /go/src/github.com/trusch/btrfaas/fui \
		-u $(shell ls -n .|tail -1|tr -s ' '|awk '{print $$3 ":" $$4}') \
		-e CGO_ENABLED=0 \
		golang:1.9 \
			go build -v -a -ldflags '-extldflags "-static"' .

vendor: glide.yaml
	docker run \
		-v $(shell pwd):/go/src/github.com/trusch/btrfaas \
		-w /go/src/github.com/trusch/btrfaas \
		-u $(shell ls -n .|tail -1|tr -s ' '|awk '{print $$3 ":" $$4}') \
		-e CGO_ENABLED=0 \
		-e GOOS=$(GOOS) \
		-e GOARCH=$(GOARCH) \
		golang:1.9 bash -c \
			"(curl https://glide.sh/get | sh) && glide --home /tmp update -v"

unit-tests: vendor
	docker run \
		-v $(shell pwd):/go/src/github.com/trusch/btrfaas \
		-w /go/src/github.com/trusch/btrfaas \
		-u $(shell ls -n .|tail -1|tr -s ' '|awk '{print $$3 ":" $$4}') \
		-e CGO_ENABLED=0 \
		golang:1.9 \
			go test -v -cover ./deployment/... ./fgateway/... ./frunner/...

integration-tests: install
	cd integration-tests && ginkgo -r -v --slowSpecThreshold 20

vet: vendor
	docker run \
		-v $(shell pwd):/go/src/github.com/trusch/btrfaas \
		-u $(shell ls -n .|tail -1|tr -s ' '|awk '{print $$3 ":" $$4}') \
		-e CGO_ENABLED=0 \
		golang:1.9 \
			go vet github.com/trusch/btrfaas/...

fmt: vendor
	docker run \
		-v $(shell pwd):/go/src/github.com/trusch/btrfaas \
		-u $(shell ls -n .|tail -1|tr -s ' '|awk '{print $$3 ":" $$4}') \
		-e CGO_ENABLED=0 \
		golang:1.9 \
			gofmt -e -s -w \
				/go/src/github.com/trusch/btrfaas/btrfaasctl \
				/go/src/github.com/trusch/btrfaas/btrfaasctl/cmd \
				/go/src/github.com/trusch/btrfaas/deployment \
				/go/src/github.com/trusch/btrfaas/deployment/docker \
				/go/src/github.com/trusch/btrfaas/deployment/swarm \
				/go/src/github.com/trusch/btrfaas/dev/btrfaas-openfaas-comparision \
				/go/src/github.com/trusch/btrfaas/faas \
				/go/src/github.com/trusch/btrfaas/faas/btrfaas \
				/go/src/github.com/trusch/btrfaas/faas/openfaas \
				/go/src/github.com/trusch/btrfaas/fgateway \
				/go/src/github.com/trusch/btrfaas/fgateway/cmd \
				/go/src/github.com/trusch/btrfaas/fgateway/forwarder \
				/go/src/github.com/trusch/btrfaas/fgateway/grpc \
				/go/src/github.com/trusch/btrfaas/fgateway/http \
				/go/src/github.com/trusch/btrfaas/fgateway/metrics \
				/go/src/github.com/trusch/btrfaas/frunner/cmd \
				/go/src/github.com/trusch/btrfaas/frunner/config \
				/go/src/github.com/trusch/btrfaas/frunner/env \
				/go/src/github.com/trusch/btrfaas/frunner/grpc \
				/go/src/github.com/trusch/btrfaas/frunner/http \
				/go/src/github.com/trusch/btrfaas/frunner/runnable \
				/go/src/github.com/trusch/btrfaas/frunner/runnable/exec \
				/go/src/github.com/trusch/btrfaas/frunner/runnable/chain \
				/go/src/github.com/trusch/btrfaas/fui \
				/go/src/github.com/trusch/btrfaas/fui/cmd \
				/go/src/github.com/trusch/btrfaas/grpc \
				/go/src/github.com/trusch/btrfaas/integration-tests

prepare-templates:
	cp grpc/frunner.proto templates/python
	cd templates/python && python -m grpc_tools.protoc -I . --python_out=. --grpc_python_out=. frunner.proto
	cp grpc/frunner.proto templates/nodejs

echo-examples: install examples/echo-go examples/echo-node examples/echo-python

examples/echo-go: templates/go
	mkdir -p examples
	btrfaasctl function init examples/echo-go --template go
	btrfaasctl function build examples/echo-go

examples/echo-node: templates/nodejs
	mkdir -p examples
	btrfaasctl function init examples/echo-node --template nodejs
	btrfaasctl function build examples/echo-node

examples/echo-python: templates/nodejs
	mkdir -p examples
	btrfaasctl function init examples/echo-python --template python
	btrfaasctl function build examples/echo-python
