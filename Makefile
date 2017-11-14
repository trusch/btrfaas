GOOS=linux
GOARCH=amd64

SRC=$(shell find ./btrfaasctl ./deployment ./fgateway ./frunner ./faas -type f -name "*.go")

all: vendor fmt vet unit-tests frunner btrfaasctl fgateway docker install integration-tests

docker: docker/frunner docker/fgateway echo-examples

install: btrfaasctl
	cp btrfaasctl/btrfaasctl $(GOPATH)/bin/

clean:
	rm -rf vendor btrfaasctl/btrfaasctl fgateway/fgateway frunner/cmd/frunner/frunner

frunner: frunner/cmd/frunner/frunner

btrfaasctl: btrfaasctl/btrfaasctl

fgateway: fgateway/fgateway

docker/frunner: frunner
	cd frunner/cmd/frunner && docker build --no-cache -t btrfaas/frunner .

docker/fgateway: fgateway
	cd fgateway && docker build --no-cache -t btrfaas/fgateway .

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
		-e GOOS=$(GOOS) \
		-e GOARCH=$(GOARCH) \
		golang:1.9 \
			go build -v -a -ldflags '-extldflags "-static"' .


frunner/cmd/frunner/frunner: vendor $(SRC)
	docker run \
		-v $(shell pwd):/go/src/github.com/trusch/btrfaas \
		-w /go/src/github.com/trusch/btrfaas/frunner/cmd/frunner \
		-u $(shell ls -n .|tail -1|tr -s ' '|awk '{print $$3 ":" $$4}') \
		-e CGO_ENABLED=0 \
		-e GOOS=$(GOOS) \
		-e GOARCH=$(GOARCH) \
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
		-e GOOS=$(GOOS) \
		-e GOARCH=$(GOARCH) \
		golang:1.9 \
			go test -v -cover ./deployment/... ./fgateway/... ./frunner/...

integration-tests: install
	cd integration-tests && ginkgo -r -v --slowSpecThreshold 20

vet: vendor
	docker run \
		-v $(shell pwd):/go/src/github.com/trusch/btrfaas \
		-u $(shell ls -n .|tail -1|tr -s ' '|awk '{print $$3 ":" $$4}') \
		-e CGO_ENABLED=0 \
		-e GOOS=$(GOOS) \
		-e GOARCH=$(GOARCH) \
		golang:1.9 \
			go vet github.com/trusch/btrfaas/...

fmt: vendor
	docker run \
		-v $(shell pwd):/go/src/github.com/trusch/btrfaas \
		-u $(shell ls -n .|tail -1|tr -s ' '|awk '{print $$3 ":" $$4}') \
		-e CGO_ENABLED=0 \
		-e GOOS=$(GOOS) \
		-e GOARCH=$(GOARCH) \
		golang:1.9 \
			go fmt github.com/trusch/btrfaas/...

echo-examples:
	cd examples/btrfaas/native-functions/echo-go && docker build --no-cache -t btrfaas/functions/echo-go .
	cp grpc/frunner.proto examples/btrfaas/native-functions/echo-python
	cd examples/btrfaas/native-functions/echo-python && python -m grpc_tools.protoc -I . --python_out=. --grpc_python_out=. frunner.proto
	cd examples/btrfaas/native-functions/echo-python && docker build --no-cache -t btrfaas/functions/echo-python .
	cp grpc/frunner.proto examples/btrfaas/native-functions/echo-node
	cd examples/btrfaas/native-functions/echo-node && docker build --no-cache -t btrfaas/functions/echo-node .
