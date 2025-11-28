# Test Data

This directory contains various test files to verify gopin's functionality across different file formats and edge cases.

## File Structure

```
testdata/
├── .github/
│   └── workflows/
│       ├── ci.yml              # Basic GitHub Actions workflow
│       ├── advanced.yml        # Advanced patterns (multi-line, flags, etc.)
│       └── complex.yaml        # Complex scenarios (matrix, conditionals)
├── scripts/
│   ├── setup.sh               # Basic shell script with multiple installs
│   ├── ci.sh                  # CI script with various patterns
│   └── install-tools.sh       # Edge cases (backslash, subshell, arrays)
├── Dockerfile                 # Docker build with go install
├── Dockerfile.dev             # Development Dockerfile
├── Makefile                   # Makefile with comprehensive patterns
├── Makefile.backup            # Original Makefile (unpinned)
├── justfile                   # Just command runner
├── taskfile.yml               # Task runner configuration
├── edge-cases.txt             # Edge cases and special patterns
└── .gopin.yaml                # gopin configuration file

## Test Patterns Covered

### 1. Basic Patterns
- `go install module@latest`
- `go install module@v1.2.3` (already pinned)
- `go install module` (no version, implicitly latest)

### 2. With Flags
- `go install -v module@latest`
- `go install -trimpath module@latest`
- `go install -tags=extended module@latest`
- `go install -v -trimpath -ldflags="-s -w" module@latest`

### 3. Multi-line Commands
- Shell script with backslash continuation
- Dockerfile RUN with multiple commands
- Makefile with multi-line targets
- YAML multi-line strings

### 4. Chained Commands
- `go install module@latest && tool --version`
- `go install module@latest; tool run`
- Multiple go install commands in one line

### 5. Module Path Variations
- Standard: `github.com/org/repo@latest`
- With subpackage: `github.com/org/repo/cmd/tool@latest`
- With version suffix: `github.com/org/repo/v2/cmd/tool@latest`
- gopkg.in: `gopkg.in/yaml.v3@latest`
- k8s.io: `k8s.io/kubectl@latest`
- Numbers in path: `github.com/99designs/gqlgen@latest`
- Hyphens and underscores: `github.com/my-org/my_tool@latest`

### 6. Version Format Variations
- `@latest` - Latest stable version
- `@v1.2.3` - Specific semantic version
- `@v1.2.3-rc1` - Pre-release version
- `@master` or `@main` - Branch reference
- `@v0.0.0-20230101120000-abcdef123456` - Pseudo-version

### 7. Edge Cases
- Different whitespace (spaces, tabs, multiple spaces)
- Inline comments: `go install module@latest # comment`
- Already pinned versions (will be updated to latest)
- Mixed pinned and unpinned in same file
- Variable usage in Makefile: `@$(VERSION)`
- Conditional installs (if statements, Make conditionals)
- Output redirects: `> /dev/null 2>&1`
- Subshells and functions
- Arrays and loops

### 8. File Format Coverage
- **GitHub Actions**: `.yml`, `.yaml` files
- **Shell Scripts**: `.sh` files (bash, sh)
- **Makefiles**: `Makefile`, `*.mk`
- **Dockerfiles**: `Dockerfile`, `Dockerfile.*`
- **Just**: `justfile`
- **Task**: `taskfile.yml`
- **Plain text**: `.txt` files

## Testing Strategy

### Unit Tests
Each package has unit tests that use simple strings and patterns from these files.

### Integration Tests
The testdata files can be used for end-to-end testing:

1. **Detection Test**: Verify all go install patterns are detected
2. **Pinning Test**: Verify @latest is replaced with specific versions
3. **Update Test**: Verify all versions are updated to latest
4. **Format Preservation**: Verify file formatting is preserved (indentation, comments, etc.)

### Coverage Goals
- ✅ Basic go install patterns
- ✅ Flags and options
- ✅ Multi-line commands
- ✅ Various module path formats
- ✅ Different version formats
- ✅ Multiple file formats
- ✅ Edge cases (whitespace, comments, conditionals)
- ✅ Already pinned versions
- ✅ Mixed pinned/unpinned

## Usage

### Manual Testing
```bash
# Run gopin on testdata
gopin run testdata/Makefile

# Check for unpinned versions
gopin check testdata/

# List all go install commands
gopin list testdata/
```

### Automated Testing
```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./pkg/detector -v
go test ./pkg/rewriter -v
```

## Adding New Test Cases

When adding new test cases:

1. **Add to appropriate file** or create new file if needed
2. **Document the pattern** in this README
3. **Add corresponding unit test** if it's a new edge case
4. **Verify detection** works correctly
5. **Verify rewriting** preserves format and intent

## Notes

- **Makefile.backup**: Contains unpinned versions as reference
- **edge-cases.txt**: Comprehensive list of edge cases for reference
- **.gopin.yaml**: Example configuration file (no output config section)
