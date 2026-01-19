.PHONY: build install clean test lint fmt vet tidy release snapshot help

# Build variables
BINARY_NAME := aloc
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS := -s -w \
	-X main.version=$(VERSION) \
	-X main.commit=$(COMMIT) \
	-X main.buildTime=$(BUILD_TIME)

# Go commands
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOVET := $(GOCMD) vet
GOFMT := gofmt
GOMOD := $(GOCMD) mod

# Default target
all: build

## build: Build the binary for current platform
build:
	CGO_ENABLED=0 $(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BINARY_NAME) ./cmd/aloc

## install: Install to GOPATH/bin
install:
	CGO_ENABLED=0 $(GOCMD) install -ldflags="$(LDFLAGS)" ./cmd/aloc

## clean: Remove build artifacts
clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME)-*
	rm -rf dist/

## test: Run tests
test:
	$(GOTEST) -v -race -cover ./...

## test-short: Run tests without race detector (faster)
test-short:
	$(GOTEST) -v -cover ./...

## lint: Run golangci-lint
lint:
	golangci-lint run ./...

## fmt: Format code
fmt:
	$(GOFMT) -s -w .

## vet: Run go vet
vet:
	$(GOVET) ./...

## tidy: Tidy go.mod
tidy:
	$(GOMOD) tidy

## check: Run all checks (fmt, vet, lint, test)
check: fmt vet lint test

## release: Create a release using GoReleaser
release:
	goreleaser release --clean

## snapshot: Build snapshot release (no publish)
snapshot:
	goreleaser release --snapshot --clean

## build-all: Build for all platforms locally
build-all:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -ldflags="$(LDFLAGS)" -o dist/$(BINARY_NAME)-darwin-amd64 ./cmd/aloc
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -ldflags="$(LDFLAGS)" -o dist/$(BINARY_NAME)-darwin-arm64 ./cmd/aloc
	GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags="$(LDFLAGS)" -o dist/$(BINARY_NAME)-linux-amd64 ./cmd/aloc
	GOOS=linux GOARCH=arm64 $(GOBUILD) -ldflags="$(LDFLAGS)" -o dist/$(BINARY_NAME)-linux-arm64 ./cmd/aloc
	GOOS=windows GOARCH=amd64 $(GOBUILD) -ldflags="$(LDFLAGS)" -o dist/$(BINARY_NAME)-windows-amd64.exe ./cmd/aloc

## help: Show this help
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed 's/^/ /'
