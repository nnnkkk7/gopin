// Package resolver provides functionality to resolve Go module versions.
package resolver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

// Resolver is the interface for resolving the latest version of modules
type Resolver interface {
	// LatestVersion retrieves the latest version of a module
	LatestVersion(ctx context.Context, modulePath string) (string, error)
}

// VersionInfo is the response from the @latest endpoint of proxy.golang.org
type VersionInfo struct {
	Version string    `json:"Version"`
	Time    time.Time `json:"Time"`
}

// ProxyResolver resolves versions using proxy.golang.org
type ProxyResolver struct {
	HTTPClient *http.Client
	ProxyURL   string
}

// NewProxyResolver creates a new ProxyResolver
func NewProxyResolver(proxyURL string, timeout time.Duration) *ProxyResolver {
	if proxyURL == "" {
		proxyURL = "https://proxy.golang.org"
	}
	return &ProxyResolver{
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
		ProxyURL: proxyURL,
	}
}

// LatestVersion retrieves the latest version of a module
func (r *ProxyResolver) LatestVersion(ctx context.Context, modulePath string) (string, error) {
	url := fmt.Sprintf("%s/%s/@latest", r.ProxyURL, escapePath(modulePath))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	resp, err := r.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request to proxy: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("module not found: %s", modulePath)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("proxy returned status %d for %s", resp.StatusCode, modulePath)
	}

	var info VersionInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if info.Version == "" {
		return "", fmt.Errorf("empty version in response for %s", modulePath)
	}

	return info.Version, nil
}

// escapePath URL-escapes the module path
// Go module proxy specification: uppercase letters are converted to !lowercase
// Example: GitHub.com/Owner/Repo → github.com/!owner/!repo
func escapePath(path string) string {
	var result strings.Builder
	for _, r := range path {
		if 'A' <= r && r <= 'Z' {
			result.WriteByte('!')
			result.WriteRune(r + ('a' - 'A'))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// GoListResolver resolves versions using the go list command
// For environments without access to proxy.golang.org (such as private modules)
type GoListResolver struct {
	GoCommand string
}

// NewGoListResolver creates a new GoListResolver
func NewGoListResolver() *GoListResolver {
	return &GoListResolver{
		GoCommand: "go",
	}
}

// goListOutput is the output structure of go list -m -json
type goListOutput struct {
	Path    string `json:"Path"`
	Version string `json:"Version"`
	Error   *struct {
		Err string `json:"Err"`
	} `json:"Error"`
}

// LatestVersion retrieves the latest version of a module
func (r *GoListResolver) LatestVersion(ctx context.Context, modulePath string) (string, error) {
	cmd := exec.CommandContext(ctx, r.GoCommand, "list", "-m", "-json", modulePath+"@latest")

	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("go list failed: %s", string(exitErr.Stderr))
		}
		return "", fmt.Errorf("go list failed: %w", err)
	}

	var info goListOutput
	if err := json.Unmarshal(out, &info); err != nil {
		return "", fmt.Errorf("parse go list output: %w", err)
	}

	if info.Error != nil {
		return "", fmt.Errorf("go list error: %s", info.Error.Err)
	}

	if info.Version == "" {
		return "", fmt.Errorf("no version resolved for %s", modulePath)
	}

	return info.Version, nil
}

// FallbackResolver tries multiple resolvers
type FallbackResolver struct {
	Primary  Resolver
	Fallback Resolver
}

// NewFallbackResolver creates a new FallbackResolver
func NewFallbackResolver(primary, fallback Resolver) *FallbackResolver {
	return &FallbackResolver{
		Primary:  primary,
		Fallback: fallback,
	}
}

// LatestVersion retrieves the latest version of a module
// If Primary fails, try Fallback
func (r *FallbackResolver) LatestVersion(ctx context.Context, modulePath string) (string, error) {
	version, err := r.Primary.LatestVersion(ctx, modulePath)
	if err == nil {
		return version, nil
	}

	// Try fallback
	fallbackVersion, fallbackErr := r.Fallback.LatestVersion(ctx, modulePath)
	if fallbackErr == nil {
		return fallbackVersion, nil
	}

	// If both fail, return the Primary error
	return "", fmt.Errorf("primary resolver: %w; fallback resolver: %v", err, fallbackErr)
}

// CachedResolver caches version resolution results
type CachedResolver struct {
	Resolver Resolver
	cache    map[string]string
}

// NewCachedResolver creates a new CachedResolver
func NewCachedResolver(resolver Resolver) *CachedResolver {
	return &CachedResolver{
		Resolver: resolver,
		cache:    make(map[string]string),
	}
}

// LatestVersion retrieves the latest version of a module
// Requests for the same module path are returned from cache
func (r *CachedResolver) LatestVersion(ctx context.Context, modulePath string) (string, error) {
	if version, ok := r.cache[modulePath]; ok {
		return version, nil
	}

	version, err := r.Resolver.LatestVersion(ctx, modulePath)
	if err != nil {
		return "", err
	}

	r.cache[modulePath] = version
	return version, nil
}
