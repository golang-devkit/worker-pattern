# Makefile for Application
.DEFAULT_GOAL := help
.PHONY: build build-docker

# ====================================================================================
# Variables
# ====================================================================================
SHELL = /usr/bin/env bash


# Application settings
PACKAGE_NAME := "github.com/golang-devkit/worker-pattern"
MAIN_GO_FILE ?= cmd/v1/*.go

# Build artifacts
BIN_DIR := $(shell pwd)/bin
BINARY  ?= $(BIN_DIR)/app

# Build-time variables
# Use ?= to replace with command line argument (exp: make build VERSION=1.2.3)
VERSION      ?= $(shell git describe --tags --always --dirty)
COMMIT_HASH  := $(shell git rev-parse HEAD)
BRANCH       := $(shell git rev-parse --abbrev-ref HEAD)
BUILD_DATE   := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
BUILD_USER   := $(shell whoami)@$(shell hostname)
REPO_URL     := $(shell git remote get-url origin 2>/dev/null || echo \
	"https://github.com/golang-devkit/worker-pattern")

# Go build flags
LDFLAGS := -ldflags="\
	-s -w \
	-X '$(PACKAGE_NAME)/pkg.Version=$(VERSION)' \
	-X '$(PACKAGE_NAME)/pkg.BuildCommit=$(COMMIT_HASH)' \
	-X '$(PACKAGE_NAME)/pkg.BuildBranch=$(BRANCH)' \
	-X '$(PACKAGE_NAME)/pkg.BuildDate=$(BUILD_DATE)' \
	-X '$(PACKAGE_NAME)/pkg.BuildUser=$(BUILD_USER)' \
	-X '$(PACKAGE_NAME)/pkg.RepoURL=$(REPO_URL)' \
"

# Build application binary
build:
	@echo "==> Building application..."
	@mkdir -p $(BIN_DIR)
	@export CGO_ENABLED=0; \
		go build -mod=vendor -v $(LDFLAGS) -o $(BINARY) -trimpath $(MAIN_GO_FILE)
	@echo "==> Build successful: $(BINARY)"

# Run the application (sẽ build nếu cần)
run: build
	@echo "==> Running application..."
	@$(BINARY)

# Run all tests với race detector
test:
	@echo "==> Running tests..."
	@go test -v -race ./...

# Clean up build artifacts
clean:
	@echo "==> Cleaning up..."
	@rm -rf $(BIN_DIR)

fetch-module: init
	@echo "==> Create vendor directory..."
	@go mod vendor && \
		echo "==> Fetch Go module completed!"

build-linux:
	@echo "==> Building application... (linux/amd64)"
	@mkdir -p $(BIN_DIR)
	@export GOOS=linux GOARCH=amd64 CGO_ENABLED=0; \
		go build -mod=vendor -v $(LDFLAGS) -o $(BINARY)_linux_amd64 -trimpath $(MAIN_GO_FILE)
	@echo "==> Build successful: $(BINARY)_linux_amd64"

# Build the Docker image (placeholder)
build-docker:
	@echo "Building Docker image... (not implemented yet)"

# Build stress test tool
stresstest:
	@echo "==> Building stress test tool..."
	@mkdir -p $(BIN_DIR)
	@go build -mod=vendor -v -o $(BIN_DIR)/stresstest -trimpath ./cmd/stresstest

# Run stress test (requires running server)
stress: stresstest
	@echo "==> Running stress test against http://27.71.229.15:3000/heavy"
	@$(BIN_DIR)/stresstest \
		-concurrency=80 \
		-rampup=80s \
		-requests=1000 \
		-timeout=30s

# Defined help command
help:
	@echo "Available commands:"
	@echo "  make build        : Build the application binary."
	@echo "  make run          : Build and run the application."
	@echo "  make test         : Run all unit tests with race detector."
	@echo "  make clean        : Remove build artifacts."
	@echo "  make build-linux  : Build for Linux (amd64)."
	@echo "  make stresstest   : Build the stress test tool."
	@echo "  make stress       : Build and run stress test tool."
	@echo "  make build-docker : Build the Docker image (TBD)."
	@echo "  make help         : Show this help message."
