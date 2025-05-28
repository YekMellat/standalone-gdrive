# Makefile for standalone-gdrive

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOVET=$(GOCMD) vet
BINARY_NAME=gdrive
BINARY_UNIX=$(BINARY_NAME)
BINARY_WIN=$(BINARY_NAME).exe

# Main target
all: test build

# Build the main binary
build:
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/gdrive

# Build examples
build-examples:
	$(GOBUILD) -o examples/list_files/list_files examples/list_files/main.go
	$(GOBUILD) -o examples/upload_file/upload_file examples/upload_file/main.go

# Run tests (skipping integration tests)
test:
	$(GOTEST) -v -short ./...

# Run all tests including integration tests (requires Google Drive access)
test-all:
	TEST_GDRIVE_ACCESS=1 $(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -cover ./...

# Generate coverage report
coverage-html:
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Run benchmarks
bench:
	$(GOTEST) -bench=. -short ./...

# Run benchmarks with memory allocation stats
bench-mem:
	$(GOTEST) -bench=. -benchmem -short ./...

# Run the OAuth flow test tool
oauth-test:
	$(GOBUILD) -o oauth-test -v ./cmd/oauth_test
	./oauth-test

# Clean up
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
	rm -f $(BINARY_WIN)
	rm -f examples/list_files/list_files
	rm -f examples/upload_file/upload_file

# Update dependencies
deps:
	$(GOMOD) tidy

# Check for issues
lint:
	$(GOVET) ./...

# Cross compilation
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v ./cmd/gdrive

build-windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BINARY_WIN) -v ./cmd/gdrive

# Build all platforms
build-all: build-linux build-windows

# Install the binary
install:
	$(GOBUILD) -o $(GOPATH)/bin/$(BINARY_NAME) -v ./cmd/gdrive

.PHONY: all build test test-all test-coverage coverage-html bench bench-mem oauth-test clean deps lint build-linux build-windows build-all install build-examples
