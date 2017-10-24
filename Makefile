SRC=$(shell find ./callable ./cmd ./env ./http -type f -name "*.go")

all: docker

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

docker: cmd/frunner/frunner cmd/frunner/Dockerfile
	cd cmd/frunner && docker build -t trusch/frunner .

clean:
	rm -rf frunner vendor
