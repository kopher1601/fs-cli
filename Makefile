BINARY_NAME=fs-cli
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION)"

.PHONY: build run test lint clean

build:
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/fs-cli/

run:
	go run ./cmd/fs-cli/

test:
	go test ./... -v

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/

install: build
	cp bin/$(BINARY_NAME) $(GOPATH)/bin/
