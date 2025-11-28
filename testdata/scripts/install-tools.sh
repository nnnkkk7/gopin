#!/usr/bin/env bash
# Install tools script with edge cases

set -e

# Multi-line with backslash continuation
go install \
  github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# With multiple flags
go install -v -trimpath github.com/goreleaser/goreleaser@latest

# No version specified (implicitly latest)
go install github.com/air-verse/air

# With comments on the same line
go install golang.org/x/tools/cmd/goimports@latest # Import formatter

# In subshell
(
  cd /tmp
  go install github.com/cosmtrek/air@latest
)

# With output redirect
go install github.com/swaggo/swag/cmd/swag@latest > /dev/null

# Array of tools
TOOLS=(
  "github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
  "golang.org/x/tools/cmd/stringer@latest"
  "github.com/golang/mock/mockgen@latest"
)

for tool in "${TOOLS[@]}"; do
  go install "$tool"
done
