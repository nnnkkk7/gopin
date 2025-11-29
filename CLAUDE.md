# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**gopin** is a CLI tool that pins versions of `go install` commands in files. It automatically converts `@latest` to specific semantic versions (e.g., `@v1.64.8`) to ensure reproducible builds and enhanced security. Think of it as [pinact](https://github.com/suzuki-shunsuke/pinact) but for Go tools instead of GitHub Actions.

## Development Commands

### Building
```bash
go build -o gopin cmd/gopin/main.go
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run integration tests manually
./gopin run --dry-run testdata/Makefile
```

### Running the CLI
```bash
# Build and run
go run cmd/gopin/main.go run

# Or after building
./gopin run --dry-run
./gopin check
./gopin list
```

## Architecture

### Core Pipeline

The tool follows a clear three-stage pipeline:

1. **Detection** ([pkg/detector](pkg/detector/)) - Scans files for `go install` patterns using regex
2. **Resolution** ([pkg/resolver](pkg/resolver/)) - Queries for latest versions via proxy.golang.org or `go list`
3. **Rewriting** ([pkg/rewriter](pkg/rewriter/)) - Replaces version strings in-place

### Package Responsibilities

- **[cmd/gopin](cmd/gopin/)** - Entry point only, delegates to CLI package
- **[pkg/cli](pkg/cli/)** - Command-line interface and orchestration (run, check, list, init commands)
- **[pkg/config](pkg/config/)** - Configuration loading from `.gopin.yaml` with defaults
- **[pkg/detector](pkg/detector/)** - Pattern matching for `go install` commands with include/exclude filters
- **[pkg/resolver](pkg/resolver/)** - Version resolution with multiple strategies:
  - `ProxyResolver` - Uses proxy.golang.org (primary)
  - `GoListResolver` - Falls back to `go list -m` for private modules
  - `FallbackResolver` - Tries primary then fallback
  - `CachedResolver` - Caches results per module path
- **[pkg/rewriter](pkg/rewriter/)** - In-place string replacement with change tracking

### Key Design Patterns

**Module Path Normalization**: The detector extracts root modules from subpaths (e.g., `github.com/golangci/golangci-lint/cmd/golangci-lint` → `github.com/golangci/golangci-lint`) in [detector.ExtractRootModule()](pkg/detector/detector.go:147-178). This is critical because the Go module proxy only recognizes root modules.

**Resolver Chain**: Resolvers are composed using decorators. The typical chain is: `CachedResolver` → `FallbackResolver` → `ProxyResolver` / `GoListResolver`. This pattern is constructed in [cli.createResolver()](pkg/cli/app.go:374-389).

**Backward Processing**: The rewriter processes matches in reverse order (last line first, rightmost column first) in [rewriter.Rewrite()](pkg/rewriter/rewriter.go:51-118). This prevents offset shifts when modifying earlier parts of the file.

**Version Pinning Behavior**: By default, gopin updates all `go install` commands to the latest version, regardless of whether they use `@latest` or are already pinned. This is controlled by [detector.NeedsPin()](pkg/detector/detector.go:137-141), which currently returns `true` for all versions.

## Configuration

Default target files (if no `.gopin.yaml` exists):
- `.github/**/*.yml`, `.github/**/*.yaml`
- `Makefile`, `makefile`, `GNUmakefile`
- `*.mk`

Configuration is loaded via [config.Load()](pkg/config/config.go:73-87) which searches standard paths:
- `.gopin.yaml`
- `.gopin.yml`
- `.github/gopin.yaml`
- `.github/gopin.yml`

## Version String

The project uses two version tracking mechanisms:

1. **Runtime version**: The `cli.Version` variable ([pkg/cli/app.go:20](pkg/cli/app.go:20)) is injected at build time using ldflags:
   ```bash
   go build -ldflags "-X github.com/nnnkkk7/gopin/pkg/cli.Version=v1.0.0"
   ```

2. **Source version**: The [version.go](version.go) file at the repository root contains a version constant used by tagpr for automated release management. This file is automatically updated by tagpr when creating release pull requests.

## Testing Approach

Tests use table-driven patterns with golden files in `testdata/`. The detector tests validate regex matching, resolver tests can be mocked to avoid external dependencies, and rewriter tests check change tracking accuracy.
