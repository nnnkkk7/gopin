#!/bin/bash

# CI script with various patterns

# Single line installs
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# With flags
go install -v github.com/goreleaser/goreleaser@latest

# Conditional install
if [ "$ENABLE_LINT" = "true" ]; then
  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
fi

# In function
install_tools() {
  go install golang.org/x/tools/cmd/goimports@latest
  go install honnef.co/go/tools/cmd/staticcheck@latest
}

# With variable
TOOL_VERSION="latest"
go install github.com/cosmtrek/air@${TOOL_VERSION}

# Chained with &&
go install github.com/spf13/cobra-cli@latest && cobra-cli init

# With stderr redirect
go install github.com/swaggo/swag/cmd/swag@latest 2>&1

# Already pinned
go install golang.org/x/tools/cmd/stringer@v0.27.0
