package cli

import (
	"context"
	"fmt"
)

// Version information injected at build time via ldflags
var (
	Version = "dev"     // Set via -ldflags "-X github.com/nnnkkk7/gopin/pkg/cli.Version=..."
	Commit  = "unknown" // Set via -ldflags "-X github.com/nnnkkk7/gopin/pkg/cli.Commit=..."
	Date    = "unknown" // Set via -ldflags "-X github.com/nnnkkk7/gopin/pkg/cli.Date=..."
)

// App represents the CLI application
type App struct {
	// TODO: Add actual application logic
}

// NewApp creates a new App instance
func NewApp() *App {
	return &App{}
}

// Run executes the CLI application
func Run(ctx context.Context, args []string) error {
	// TODO: Implement actual CLI logic
	// For now, just print version information as a placeholder
	fmt.Printf("gopin version %s (commit: %s, built: %s)\n", Version, Commit, Date)
	fmt.Println("This is a placeholder. Actual CLI implementation coming soon.")
	return nil
}

// Run is a method on App for compatibility with main.go
func (a *App) Run(ctx context.Context, args []string) error {
	return Run(ctx, args)
}
