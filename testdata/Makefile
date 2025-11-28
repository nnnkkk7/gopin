# Makefile with various go install patterns

.PHONY: lint
lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8
	golangci-lint run

.PHONY: generate
generate:
	go install golang.org/x/tools/cmd/stringer@v0.39.0
	go generate ./...

# Multiple installs in one target
.PHONY: install-tools
install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest

# With flags
.PHONY: install-goreleaser
install-goreleaser:
	go install -v github.com/goreleaser/goreleaser@latest

# Chained commands
.PHONY: setup-dev
setup-dev:
	go install github.com/cosmtrek/air@latest && air -v
	go install github.com/golang/mock/mockgen@latest && mockgen --version

# Mixed pinned and unpinned
.PHONY: setup-ci
setup-ci:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/stringer@v0.27.0
	go install github.com/securego/gosec/v2/cmd/gosec@v2.21.4

# With continuation backslash
.PHONY: install-all
install-all:
	go install \
		github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install \
		golang.org/x/tools/cmd/goimports@latest

# Variable usage
GOLANGCI_VERSION := latest
.PHONY: lint-var
lint-var:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_VERSION)

# With inline comments
.PHONY: security
security:
	go install github.com/securego/gosec/v2/cmd/gosec@latest # Security scanner
	gosec ./...

# No version specified
.PHONY: install-air
install-air:
	go install github.com/air-verse/air

# Conditional install
.PHONY: install-conditional
install-conditional:
ifeq ($(ENABLE_LINT),true)
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
endif

# Output redirect
.PHONY: install-quiet
install-quiet:
	@go install github.com/swaggo/swag/cmd/swag@latest > /dev/null 2>&1

# Multi-line command
.PHONY: test
test:
	go install gotest.tools/gotestsum@latest && \
		gotestsum --format testname
