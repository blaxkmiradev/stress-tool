# Makefile

.PHONY: build clean run test deps help

BINARY=layer7-flood
VERSION=1.0.0
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

help:
	@echo "Available targets:"
	@echo "  make build   - Build binary for current OS"
	@echo "  make deps    - Download dependencies"
	@echo "  make run     - Run with default config"
	@echo "  make test    - Run tests"
	@echo "  make clean   - Clean build artifacts"
	@echo "  make all     - Build all platforms"

deps:
	go mod download
	go mod tidy

build:
	CGO_ENABLED=0 go build -ldflags="-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)" -o bin/$(BINARY) cmd/ddos/main.go

all: deps
	mkdir -p bin
	GOOS=linux GOARCH=amd64 go build -o bin/$(BINARY)-linux-amd64 cmd/ddos/main.go
	GOOS=darwin GOARCH=amd64 go build -o bin/$(BINARY)-darwin-amd64 cmd/ddos/main.go
	GOOS=windows GOARCH=amd64 go build -o bin/$(BINARY)-windows-amd64.exe cmd/ddos/main.go
	@echo "Build complete for all platforms"

run: build
	./bin/$(BINARY) -config configs/default.yaml

test:
	go test -v ./...

clean:
	rm -rf bin/
	go clean -cache

install: build
	sudo cp bin/$(BINARY) /usr/local/bin/

stress-local: build
	./bin/$(BINARY) -target http://localhost:8080 -threads 200 -duration 60 -attack mixed

stress-proxy: build
	./bin/$(BINARY) -config configs/advanced.yaml
