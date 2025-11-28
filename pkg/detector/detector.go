// Package detector provides functionality to detect go install patterns in files.
package detector

import (
	"bufio"
	"regexp"
	"strings"
)

// GoInstallPattern is a regular expression to detect go install patterns
// Supported formats:
//   - go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
//   - go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.61.0
//   - go install github.com/golangci/golangci-lint/cmd/golangci-lint
//   - go install -v github.com/xxx/yyy@latest
//   - go install -tags=... github.com/xxx/yyy@latest
var GoInstallPattern = regexp.MustCompile(
	`go\s+install\s+` + // go install
		`(?:-[a-zA-Z0-9_=,]+\s+)*` + // Optional flags (-v, -tags=xxx, etc.)
		`([^\s@#]+)` + // Module path (group 1)
		`(?:@([^\s#]+))?`, // Version (group 2, optional)
)

// Match represents the result of a go install pattern match
type Match struct {
	FullMatch   string // The entire matched string
	ModulePath  string // Module path (e.g., github.com/golangci/golangci-lint/cmd/golangci-lint)
	Version     string // Version (e.g., latest, v1.61.0, or empty string)
	Line        int    // Line number (1-indexed)
	StartColumn int    // Start column (0-indexed)
	EndColumn   int    // End column (0-indexed)
}

// Detector detects go install patterns
type Detector struct {
	IncludePatterns []*regexp.Regexp
	ExcludePatterns []*regexp.Regexp
}

// NewDetector creates a new Detector
func NewDetector(includes, excludes []string) (*Detector, error) {
	d := &Detector{}

	for _, pattern := range includes {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, err
		}
		d.IncludePatterns = append(d.IncludePatterns, re)
	}

	for _, pattern := range excludes {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, err
		}
		d.ExcludePatterns = append(d.ExcludePatterns, re)
	}

	return d, nil
}

// Detect detects go install patterns from file content
func (d *Detector) Detect(content string) ([]Match, error) {
	var matches []Match

	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Detect all matches in this line
		allMatches := GoInstallPattern.FindAllStringSubmatchIndex(line, -1)
		for _, loc := range allMatches {
			if len(loc) < 4 {
				continue
			}

			fullMatch := line[loc[0]:loc[1]]
			modulePath := line[loc[2]:loc[3]]

			var version string
			if loc[4] != -1 && loc[5] != -1 {
				version = line[loc[4]:loc[5]]
			}

			// Apply include/exclude filters
			if !d.ShouldProcess(modulePath) {
				continue
			}

			matches = append(matches, Match{
				FullMatch:   fullMatch,
				ModulePath:  modulePath,
				Version:     version,
				Line:        lineNum,
				StartColumn: loc[0],
				EndColumn:   loc[1],
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return matches, nil
}

// ShouldProcess determines if a module path should be processed
func (d *Detector) ShouldProcess(modulePath string) bool {
	// Exclude if it matches exclude patterns
	for _, re := range d.ExcludePatterns {
		if re.MatchString(modulePath) {
			return false
		}
	}

	// If no include patterns are specified, all modules are targets
	if len(d.IncludePatterns) == 0 {
		return true
	}

	// Only process modules that match include patterns
	for _, re := range d.IncludePatterns {
		if re.MatchString(modulePath) {
			return true
		}
	}

	return false
}

// NeedsPin determines if a version needs to be pinned/updated
func NeedsPin(version string) bool {
	// All versions should be updated to latest
	// This includes @latest, @v1.0.0, @master, or no version
	return true
}

// ExtractRootModule extracts the root module from a module path
// Example: github.com/golangci/golangci-lint/cmd/golangci-lint
//
//	→ github.com/golangci/golangci-lint
func ExtractRootModule(modulePath string) string {
	parts := strings.Split(modulePath, "/")

	// Common patterns:
	// github.com/owner/repo/...
	// gitlab.com/owner/repo/...
	// bitbucket.org/owner/repo/...
	if len(parts) >= 3 {
		switch parts[0] {
		case "github.com", "gitlab.com", "bitbucket.org":
			return strings.Join(parts[:3], "/")
		case "golang.org":
			// golang.org/x/tools, etc.
			if len(parts) >= 3 && parts[1] == "x" {
				return strings.Join(parts[:3], "/")
			}
		case "google.golang.org":
			// google.golang.org/protobuf, etc.
			if len(parts) >= 2 {
				return strings.Join(parts[:2], "/")
			}
		}
	}

	// gopkg.in/yaml.v3, etc.
	if len(parts) >= 2 && parts[0] == "gopkg.in" {
		return strings.Join(parts[:2], "/")
	}

	// For other cases, return the entire path
	return modulePath
}
