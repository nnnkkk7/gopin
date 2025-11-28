// Package config provides configuration loading for gopin.
package config

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/goccy/go-yaml"
)

// Config represents gopin configuration
type Config struct {
	Version        int            `yaml:"version"`
	Files          []FilePattern  `yaml:"files"`
	IgnoreModules  []IgnoreModule `yaml:"ignore_modules"`
	Resolver       ResolverConfig `yaml:"resolver"`
	IncludeModules []string       `yaml:"-"` // Set only from CLI
	ExcludeModules []string       `yaml:"-"` // Set only from CLI
}

// FilePattern represents a file pattern configuration
type FilePattern struct {
	Pattern string `yaml:"pattern"`
}

// IgnoreModule represents a module to ignore
type IgnoreModule struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
	Reason  string `yaml:"reason"`
}

// ResolverConfig represents resolver configuration
type ResolverConfig struct {
	Type     string        `yaml:"type"`
	ProxyURL string        `yaml:"proxy_url"`
	Timeout  time.Duration `yaml:"timeout"`
	Fallback bool          `yaml:"fallback"`
}

// Default returns default configuration
func Default() *Config {
	return &Config{
		Version: 1,
		Files: []FilePattern{
			{Pattern: ".github/**/*.yml"},
			{Pattern: ".github/**/*.yaml"},
			{Pattern: "Makefile"},
			{Pattern: "makefile"},
			{Pattern: "GNUmakefile"},
			{Pattern: "*.mk"},
		},
		Resolver: ResolverConfig{
			Type:     "proxy",
			ProxyURL: "https://proxy.golang.org",
			Timeout:  30 * time.Second,
			Fallback: true,
		},
	}
}

// configPaths is the list of config file search paths
var configPaths = []string{
	".gopin.yaml",
	".gopin.yml",
	".github/gopin.yaml",
	".github/gopin.yml",
}

// Load loads configuration from file
func Load(path string) (*Config, error) {
	if path != "" {
		return loadFile(path)
	}

	// Search for config file
	for _, p := range configPaths {
		if _, err := os.Stat(p); err == nil {
			return loadFile(p)
		}
	}

	// Return default if config file not found
	return Default(), nil
}

// loadFile loads configuration from specified file
func loadFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := Default() // Start with default values
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	// Validate version
	if cfg.Version != 1 {
		return nil, errors.New("unsupported config version")
	}

	// Convert patterns to absolute paths
	dir := filepath.Dir(path)
	for i, f := range cfg.Files {
		if !filepath.IsAbs(f.Pattern) {
			cfg.Files[i].Pattern = filepath.Join(dir, f.Pattern)
		}
	}

	// Convert ignore_modules to ExcludeModules
	for _, m := range cfg.IgnoreModules {
		cfg.ExcludeModules = append(cfg.ExcludeModules, m.Name)
	}

	return cfg, nil
}

// UnmarshalYAML implements custom unmarshaling for ResolverConfig
func (r *ResolverConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Set defaults first
	*r = ResolverConfig{
		Type:     "proxy",
		ProxyURL: "https://proxy.golang.org",
		Timeout:  30 * time.Second,
		Fallback: true,
	}

	// Create a temporary struct to avoid infinite recursion
	type resolverConfig ResolverConfig
	temp := (*resolverConfig)(r)

	if err := unmarshal(temp); err != nil {
		return err
	}

	return nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Version != 1 {
		return errors.New("unsupported config version")
	}

	if len(c.Files) == 0 {
		return errors.New("no file patterns specified")
	}

	return nil
}
