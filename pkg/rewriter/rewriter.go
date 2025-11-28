// Package rewriter provides functionality to rewrite go install commands with pinned versions.
package rewriter

import (
	"fmt"
	"sort"
	"strings"

	"github.com/nnnkkk7/gopin/pkg/detector"
)

// Change represents a single rewrite change
type Change struct {
	Line       int    // Line number (1-indexed)
	OldText    string // Original text
	NewText    string // New text
	ModulePath string // Module path
	OldVersion string // Original version
	NewVersion string // New version
}

// Result represents the rewriting result
type Result struct {
	FilePath string   // File path
	Changes  []Change // List of changes
	Content  string   // New file content
}

// Options represents rewriter options
type Options struct {
	// Reserved for future options
}

// DefaultOptions returns default options
func DefaultOptions() Options {
	return Options{}
}

// Rewriter rewrites file content
type Rewriter struct {
	Options Options
}

// NewRewriter creates a new Rewriter
func NewRewriter(opts Options) *Rewriter {
	return &Rewriter{Options: opts}
}

// Rewrite rewrites matched occurrences to new versions
// versions is a map from modulePath to newVersion
func (w *Rewriter) Rewrite(content string, matches []detector.Match, versions map[string]string) (*Result, error) {
	if len(matches) == 0 {
		return &Result{Content: content}, nil
	}

	// Sort matches by line number and column (descending order to process from the end)
	sortedMatches := make([]detector.Match, len(matches))
	copy(sortedMatches, matches)
	sort.Slice(sortedMatches, func(i, j int) bool {
		if sortedMatches[i].Line != sortedMatches[j].Line {
			return sortedMatches[i].Line > sortedMatches[j].Line
		}
		return sortedMatches[i].StartColumn > sortedMatches[j].StartColumn
	})

	lines := strings.Split(content, "\n")
	var changes []Change

	for _, m := range sortedMatches {
		newVersion, ok := versions[m.ModulePath]
		if !ok {
			// Skip if version was not resolved
			continue
		}

		// Line number is 1-indexed
		lineIdx := m.Line - 1
		if lineIdx < 0 || lineIdx >= len(lines) {
			continue
		}

		line := lines[lineIdx]
		oldText := m.FullMatch

		// Generate new text
		var newText string
		if m.Version == "" {
			// No version specified -> add @vX.Y.Z to the end
			newText = oldText + "@" + newVersion
		} else {
			// Replace existing version
			newText = strings.Replace(oldText, "@"+m.Version, "@"+newVersion, 1)
		}

		// Update line
		newLine := line[:m.StartColumn] + newText + line[m.EndColumn:]
		lines[lineIdx] = newLine

		changes = append(changes, Change{
			Line:       m.Line,
			OldText:    oldText,
			NewText:    newText,
			ModulePath: m.ModulePath,
			OldVersion: m.Version,
			NewVersion: newVersion,
		})
	}

	// Sort changes by line number (ascending order)
	sort.Slice(changes, func(i, j int) bool {
		return changes[i].Line < changes[j].Line
	})

	return &Result{
		Changes: changes,
		Content: strings.Join(lines, "\n"),
	}, nil
}

// FormatDiff generates a diff format string
func FormatDiff(filePath string, changes []Change) string {
	if len(changes) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("--- %s\n", filePath))
	sb.WriteString(fmt.Sprintf("+++ %s\n", filePath))

	for _, c := range changes {
		sb.WriteString(fmt.Sprintf("@@ -%d +%d @@\n", c.Line, c.Line))
		sb.WriteString(fmt.Sprintf("- %s\n", c.OldText))
		sb.WriteString(fmt.Sprintf("+ %s\n", c.NewText))
	}

	return sb.String()
}

// FormatChangeSummary generates a change summary
func FormatChangeSummary(changes []Change) string {
	if len(changes) == 0 {
		return "No changes"
	}

	var sb strings.Builder
	for _, c := range changes {
		oldVer := c.OldVersion
		if oldVer == "" {
			oldVer = "(none)"
		}
		sb.WriteString(fmt.Sprintf("  %s: %s -> %s\n", c.ModulePath, oldVer, c.NewVersion))
	}
	return sb.String()
}

// HasChanges returns whether there are any changes
func HasChanges(results []*Result) bool {
	for _, r := range results {
		if len(r.Changes) > 0 {
			return true
		}
	}
	return false
}
