GOOS=linux
GOARCH=amd64

SRC=$(shell find ./btrfaasctl ./deployment ./fgateway ./frunner -type f -name "*.go")

all: vendor fmt vet test frunner btrfaasctl fgateway docker

docker: docker/frunner docker/fgateway

install: btrfaasctl
	cp btrfaasctl/btrfaasctl $(GOPATH)/bin/

frunner: frunner/cmd/frunner/frunner

btrfaasctl: btrfaasctl/btrfaasctl

fgateway: fgateway/fgateway

docker/frunner: frunner
	cd frunner/cmd/frunner && docker build -t trusch/frunner .

docker/fgateway: fgateway
	cd fgateway && docker build -t trusch/fgateway .

btrfaasctl/btrfaasctl: vendor $(SRC)
	docker run \
		-v $(shell pwd):/go/src/github.com/trusch/btrfaas \
		-w /go/src/github.com/trusch/btrfaas/btrfaasctl \
		-u $(shell stat -c '%u:%g' .) \
		-e CGO_ENABLED=0 \
		-e GOOS=$(GOOS) \
		-e GOARCH=$(GOARCH) \
		golang:1.9 \
			go build -v -a -ldflags '-extldflags "-static"' .

fgateway/fgateway: vendor $(SRC)
	docker run \
		-v $(shell pwd):/go/src/github.com/trusch/btrfaas \
		-w /go/src/github.com/trusch/btrfaas/fgateway \
		-u $(shell stat -c '%u:%g' .) \
		-e CGO_ENABLED=0 \
		-e GOOS=$(GOOS) \
		-e GOARCH=$(GOARCH) \
		golang:1.9 \
			go build -v -a -ldflags '-extldflags "-static"' .


frunner/cmd/frunner/frunner: vendor $(SRC)
	docker run \
		-v $(shell pwd):/go/src/github.com/trusch/btrfaas \
		-w /go/src/github.com/trusch/btrfaas/frunner/cmd/frunner \
		-u $(shell stat -c '%u:%g' .) \
		-e CGO_ENABLED=0 \
		-e GOOS=$(GOOS) \
		-e GOARCH=$(GOARCH) \
		golang:1.9 \
			go build -v -a -ldflags '-extldflags "-static"' .

vendor: glide.yaml
	docker run \
		-v $(shell pwd):/go/src/github.com/trusch/btrfaas \
		-w /go/src/github.com/trusch/btrfaas \
		-u $(shell stat -c '%u:%g' .) \
		-e CGO_ENABLED=0 \
		-e GOOS=$(GOOS) \
		-e GOARCH=$(GOARCH) \
		golang:1.9 bash -c \
			"(curl https://glide.sh/get | sh) && glide --home /tmp update"

test: vendor
	docker run \
		-v $(shell pwd):/go/src/github.com/trusch/btrfaas \
		-w /go/src/github.com/trusch/btrfaas \
		-u $(shell stat -c '%u:%g' .) \
		-e CGO_ENABLED=0 \
		-e GOOS=$(GOOS) \
		-e GOARCH=$(GOARCH) \
		golang:1.9 \
			go test -v -cover ./...

vet: vendor
	docker run \
		-v $(shell pwd):/go/src/github.com/trusch/btrfaas \
		-u $(shell stat -c '%u:%g' .) \
		-e CGO_ENABLED=0 \
		-e GOOS=$(GOOS) \
		-e GOARCH=$(GOARCH) \
		golang:1.9 \
			go vet github.com/trusch/btrfaas/...

fmt: vendor
	docker run \
		-v $(shell pwd):/go/src/github.com/trusch/btrfaas \
		-u $(shell stat -c '%u:%g' .) \
		-e CGO_ENABLED=0 \
		-e GOOS=$(GOOS) \
		-e GOARCH=$(GOARCH) \
		golang:1.9 \
			go fmt github.com/trusch/btrfaas/...
