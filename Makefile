export GO111MODULE ?= on

GIT_VERSION := $(shell git describe --always --tags)
BASE_PACKAGE_NAME := github.com/batx-dev/batproxy
DEFAULT_LDFLAGS := "-X $(BASE_PACKAGE_NAME).Version=$(GIT_VERSION)"
IMG := docker.io/ashertz/batproxy:$(GIT_VERSION)

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

.PHONE: build
build: fmt vet
	go build -ldflags=$(DEFAULT_LDFLAGS) -o bin/batproxy ./cmd

.PHONE: run
run: fmt vet
	go run ./cmd run \
		--dsn .batproxy/batproxy.db

# Build the docker image
.PHONE: docker-build
docker-build:
	docker build --platform linux/amd64 . -t ${IMG}

.PHONE: docker-push
# Push the docker image
docker-push:
	docker push ${IMG}

