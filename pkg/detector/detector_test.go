package detector

import (
	"testing"
)

func TestDetector_Detect(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []Match
	}{
		{
			name:    "simple @latest",
			content: `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`,
			want: []Match{
				{
					FullMatch:   "go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest",
					ModulePath:  "github.com/golangci/golangci-lint/cmd/golangci-lint",
					Version:     "latest",
					Line:        1,
					StartColumn: 0,
				},
			},
		},
		{
			name:    "with version",
			content: `go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.61.0`,
			want: []Match{
				{
					FullMatch:   "go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.61.0",
					ModulePath:  "github.com/golangci/golangci-lint/cmd/golangci-lint",
					Version:     "v1.61.0",
					Line:        1,
					StartColumn: 0,
				},
			},
		},
		{
			name:    "no version",
			content: `go install github.com/example/tool`,
			want: []Match{
				{
					FullMatch:   "go install github.com/example/tool",
					ModulePath:  "github.com/example/tool",
					Version:     "",
					Line:        1,
					StartColumn: 0,
				},
			},
		},
		{
			name:    "with flags",
			content: `go install -v github.com/example/tool@latest`,
			want: []Match{
				{
					FullMatch:   "go install -v github.com/example/tool@latest",
					ModulePath:  "github.com/example/tool",
					Version:     "latest",
					Line:        1,
					StartColumn: 0,
				},
			},
		},
		{
			name:    "with tags flag",
			content: `go install -tags=netgo github.com/example/tool@latest`,
			want: []Match{
				{
					FullMatch:   "go install -tags=netgo github.com/example/tool@latest",
					ModulePath:  "github.com/example/tool",
					Version:     "latest",
					Line:        1,
					StartColumn: 0,
				},
			},
		},
		{
			name: "in yaml",
			content: `jobs:
  lint:
    steps:
      - run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`,
			want: []Match{
				{
					FullMatch:   "go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest",
					ModulePath:  "github.com/golangci/golangci-lint/cmd/golangci-lint",
					Version:     "latest",
					Line:        4,
					StartColumn: 14,
				},
			},
		},
		{
			name: "multiple matches",
			content: `go install github.com/tool1@latest
go install github.com/tool2@v1.0.0`,
			want: []Match{
				{
					FullMatch:   "go install github.com/tool1@latest",
					ModulePath:  "github.com/tool1",
					Version:     "latest",
					Line:        1,
					StartColumn: 0,
				},
				{
					FullMatch:   "go install github.com/tool2@v1.0.0",
					ModulePath:  "github.com/tool2",
					Version:     "v1.0.0",
					Line:        2,
					StartColumn: 0,
				},
			},
		},
		{
			name: "in makefile",
			content: `.PHONY: lint
lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run`,
			want: []Match{
				{
					FullMatch:   "go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest",
					ModulePath:  "github.com/golangci/golangci-lint/cmd/golangci-lint",
					Version:     "latest",
					Line:        3,
					StartColumn: 1,
				},
			},
		},
		{
			name:    "no match",
			content: `go build ./...`,
			want:    nil,
		},
		{
			name:    "pseudo version",
			content: `go install github.com/example/tool@v0.0.0-20240101120000-abcdef123456`,
			want: []Match{
				{
					FullMatch:   "go install github.com/example/tool@v0.0.0-20240101120000-abcdef123456",
					ModulePath:  "github.com/example/tool",
					Version:     "v0.0.0-20240101120000-abcdef123456",
					Line:        1,
					StartColumn: 0,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := NewDetector(nil, nil)
			if err != nil {
				t.Fatalf("NewDetector() error = %v", err)
			}

			got, err := d.Detect(tt.content)
			if err != nil {
				t.Fatalf("Detect() error = %v", err)
			}

			if len(got) != len(tt.want) {
				t.Fatalf("Detect() got %d matches, want %d", len(got), len(tt.want))
			}

			for i, g := range got {
				w := tt.want[i]
				if g.FullMatch != w.FullMatch {
					t.Errorf("Match[%d].FullMatch = %q, want %q", i, g.FullMatch, w.FullMatch)
				}
				if g.ModulePath != w.ModulePath {
					t.Errorf("Match[%d].ModulePath = %q, want %q", i, g.ModulePath, w.ModulePath)
				}
				if g.Version != w.Version {
					t.Errorf("Match[%d].Version = %q, want %q", i, g.Version, w.Version)
				}
				if g.Line != w.Line {
					t.Errorf("Match[%d].Line = %d, want %d", i, g.Line, w.Line)
				}
			}
		})
	}
}

func TestNeedsPin(t *testing.T) {
	tests := []struct {
		version string
		want    bool
	}{
		{"", true},
		{"latest", true},
		{"v1.0.0", true},  // Changed: now all versions should be updated
		{"v1.61.0", true}, // Changed: now all versions should be updated
		{"v0.0.0-20240101120000-abcdef123456", true}, // Changed: now all versions should be updated
		{"master", true},  // Added: branch references should also be updated
		{"main", true},    // Added: branch references should also be updated
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			got := NeedsPin(tt.version)
			if got != tt.want {
				t.Errorf("NeedsPin(%q) = %v, want %v", tt.version, got, tt.want)
			}
		})
	}
}

func TestExtractRootModule(t *testing.T) {
	tests := []struct {
		modulePath string
		want       string
	}{
		{
			"github.com/golangci/golangci-lint/cmd/golangci-lint",
			"github.com/golangci/golangci-lint",
		},
		{
			"github.com/owner/repo",
			"github.com/owner/repo",
		},
		{
			"golang.org/x/tools/cmd/stringer",
			"golang.org/x/tools",
		},
		{
			"google.golang.org/protobuf/cmd/protoc-gen-go",
			"google.golang.org/protobuf",
		},
		{
			"gopkg.in/yaml.v3",
			"gopkg.in/yaml.v3",
		},
		{
			"gitlab.com/owner/repo/cmd/tool",
			"gitlab.com/owner/repo",
		},
		{
			"example.com/custom/module",
			"example.com/custom/module",
		},
	}

	for _, tt := range tests {
		t.Run(tt.modulePath, func(t *testing.T) {
			got := ExtractRootModule(tt.modulePath)
			if got != tt.want {
				t.Errorf("ExtractRootModule(%q) = %q, want %q", tt.modulePath, got, tt.want)
			}
		})
	}
}

func TestDetector_ShouldProcess(t *testing.T) {
	tests := []struct {
		name       string
		includes   []string
		excludes   []string
		modulePath string
		want       bool
	}{
		{
			name:       "no filters",
			includes:   nil,
			excludes:   nil,
			modulePath: "github.com/example/tool",
			want:       true,
		},
		{
			name:       "include match",
			includes:   []string{"github.com/example/.*"},
			excludes:   nil,
			modulePath: "github.com/example/tool",
			want:       true,
		},
		{
			name:       "include no match",
			includes:   []string{"github.com/other/.*"},
			excludes:   nil,
			modulePath: "github.com/example/tool",
			want:       false,
		},
		{
			name:       "exclude match",
			includes:   nil,
			excludes:   []string{"github.com/example/.*"},
			modulePath: "github.com/example/tool",
			want:       false,
		},
		{
			name:       "exclude no match",
			includes:   nil,
			excludes:   []string{"github.com/other/.*"},
			modulePath: "github.com/example/tool",
			want:       true,
		},
		{
			name:       "include and exclude",
			includes:   []string{"github.com/.*"},
			excludes:   []string{"github.com/internal/.*"},
			modulePath: "github.com/example/tool",
			want:       true,
		},
		{
			name:       "include and exclude (excluded)",
			includes:   []string{"github.com/.*"},
			excludes:   []string{"github.com/internal/.*"},
			modulePath: "github.com/internal/tool",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := NewDetector(tt.includes, tt.excludes)
			if err != nil {
				t.Fatalf("NewDetector() error = %v", err)
			}

			got := d.ShouldProcess(tt.modulePath)
			if got != tt.want {
				t.Errorf("ShouldProcess(%q) = %v, want %v", tt.modulePath, got, tt.want)
			}
		})
	}
}
