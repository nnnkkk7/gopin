package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg.Version != 1 {
		t.Errorf("Default().Version = %d, want 1", cfg.Version)
	}

	// Expected file patterns (new default)
	expectedPatterns := []string{
		".github/**/*.yml",
		".github/**/*.yaml",
		"Makefile",
		"makefile",
		"GNUmakefile",
		"*.mk",
	}

	if len(cfg.Files) != len(expectedPatterns) {
		t.Errorf("Default().Files count = %d, want %d", len(cfg.Files), len(expectedPatterns))
	}

	// Verify each pattern
	for i, expected := range expectedPatterns {
		if i >= len(cfg.Files) {
			break
		}
		if cfg.Files[i].Pattern != expected {
			t.Errorf("Default().Files[%d].Pattern = %q, want %q", i, cfg.Files[i].Pattern, expected)
		}
	}

	// Ensure scripts patterns are not included
	for _, f := range cfg.Files {
		if f.Pattern == "scripts/*.sh" || f.Pattern == "scripts/*.bash" {
			t.Errorf("Default().Files should not include scripts pattern: %q", f.Pattern)
		}
	}

	if cfg.Resolver.Type != "proxy" {
		t.Errorf("Default().Resolver.Type = %q, want %q", cfg.Resolver.Type, "proxy")
	}

	if cfg.Resolver.ProxyURL != "https://proxy.golang.org" {
		t.Errorf("Default().Resolver.ProxyURL = %q, want %q", cfg.Resolver.ProxyURL, "https://proxy.golang.org")
	}

	if cfg.Resolver.Timeout != 30*time.Second {
		t.Errorf("Default().Resolver.Timeout = %v, want %v", cfg.Resolver.Timeout, 30*time.Second)
	}
}

func TestLoad(t *testing.T) {
	tests := []struct {
		want       *Config
		name       string
		configYAML string
		fileName   string
		wantErr    bool
	}{
		{
			name: "valid config",
			configYAML: `version: 1
files:
  - pattern: "Makefile"
  - pattern: ".github/workflows/*.yml"
ignore_modules:
  - name: "github.com/internal/.*"
    reason: "internal tools"
`,
			fileName: ".gopin.yaml",
			want: &Config{
				Version: 1,
				Files: []FilePattern{
					{Pattern: "Makefile"},
					{Pattern: ".github/workflows/*.yml"},
				},
				IgnoreModules: []IgnoreModule{
					{Name: "github.com/internal/.*", Reason: "internal tools"},
				},
				Resolver: ResolverConfig{
					Type:     "proxy",
					ProxyURL: "https://proxy.golang.org",
					Timeout:  30 * time.Second,
					Fallback: true,
				},
				ExcludeModules: []string{"github.com/internal/.*"},
			},
			wantErr: false,
		},
		{
			name: "minimal config",
			configYAML: `version: 1
files:
  - pattern: "Makefile"
`,
			fileName: ".gopin.yaml",
			want: &Config{
				Version: 1,
				Files: []FilePattern{
					{Pattern: "Makefile"},
				},
				Resolver: ResolverConfig{
					Type:     "proxy",
					ProxyURL: "https://proxy.golang.org",
					Timeout:  30 * time.Second,
					Fallback: true,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, tt.fileName)

			// Write config file
			// #nosec G306 - test file
			if err := os.WriteFile(configPath, []byte(tt.configYAML), 0o644); err != nil {
				t.Fatalf("Failed to write config file: %v", err)
			}

			// Load config
			got, err := Load(configPath)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Load() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			// Compare basic fields
			if got.Version != tt.want.Version {
				t.Errorf("Version = %d, want %d", got.Version, tt.want.Version)
			}

			// Compare file patterns (only check count and basenames since paths are absolute)
			if len(got.Files) != len(tt.want.Files) {
				t.Errorf("Files count = %d, want %d", len(got.Files), len(tt.want.Files))
			}

			for i := range got.Files {
				gotBase := filepath.Base(got.Files[i].Pattern)
				wantBase := filepath.Base(tt.want.Files[i].Pattern)
				if gotBase != wantBase {
					t.Errorf("Files[%d] basename = %s, want %s", i, gotBase, wantBase)
				}
			}

			// Compare other fields using go-cmp
			if diff := cmp.Diff(tt.want.Resolver, got.Resolver); diff != "" {
				t.Errorf("Resolver mismatch (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(tt.want.IgnoreModules, got.IgnoreModules); diff != "" {
				t.Errorf("IgnoreModules mismatch (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(tt.want.ExcludeModules, got.ExcludeModules); diff != "" {
				t.Errorf("ExcludeModules mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestLoad_DefaultWhenNotFound(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Change to temp directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldWd) }() // Explicitly ignore error in cleanup

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Load config (should return default)
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Should return default config
	defaultCfg := Default()
	if diff := cmp.Diff(defaultCfg, cfg); diff != "" {
		t.Errorf("Load() should return default config when file not found, mismatch (-want +got):\n%s", diff)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		config  *Config
		name    string
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  Default(),
			wantErr: false,
		},
		{
			name: "invalid version",
			config: &Config{
				Version: 2,
				Files: []FilePattern{
					{Pattern: "Makefile"},
				},
			},
			wantErr: true,
		},
		{
			name: "no file patterns",
			config: &Config{
				Version: 1,
				Files:   []FilePattern{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResolverConfig_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		want    ResolverConfig
		wantErr bool
	}{
		{
			name: "full config",
			yaml: `  type: golist
  proxy_url: https://custom.proxy
  timeout: 60s
  fallback: false
`,
			want: ResolverConfig{
				Type:     "golist",
				ProxyURL: "https://custom.proxy",
				Timeout:  60 * time.Second,
				Fallback: false,
			},
			wantErr: false,
		},
		{
			name: "default values",
			yaml: ``,
			want: ResolverConfig{
				Type:     "proxy",
				ProxyURL: "https://proxy.golang.org",
				Timeout:  30 * time.Second,
				Fallback: true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a full config with resolver
			fullYAML := `version: 1
files:
  - pattern: "Makefile"
resolver:
` + tt.yaml

			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, ".gopin.yaml")

			// #nosec G306 - test file
			if err := os.WriteFile(configPath, []byte(fullYAML), 0o644); err != nil {
				t.Fatalf("Failed to write config file: %v", err)
			}

			cfg, err := Load(configPath)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Load() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			if diff := cmp.Diff(tt.want, cfg.Resolver); diff != "" {
				t.Errorf("ResolverConfig mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
