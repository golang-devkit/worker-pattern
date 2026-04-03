#!Makefile

# ====================================================================================
# Variables
# ====================================================================================
SHELL = /usr/bin/env bash

GO_TOOLCHAIN = $(shell go version | awk '{print $$3}')

clean-mod:
	@echo "==> Clean with flag -modcache"
	@go clean -modcache
	@echo "==> Done!"; \
		echo "Please excute 'make fetch-module' or 'go mod download' ..."

fetch-mod:
	@echo "==> Fetch Go module..."; \
		go mod tidy

init: clean-mod
	@echo "==> Remove go module exist..."; \
		rm -rf go.mod go.sum vendor/
	@echo "==> Initializing Go module..."; \
		go mod init github.com/golang-devkit/worker-pattern; \
		go mod edit -go=1.25.7; \
		go mod edit -toolchain $(GO_TOOLCHAIN)
		# Replace module if needed, for example:
		# go mod edit -replace=github.com/old/module=github.com/new/module
	@echo "==> Fetch Go module..."; \
		go mod tidy
	@echo "✅ Fetch Go module completed!"

upgrade-module:
	@echo "==> Upgrading required packages to latest version"; \
		go get -u ./...; \
		go mod tidy
	@echo "✅ Upgrade completed!"

upgrade-module-all:
	@echo "==> Upgrading required packages and all dependency to latest version"; \
		go get -u all; \
		go mod tidy
	@echo "✅ Upgrade completed!"

fetch-module: fetch-mod upgrade-module
	@echo "==> Create vendor directory..."; \
		go mod vendor && echo "✅ Fetch Go module completed!"
	@echo "==> Run govulncheck..."; \
		govulncheck \
			-show version \
			-C $(shell pwd) ./... #Please use flag "-show verbose" to show details
	@echo "✅ Successful!"

fetch-pkg-worker:
	@echo "==> Fetch Go module..."; \
		cd pkg/worker && go mod tidy; \
		echo "✅ Fetch Go module completed!"
	@echo "==> Building worker package..."; \
		cd pkg/worker && \
		go build -o /tmp/ -trimpath ./...
	@echo "✅ Build completed!"