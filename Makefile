export GO111MODULE=on
BINARY_NAME=ranchhand

.PHONY: test

all: test build

default: build

test:
	@go test -v ./...

build:
	@go build -v -o $(BINARY_NAME)

clean:
	@go clean
