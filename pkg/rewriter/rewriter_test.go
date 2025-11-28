package rewriter

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/nnnkkk7/gopin/pkg/detector"
)

func TestRewriter_Rewrite(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		versions map[string]string
		want     string
		wantLen  int // expected number of changes
	}{
		{
			name:    "single replacement with @latest",
			content: `go install github.com/example/tool@latest`,
			versions: map[string]string{
				"github.com/example/tool": "v1.0.0",
			},
			want:    `go install github.com/example/tool@v1.0.0`,
			wantLen: 1,
		},
		{
			name:    "no version specified",
			content: `go install github.com/example/tool`,
			versions: map[string]string{
				"github.com/example/tool": "v1.0.0",
			},
			want:    `go install github.com/example/tool@v1.0.0`,
			wantLen: 1,
		},
		{
			name: "multiple replacements",
			content: `go install github.com/tool1@latest
go install github.com/tool2@latest`,
			versions: map[string]string{
				"github.com/tool1": "v1.0.0",
				"github.com/tool2": "v2.0.0",
			},
			want: `go install github.com/tool1@v1.0.0
go install github.com/tool2@v2.0.0`,
			wantLen: 2,
		},
		{
			name: "in makefile with tabs",
			content: `.PHONY: lint
lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run`,
			versions: map[string]string{
				"github.com/golangci/golangci-lint/cmd/golangci-lint": "v1.61.0",
			},
			want: `.PHONY: lint
lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.61.0
	golangci-lint run`,
			wantLen: 1,
		},
		{
			name:     "no changes needed",
			content:  `go build ./...`,
			versions: map[string]string{},
			want:    `go build ./...`,
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use detector to get matches
			det, err := detector.NewDetector(nil, nil)
			if err != nil {
				t.Fatalf("NewDetector() error = %v", err)
			}

			matches, err := det.Detect(tt.content)
			if err != nil {
				t.Fatalf("Detect() error = %v", err)
			}

			rew := NewRewriter(Options{})
			result, err := rew.Rewrite(tt.content, matches, tt.versions)
			if err != nil {
				t.Fatalf("Rewrite() error = %v", err)
			}

			if diff := cmp.Diff(tt.wantLen, len(result.Changes)); diff != "" {
				t.Errorf("Rewrite() changes count mismatch (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(tt.want, result.Content); diff != "" {
				t.Errorf("Rewrite() content mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestFormatDiff(t *testing.T) {
	tests := []struct {
		name         string
		filePath     string
		changes      []Change
		wantContains []string
	}{
		{
			name:     "single change",
			filePath: "Makefile",
			changes: []Change{
				{
					Line:       3,
					OldText:    "go install github.com/example/tool@latest",
					NewText:    "go install github.com/example/tool@v1.0.0",
					ModulePath: "github.com/example/tool",
					OldVersion: "latest",
					NewVersion: "v1.0.0",
				},
			},
			wantContains: []string{
				"--- Makefile",
				"+++ Makefile",
				"- go install github.com/example/tool@latest",
				"+ go install github.com/example/tool@v1.0.0",
			},
		},
		{
			name:         "no changes",
			filePath:     "Makefile",
			changes:      []Change{},
			wantContains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff := FormatDiff(tt.filePath, tt.changes)

			if len(tt.wantContains) == 0 {
				if diff != "" {
					t.Errorf("FormatDiff() expected empty, got %q", diff)
				}
				return
			}

			for _, want := range tt.wantContains {
				if !cmp.Equal(true, containsString(diff, want)) {
					t.Errorf("FormatDiff() missing expected string %q", want)
				}
			}
		})
	}
}

func TestFormatChangeSummary(t *testing.T) {
	tests := []struct {
		name         string
		changes      []Change
		wantContains []string
	}{
		{
			name: "multiple changes",
			changes: []Change{
				{
					Line:       1,
					ModulePath: "github.com/example/tool",
					OldVersion: "latest",
					NewVersion: "v1.0.0",
				},
				{
					Line:       2,
					ModulePath: "github.com/another/tool",
					OldVersion: "",
					NewVersion: "v2.0.0",
				},
			},
			wantContains: []string{
				"github.com/example/tool",
				"latest -> v1.0.0",
				"(none) -> v2.0.0",
			},
		},
		{
			name:         "no changes",
			changes:      []Change{},
			wantContains: []string{"No changes"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := FormatChangeSummary(tt.changes)

			for _, want := range tt.wantContains {
				if !containsString(summary, want) {
					t.Errorf("FormatChangeSummary() missing expected string %q", want)
				}
			}
		})
	}
}

func TestHasChanges(t *testing.T) {
	tests := []struct {
		name    string
		results []*Result
		want    bool
	}{
		{
			name: "has changes",
			results: []*Result{
				{
					Changes: []Change{{Line: 1}},
				},
			},
			want: true,
		},
		{
			name: "no changes",
			results: []*Result{
				{
					Changes: []Change{},
				},
			},
			want: false,
		},
		{
			name:    "empty results",
			results: []*Result{},
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasChanges(tt.results)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("HasChanges() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// containsString checks if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
