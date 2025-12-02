# gopin

[![CI](https://github.com/nnnkkk7/gopin/actions/workflows/ci.yml/badge.svg)](https://github.com/nnnkkk7/gopin/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/nnnkkk7/gopin)](https://goreportcard.com/report/github.com/nnnkkk7/gopin)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**gopin** is a CLI tool to pin versions of `go install` commands in your files for reproducible builds and enhanced security.

## Motivation

Using `@latest` in `go install` commands is convenient but problematic:

- **No reproducibility**: Different runs may install different versions
- **Security risk**: A malicious version could be installed unknowingly
- **CI/CD instability**: Team members might use different tool versions
- **Debugging difficulty**: Hard to reproduce past builds

gopin solves these problems by automatically updating all `go install` commands to the latest specific semantic versions:

- **Pin `@latest`**: Convert `@latest` to specific versions (e.g., `@v2.6.2`)
- **Update outdated versions**: Update already-pinned versions to the latest (e.g., `@v1.0.0` → `@v2.6.2`)
- **Add missing versions**: Add version specifiers to commands without them

## Installation

### Go Install

```bash
go install github.com/nnnkkk7/gopin/cmd/gopin@latest
```

### Verify Installation

```bash
gopin version
```

### Homebrew (macOS/Linux)

```bash
brew install nnnkkk7/tap/gopin
```

### Binary Download

Download the pre-built binary for your platform from [Releases](https://github.com/nnnkkk7/gopin/releases):

```bash
# macOS (Apple Silicon)
curl -L https://github.com/nnnkkk7/gopin/releases/latest/download/gopin_Darwin_arm64.tar.gz | tar xz
sudo mv gopin /usr/local/bin/

# macOS (Intel)
curl -L https://github.com/nnnkkk7/gopin/releases/latest/download/gopin_Darwin_x86_64.tar.gz | tar xz
sudo mv gopin /usr/local/bin/

# Linux (amd64)
curl -L https://github.com/nnnkkk7/gopin/releases/latest/download/gopin_Linux_x86_64.tar.gz | tar xz
sudo mv gopin /usr/local/bin/

# Linux (arm64)
curl -L https://github.com/nnnkkk7/gopin/releases/latest/download/gopin_Linux_arm64.tar.gz | tar xz
sudo mv gopin /usr/local/bin/
```

### macOS Security Note

On macOS, you may see a security warning: **"gopin cannot be opened because Apple cannot check it for malicious software"**

This happens because the binary is not code-signed. To resolve this:

#### Option 1: Remove quarantine attribute (Recommended)

```bash
xattr -d com.apple.quarantine $(which gopin)
```

#### Option 2: Allow in System Settings

1. Try to run `gopin` and the warning will appear
2. Open **System Settings** → **Privacy & Security**
3. Scroll down and click **"Open Anyway"** next to the gopin warning
4. Confirm by clicking **"Open"**

#### Option 3: Use Go Install (No security warnings)

```bash
go install github.com/nnnkkk7/gopin/cmd/gopin@latest
```

## Quick Start

```bash
# Pin all @latest versions in default target files
gopin run

# Preview changes without applying
gopin run --dry-run

# Check for unpinned versions (useful for CI)
gopin check

# List all go install commands
gopin list
```

## Usage

### `gopin run`

Update all go install commands to latest versions.

```bash
# Update all go install to latest
gopin run

# Update specific files
gopin run Makefile .github/workflows/*.yml

# Preview changes without applying
gopin run --dry-run

# Show diff output
gopin run --diff

# Update only specific modules
gopin run --include "golangci-lint.*"

# Exclude specific modules
gopin run --exclude "internal/.*"
```

**Note:** To keep a specific version, add it to `ignore_modules` in `.gopin.yaml`:

```yaml
ignore_modules:
  - name: "github.com/special/tool"
    reason: "Must stay at v1.50.0"
```

### `gopin check`

Check for unpinned go install commands. Exits with code 1 if unpinned versions are found.

```bash
# Check for unpinned versions
gopin check

# Fix unpinned versions
gopin check --fix
```

### `gopin list`

List go install commands in files.

```bash
# List all go install commands
gopin list

# List only unpinned commands
gopin list --unpinned
```

### `gopin init`

Create a configuration file.

```bash
gopin init
```

## Configuration

Create `.gopin.yaml` in your repository root:

```yaml
version: 1

# Target file patterns
files:
  - pattern: ".github/**/*.yml"
  - pattern: ".github/**/*.yaml"
  - pattern: "Makefile"
  - pattern: "makefile"
  - pattern: "GNUmakefile"
  - pattern: "*.mk"

# Modules to ignore (optional)
ignore_modules:
  - name: "github.com/internal/.*"
    reason: "internal tools"
```

### Default Target Files

If no configuration file is found, gopin uses these default patterns:

- `.github/**/*.yml` (all YAML files in .github directory, recursively)
- `.github/**/*.yaml`
- `Makefile`
- `makefile`
- `GNUmakefile`
- `*.mk`

## Examples

### Example 1: Pin `@latest` to specific version

**Before:**

```makefile
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
```

**After:**

```makefile
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.6.2
```

### Example 2: Update outdated pinned version

**Before:**

```makefile
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.0.0
```

**After:**

```makefile
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.6.2
```

### Example 3: Add version to unversioned command

**Before:**

```makefile
go install github.com/air-verse/air
```

**After:**

```makefile
go install github.com/air-verse/air@v1.61.7
```

## CI Integration

### GitHub Actions

```yaml
name: gopin

on:
  pull_request:

jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
      - run: go install github.com/nnnkkk7/gopin/cmd/gopin@latest
      - run: gopin check
```

### Auto-fix in PR

```yaml
name: gopin

on:
  pull_request:

permissions:
  contents: write

jobs:
  pin:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          ref: ${{ github.head_ref }}
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
      - run: go install github.com/nnnkkk7/gopin/cmd/gopin@latest
      - run: gopin run
      - name: Commit changes
        run: |
          git config user.name "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"
          git add -A
          git diff --cached --quiet || git commit -m "chore: pin go install versions"
          git push
```

## How It Works

1. **Scan files**: Find files matching the configured patterns
2. **Detect patterns**: Find `go install <module>@<version>` using regex
3. **Resolve versions**: Query `proxy.golang.org` (or `go list`) for latest version
4. **Rewrite files**: Update all versions to the latest (e.g., `@latest` → `@v2.6.2`, `@v1.0.0` → `@v2.6.2`)

### Version Resolution

gopin uses the Go module proxy (`proxy.golang.org`) to resolve module versions:

```
https://proxy.golang.org/<module>/@latest
```

For private modules or environments without proxy access, gopin falls back to `go list -m -json <module>@latest`.

## Comparison with Alternatives

| Tool | Scope | Pros | Cons |
|------|-------|------|------|
| **gopin** | go install in any file | Automated, works with existing code | New tool to learn |
| **Go 1.24 tool directive** | go.mod only | Official feature | Limited to go.mod |
| **aqua** | Multi-language | Unified version management | Additional setup |
| **Manual** | Any | Simple | Time-consuming |

## Development

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

# Run integration tests
./gopin run --dry-run testdata/Makefile
```

### Project Structure

```
gopin/
├── cmd/gopin/           # Main entry point
├── pkg/
│   ├── detector/        # Pattern detection
│   ├── resolver/        # Version resolution
│   ├── rewriter/        # File rewriting
│   ├── config/          # Configuration
│   └── cli/             # CLI commands
├── testdata/            # Test fixtures
└── README.md
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

[MIT](LICENSE)
