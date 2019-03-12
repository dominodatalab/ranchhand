export GO111MODULE=on
BINARY_NAME=ranchhand

default: build

build:
	@go build -v -o $(BINARY_NAME)
