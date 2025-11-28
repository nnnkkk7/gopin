# Test Data Summary

## Overview
This testdata directory contains comprehensive test cases for gopin across various file formats and edge cases.

## Statistics

### Files by Type
- **GitHub Actions Workflows**: 3 files (.github/**/*.yml, *.yaml)
- **Shell Scripts**: 3 files (scripts/*.sh)
- **Dockerfiles**: 2 files (Dockerfile, Dockerfile.dev)
- **Makefiles**: 2 files (Makefile, Makefile.backup)
- **Task Runners**: 2 files (justfile, taskfile.yml)
- **Test Cases**: 1 file (edge-cases.txt)
- **Documentation**: 2 files (README.md, SUMMARY.md)
- **Configuration**: 1 file (.gopin.yaml)

**Total Test Files**: 14

### Test Coverage

#### File Formats
- ✅ GitHub Actions YAML (.yml, .yaml)
- ✅ Shell Scripts (.sh)
- ✅ Dockerfiles
- ✅ Makefiles
- ✅ Justfile
- ✅ Taskfile
- ✅ Plain text (.txt)

#### Go Install Patterns
- ✅ `go install module@latest`
- ✅ `go install module@v1.2.3` (pinned)
- ✅ `go install module` (no version)
- ✅ `go install -v module@latest` (with flags)
- ✅ `go install -tags=extended module@latest`
- ✅ Multi-line commands (backslash continuation)
- ✅ Chained commands (&&, ;)
- ✅ Multiple installs in one RUN/step
- ✅ Conditional installs (if statements)
- ✅ Variable usage in Makefile
- ✅ Inline comments

#### Module Path Formats
- ✅ github.com/owner/repo
- ✅ github.com/owner/repo/cmd/tool
- ✅ github.com/owner/repo/v2/cmd/tool
- ✅ golang.org/x/tools
- ✅ gopkg.in/yaml.v3
- ✅ k8s.io/kubectl
- ✅ honnef.co/go/tools
- ✅ Paths with numbers (99designs)
- ✅ Paths with hyphens and underscores

#### Version Formats
- ✅ @latest
- ✅ @v1.2.3
- ✅ @v1.2.3-rc1
- ✅ @master, @main
- ✅ @v0.0.0-20230101120000-abcdef123456 (pseudo-version)
- ✅ (no version - implicitly latest)

#### Edge Cases
- ✅ Different whitespace (spaces, tabs)
- ✅ Multiple flags
- ✅ Inline comments (not pinned comments)
- ✅ Already pinned versions
- ✅ Mixed pinned and unpinned
- ✅ Subshells and functions
- ✅ Output redirects
- ✅ Arrays and loops

## Key Test Files

### Basic Examples
- **Makefile.backup**: Original unpinned Makefile for comparison
- **.github/workflows/ci.yml**: Simple CI workflow

### Advanced Patterns
- **Makefile**: Comprehensive Makefile with all patterns
- **.github/workflows/advanced.yml**: Multi-install, flags, mixed versions
- **.github/workflows/complex.yaml**: Matrix builds, conditionals

### Shell Scripts
- **scripts/setup.sh**: Basic installation script
- **scripts/ci.sh**: Various patterns (conditionals, functions, variables)
- **scripts/install-tools.sh**: Edge cases (backslash, subshell, arrays)

### Containers
- **Dockerfile**: Multi-stage build with go install
- **Dockerfile.dev**: Development environment setup

### Task Runners
- **justfile**: Just task runner examples
- **taskfile.yml**: Task runner with multiple targets

### Edge Cases
- **edge-cases.txt**: Comprehensive list of edge cases for reference

## Usage Examples

### List all go install commands
```bash
cd /path/to/gopin
./gopin list testdata/
```

### List only unpinned versions
```bash
./gopin list --unpinned testdata/
```

### Check for unpinned versions
```bash
./gopin check testdata/
```

### Pin all @latest versions (dry-run)
```bash
./gopin run --dry-run testdata/Makefile
```

### Pin all @latest versions in all files
```bash
./gopin run testdata/*.sh testdata/Makefile testdata/.github/workflows/*.yml
```

## Known False Positives

Some patterns may be incorrectly detected as go install commands:

1. **YAML keywords**: `commands@(none)`, `with@(none)` - These are YAML structure keywords
2. **Backslash escapes**: `\@(none)` - Escaped backslashes in examples
3. **Variable syntax**: Makefile variables like `@$(VERSION)` may be detected

These false positives are expected in test files that demonstrate various patterns and edge cases.

## Testing Approach

1. **Unit Tests**: Each package tests specific functionality
2. **Integration Tests**: Full end-to-end workflow testing
3. **Manual Testing**: Using testdata files for manual verification
4. **Regression Testing**: Ensuring changes don't break existing functionality

## Future Enhancements

Potential additions to testdata:

- [ ] More complex Dockerfile patterns (multi-stage builds)
- [ ] PowerShell scripts (.ps1)
- [ ] GitHub Actions composite actions
- [ ] More conditional patterns
- [ ] Error cases (malformed commands)
- [ ] Unicode and internationalization
- [ ] Very long module paths
- [ ] Nested tool installations
