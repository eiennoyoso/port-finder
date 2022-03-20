SHELL=bash

GOPATH=$(shell go env GOPATH)
BINARY_NAME=port-finder

default: build

deps:
	go get -v -t -d ./...

build: deps
	CGO_ENABLED=0 go build -v -x -a $(LDFLAGS) -o $(CURDIR)/bin/$(BINARY_NAME)
	chmod +x $(CURDIR)/bin/$(BINARY_NAME)

clean:
	rm -rf $(CURDIR)/bin/*
	go clean -i -cache

# Install binary locally
install:
	cp $(CURDIR)/bin/$(BINARY_NAME) /usr/local/bin

# build docker image from latest github tag
docker-build:
	docker build \
		--tag sokil/port-finder:latest \
		-f ./Dockerfile .

docker-publish:
	docker push sokil/port-finder:latest