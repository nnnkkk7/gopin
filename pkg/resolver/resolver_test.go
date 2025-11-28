package resolver

import (
	"context"
	"testing"
	"time"
)

func TestProxyResolver_LatestVersion(t *testing.T) {
	resolver := NewProxyResolver("", 30*time.Second)
	ctx := context.Background()

	tests := []struct {
		name       string
		modulePath string
		wantErr    bool
	}{
		{
			name:       "valid module - golang.org/x/tools",
			modulePath: "golang.org/x/tools",
			wantErr:    false,
		},
		{
			name:       "nonexistent module",
			modulePath: "github.com/nonexistent/module/that/does/not/exist",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := resolver.LatestVersion(ctx, tt.modulePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("LatestVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if version == "" {
					t.Errorf("LatestVersion() returned empty version for valid module")
				}
				if version[0] != 'v' {
					t.Errorf("LatestVersion() = %q, expected version to start with 'v'", version)
				}
				t.Logf("Resolved %s to %s", tt.modulePath, version)
			}
		})
	}
}

func TestEscapePath(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{
			input: "github.com/example/tool",
			want:  "github.com/example/tool",
		},
		{
			input: "GitHub.com/Owner/Repo",
			want:  "!git!hub.com/!owner/!repo",
		},
		{
			input: "Example.COM/Package",
			want:  "!example.!c!o!m/!package",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := escapePath(tt.input)
			if got != tt.want {
				t.Errorf("escapePath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestGoListResolver_LatestVersion(t *testing.T) {
	// This test requires 'go' command to be available
	resolver := NewGoListResolver()
	ctx := context.Background()

	tests := []struct {
		name       string
		modulePath string
		wantErr    bool
	}{
		{
			name:       "valid module - golang.org/x/tools",
			modulePath: "golang.org/x/tools",
			wantErr:    false,
		},
		{
			name:       "nonexistent module",
			modulePath: "github.com/nonexistent/module/that/does/not/exist",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := resolver.LatestVersion(ctx, tt.modulePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("LatestVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if version == "" {
					t.Errorf("LatestVersion() returned empty version for valid module")
				}
				t.Logf("Resolved %s to %s", tt.modulePath, version)
			}
		})
	}
}

func TestCachedResolver(t *testing.T) {
	proxy := NewProxyResolver("", 30*time.Second)
	cached := NewCachedResolver(proxy)
	ctx := context.Background()

	modulePath := "golang.org/x/tools"

	// First call - should hit the proxy
	version1, err := cached.LatestVersion(ctx, modulePath)
	if err != nil {
		t.Fatalf("First LatestVersion() error = %v", err)
	}

	// Second call - should return from cache
	version2, err := cached.LatestVersion(ctx, modulePath)
	if err != nil {
		t.Fatalf("Second LatestVersion() error = %v", err)
	}

	if version1 != version2 {
		t.Errorf("Cached version mismatch: %q != %q", version1, version2)
	}

	// Verify it's actually cached (should be instant)
	start := time.Now()
	_, err = cached.LatestVersion(ctx, modulePath)
	duration := time.Since(start)
	if err != nil {
		t.Fatalf("Cached LatestVersion() error = %v", err)
	}
	if duration > 10*time.Millisecond {
		t.Errorf("Cached lookup took %v, expected < 10ms", duration)
	}
}

func TestFallbackResolver(t *testing.T) {
	proxy := NewProxyResolver("", 30*time.Second)
	golist := NewGoListResolver()
	fallback := NewFallbackResolver(proxy, golist)
	ctx := context.Background()

	modulePath := "golang.org/x/tools"

	version, err := fallback.LatestVersion(ctx, modulePath)
	if err != nil {
		t.Fatalf("FallbackResolver LatestVersion() error = %v", err)
	}

	if version == "" {
		t.Errorf("FallbackResolver returned empty version")
	}

	t.Logf("Resolved %s to %s via fallback", modulePath, version)
}
