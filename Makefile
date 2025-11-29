# Makefile for gopin
.PHONY: help build test lint lint-fix clean install

# Default target
.DEFAULT_GOAL := help

# Variables
BINARY_NAME := gopin
BUILD_DIR := dist
GO := go
GOLANGCI_LINT := golangci-lint
GOLANGCI_LINT_VERSION := v2.6.2

## help: Display this help message
help:
	@echo "Available targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /' | column -t -s ':'

## build: Build the binary
build:
	$(GO) build -o $(BINARY_NAME) cmd/gopin/main.go

## test: Run tests
test:
	$(GO) test -v ./...

## test-cover: Run tests with coverage
test-cover:
	$(GO) test -cover ./...

## lint: Run golangci-lint (install if needed)
lint: install-lint-deps
	$(GOLANGCI_LINT) run

## lint-fix: Run golangci-lint with auto-fix (install if needed)
lint-fix: install-lint-deps
	$(GOLANGCI_LINT) run --fix

## install-lint-deps: Install golangci-lint if not present
install-lint-deps:
	@if ! command -v $(GOLANGCI_LINT) &> /dev/null; then \
		echo "golangci-lint not found. Installing $(GOLANGCI_LINT_VERSION)..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $$(go env GOPATH)/bin $(GOLANGCI_LINT_VERSION); \
	fi

## install: Install the binary to $GOPATH/bin
install:
	$(GO) install ./cmd/gopin

## clean: Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -rf $(BUILD_DIR)

## ci: Run CI checks (lint + test)
ci: lint test

## all: Build, test, and lint
all: lint test build
