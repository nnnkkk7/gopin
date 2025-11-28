#!/bin/bash
set -euo pipefail

# Install development tools
echo "Installing development tools..."

# Basic tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install golang.org/x/tools/cmd/goimports@latest

# Code generation tools
go install golang.org/x/tools/cmd/stringer@latest
go install github.com/golang/mock/mockgen@latest

# Testing tools
go install gotest.tools/gotestsum@latest
go install github.com/rakyll/gotest@latest

# Security tools
go install github.com/securego/gosec/v2/cmd/gosec@latest
go install golang.org/x/vuln/cmd/govulncheck@latest

echo "All tools installed successfully!"
