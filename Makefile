CORE_SRC=$(shell find deployment faas grpc pki template -name "*.go")
BTRFAASCTL_SRC=$(shell find btrfaasctl -name "*.go")
FGATEWAY_SRC=$(shell find fgateway -name "*.go")
FRUNNER_SRC=$(shell find frunner -name "*.go")
FUI_SRC=$(shell find frunner -name "*.go")

SRC=$(CORE_SRC) $(BTRFAASCTL_SRC) $(FGATEWAY_SRC) $(FRUNNER_SRC) $(FUI_SRC)

all: lint binaries docker

fast: .docker/fgateway/amd64 .docker/frunner/amd64 install

binaries: amd64 arm arm64
amd64: gopath/bin/fgateway gopath/bin/frunner gopath/bin/btrfaasctl gopath/bin/fui
arm: gopath/bin/linux_arm/fgateway gopath/bin/linux_arm/frunner gopath/bin/linux_arm/btrfaasctl gopath/bin/linux_arm/fui
arm64: gopath/bin/linux_arm64/fgateway gopath/bin/linux_arm64/frunner gopath/bin/linux_arm64/btrfaasctl gopath/bin/linux_arm64/fui

install: gopath/bin/btrfaasctl
	cp gopath/bin/btrfaasctl $(GOPATH)/bin/

clean:
	sudo rm -rf gopath/ vendor/ .docker/

####################################
#             FGATEWAY             #
####################################
gopath/bin/fgateway: $(CORE_SRC) $(FGATEWAY_SRC) vendor
	docker run --rm \
		-v $(shell pwd)/gopath:/go \
		-v $(shell pwd):/go/src/github.com/trusch/btrfaas \
		-w /go/src/github.com/trusch/btrfaas/fgateway \
		-e CGO_ENABLED=0 \
		-e GOOS=linux \
		-e GOARCH=amd64 \
		golang:1.9 go install -v -ldflags '-extldflags "-static"' .

gopath/bin/linux_arm/fgateway: $(CORE_SRC) $(FGATEWAY_SRC) vendor
	docker run --rm \
		-v $(shell pwd)/gopath:/go \
		-v $(shell pwd):/go/src/github.com/trusch/btrfaas \
		-w /go/src/github.com/trusch/btrfaas/fgateway \
		-e CGO_ENABLED=0 \
		-e GOOS=linux \
		-e GOARCH=arm \
		golang:1.9 go install -v -ldflags '-extldflags "-static"' .

gopath/bin/linux_arm64/fgateway: $(CORE_SRC) $(FGATEWAY_SRC) vendor
	docker run --rm \
		-v $(shell pwd)/gopath:/go \
		-v $(shell pwd):/go/src/github.com/trusch/btrfaas \
		-w /go/src/github.com/trusch/btrfaas/fgateway \
		-e CGO_ENABLED=0 \
		-e GOOS=linux \
		-e GOARCH=arm64 \
		golang:1.9 go install -v -ldflags '-extldflags "-static"' .

####################################
#             FRUNNER             #
####################################
gopath/bin/frunner: $(CORE_SRC) $(FRUNNER_SRC) vendor
	docker run --rm \
		-v $(shell pwd)/gopath:/go \
		-v $(shell pwd):/go/src/github.com/trusch/btrfaas \
		-w /go/src/github.com/trusch/btrfaas/frunner \
		-e CGO_ENABLED=0 \
		-e GOOS=linux \
		-e GOARCH=amd64 \
		golang:1.9 go install -v -ldflags '-extldflags "-static"' .

gopath/bin/linux_arm/frunner: $(CORE_SRC) $(FRUNNER_SRC) vendor
	docker run --rm \
		-v $(shell pwd)/gopath:/go \
		-v $(shell pwd):/go/src/github.com/trusch/btrfaas \
		-w /go/src/github.com/trusch/btrfaas/frunner \
		-e CGO_ENABLED=0 \
		-e GOOS=linux \
		-e GOARCH=arm \
		golang:1.9 go install -v -ldflags '-extldflags "-static"' .

gopath/bin/linux_arm64/frunner: $(CORE_SRC) $(FRUNNER_SRC) vendor
	docker run --rm \
		-v $(shell pwd)/gopath:/go \
		-v $(shell pwd):/go/src/github.com/trusch/btrfaas \
		-w /go/src/github.com/trusch/btrfaas/frunner \
		-e CGO_ENABLED=0 \
		-e GOOS=linux \
		-e GOARCH=arm64 \
		golang:1.9 go install -v -ldflags '-extldflags "-static"' .

####################################
#             BTRFAASCTL           #
####################################
gopath/bin/btrfaasctl: $(CORE_SRC) $(BTRFAASCTL_SRC) vendor
	docker run --rm \
		-v $(shell pwd)/gopath:/go \
		-v $(shell pwd):/go/src/github.com/trusch/btrfaas \
		-w /go/src/github.com/trusch/btrfaas/btrfaasctl \
		-e CGO_ENABLED=0 \
		-e GOOS=linux \
		-e GOARCH=amd64 \
		golang:1.9 go install -v -ldflags '-extldflags "-static"' .

gopath/bin/linux_arm/btrfaasctl: $(CORE_SRC) $(BTRFAASCTL_SRC) vendor
	docker run --rm \
		-v $(shell pwd)/gopath:/go \
		-v $(shell pwd):/go/src/github.com/trusch/btrfaas \
		-w /go/src/github.com/trusch/btrfaas/btrfaasctl \
		-e CGO_ENABLED=0 \
		-e GOOS=linux \
		-e GOARCH=arm \
		golang:1.9 go install -v -ldflags '-extldflags "-static"' .

gopath/bin/linux_arm64/btrfaasctl: $(CORE_SRC) $(BTRFAASCTL_SRC) vendor
	docker run --rm \
		-v $(shell pwd)/gopath:/go \
		-v $(shell pwd):/go/src/github.com/trusch/btrfaas \
		-w /go/src/github.com/trusch/btrfaas/btrfaasctl \
		-e CGO_ENABLED=0 \
		-e GOOS=linux \
		-e GOARCH=arm64 \
		golang:1.9 go install -v -ldflags '-extldflags "-static"' .

####################################
#             FUI                  #
####################################
gopath/bin/fui: $(CORE_SRC) $(FUI_SRC) vendor
	docker run --rm \
		-v $(shell pwd)/gopath:/go \
		-v $(shell pwd):/go/src/github.com/trusch/btrfaas \
		-w /go/src/github.com/trusch/btrfaas/fui \
		-e CGO_ENABLED=0 \
		-e GOOS=linux \
		-e GOARCH=amd64 \
		golang:1.9 go install -v -ldflags '-extldflags "-static"' .

gopath/bin/linux_arm/fui: $(CORE_SRC) $(FUI_SRC) vendor
	docker run --rm \
		-v $(shell pwd)/gopath:/go \
		-v $(shell pwd):/go/src/github.com/trusch/btrfaas \
		-w /go/src/github.com/trusch/btrfaas/fui \
		-e CGO_ENABLED=0 \
		-e GOOS=linux \
		-e GOARCH=arm \
		golang:1.9 go install -v -ldflags '-extldflags "-static"' .

gopath/bin/linux_arm64/fui: $(CORE_SRC) $(FUI_SRC) vendor
	docker run --rm \
		-v $(shell pwd)/gopath:/go \
		-v $(shell pwd):/go/src/github.com/trusch/btrfaas \
		-w /go/src/github.com/trusch/btrfaas/fui \
		-e CGO_ENABLED=0 \
		-e GOOS=linux \
		-e GOARCH=arm64 \
		golang:1.9 go install -v -ldflags '-extldflags "-static"' .


####################################
#           VENDOR STUFF           #
####################################
gopath/bin/glide:
	mkdir -p gopath/bin gopath/src
	docker run --rm \
	-v $(shell pwd)/gopath:/go \
	-e GOOS=$(GOOS) \
	-e GOARCH=$(GOARCH) \
	golang:1.9 bash -c "curl https://glide.sh/get | sh"

vendor: gopath/bin/glide glide.yaml
	docker run --rm \
	-v $(shell pwd)/gopath:/go \
	-v $(shell pwd):/go/src/github.com/trusch/btrfaas \
	-w /go/src/github.com/trusch/btrfaas \
	-e GOOS=$(GOOS) \
	-e GOARCH=$(GOARCH) \
	golang:1.9 glide --home /go/.glide install -v


####################################
#           DOCKER STUFF           #
####################################
docker: .docker/fgateway/amd64 \
	.docker/fgateway/arm \
	.docker/fgateway/arm64 \
	.docker/frunner/amd64 \
	.docker/frunner/arm \
	.docker/frunner/arm64 \
	.docker/fui/amd64 \
	.docker/fui/arm \
	.docker/fui/arm64

.docker/fgateway/amd64: gopath/bin/fgateway
	cp gopath/bin/fgateway fgateway/
	cd fgateway && docker build -t btrfaas/fgateway:latest .
	mkdir -p $@ && touch $@

.docker/fgateway/arm: gopath/bin/linux_arm/fgateway
	cp gopath/bin/linux_arm/fgateway fgateway/
	cd fgateway && docker build -t btrfaas/fgateway:latest-arm -f Dockerfile.arm .
	mkdir -p $@ && touch $@

.docker/fgateway/arm64: gopath/bin/linux_arm64/fgateway
	cp gopath/bin/linux_arm64/fgateway fgateway/
	cd fgateway && docker build -t btrfaas/fgateway:latest-arm64 -f Dockerfile.arm64 .
	mkdir -p $@ && touch $@

.docker/frunner/amd64: gopath/bin/frunner
	cp gopath/bin/frunner frunner/
	cd frunner && docker build -t btrfaas/frunner:latest .
	mkdir -p $@ && touch $@

.docker/frunner/arm: gopath/bin/linux_arm/frunner
	cp gopath/bin/linux_arm/frunner frunner/
	cd frunner && docker build -t btrfaas/frunner:latest-arm -f Dockerfile.arm .
	mkdir -p $@ && touch $@

.docker/frunner/arm64: gopath/bin/linux_arm64/frunner
	cp gopath/bin/linux_arm64/frunner frunner/
	cd frunner && docker build -t btrfaas/frunner:latest-arm64 -f Dockerfile.arm64 .
	mkdir -p $@ && touch $@

.docker/fui/amd64: gopath/bin/fui
	cp gopath/bin/fui fui/
	cd fui && docker build -t btrfaas/fui:latest .
	mkdir -p $@ && touch $@

.docker/fui/arm: gopath/bin/linux_arm/fui
	cp gopath/bin/linux_arm/fui fui/
	cd fui && docker build -t btrfaas/fui:latest-arm -f Dockerfile.arm .
	mkdir -p $@ && touch $@

.docker/fui/arm64: gopath/bin/linux_arm64/fui
	cp gopath/bin/linux_arm64/fui fui/
	cd fui && docker build -t btrfaas/fui:latest-arm64 -f Dockerfile.arm64 .
	mkdir -p $@ && touch $@

####################################
#           LINTING                #
####################################
lint: vet fmt

vet:
	docker run \
		-v $(shell pwd):/go/src/github.com/trusch/btrfaas \
		golang:1.9 go vet github.com/trusch/btrfaas/...

fmt:
	docker run \
		-v $(shell pwd):/go/src/github.com/trusch/btrfaas \
		-w /go/src/github.com/trusch/btrfaas \
		-e CGO_ENABLED=0 \
		golang:1.9 gofmt -e -s -w $(SRC)
