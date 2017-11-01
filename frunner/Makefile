SRC=$(shell find ./cmd ./config ./env ./framer ./http ./runnable ./grpc -type f -name "*.go") grpc/frunner.pb.go

all: fmt vet test cmd/frunner/frunner docker

install: cmd/frunner/frunner
	cp cmd/frunner/frunner $(GOPATH)/bin/

cmd/frunner/frunner: $(SRC) vendor
	docker run \
		-v $(shell pwd):/go/src/github.com/trusch/frunner \
		-w /go/src/github.com/trusch/frunner/cmd/frunner \
		-u $(shell stat -c '%u:%g' .) \
		-e CGO_ENABLED=0 \
		-e GOOS=linux \
		golang:1.9 \
			go build -v -a -ldflags '-extldflags "-static"' .

vendor: glide.yaml
	docker run \
		-v $(shell pwd):/go/src/github.com/trusch/frunner \
		-w /go/src/github.com/trusch/frunner \
		-u $(shell stat -c '%u:%g' .) \
		-e CGO_ENABLED=0 \
		-e GOOS=linux \
		golang:1.9 bash -c \
			"(curl https://glide.sh/get | sh) && glide --home /tmp update"

test: vendor
	docker run \
		-v $(shell pwd):/go/src/github.com/trusch/frunner \
		-w /go/src/github.com/trusch/frunner \
		-u $(shell stat -c '%u:%g' .) \
		-e CGO_ENABLED=0 \
		-e GOOS=linux \
		golang:1.9 \
			go test -v -cover ./...

vet: vendor
	docker run \
		-v $(shell pwd):/go/src/github.com/trusch/frunner \
		-u $(shell stat -c '%u:%g' .) \
		-e CGO_ENABLED=0 \
		-e GOOS=linux \
		golang:1.9 \
			go vet github.com/trusch/frunner/...

fmt: vendor
	docker run \
		-v $(shell pwd):/go/src/github.com/trusch/frunner \
		-u $(shell stat -c '%u:%g' .) \
		-e CGO_ENABLED=0 \
		-e GOOS=linux \
		golang:1.9 \
			go fmt github.com/trusch/frunner/...


docker: cmd/frunner/frunner cmd/frunner/Dockerfile
	cd cmd/frunner && docker build -t trusch/frunner .

grpc/frunner.pb.go: grpc/frunner.proto
	cd grpc && protoc --go_out=plugins=grpc:. frunner.proto

clean:
	rm -rf frunner vendor
