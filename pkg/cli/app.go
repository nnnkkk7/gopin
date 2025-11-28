// Package cli provides the command-line interface for gopin.
package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v3"

	"github.com/nnnkkk7/gopin/pkg/config"
	"github.com/nnnkkk7/gopin/pkg/detector"
	"github.com/nnnkkk7/gopin/pkg/resolver"
	"github.com/nnnkkk7/gopin/pkg/rewriter"
)

// Version is the version string embedded at build time
var Version = "dev"

// NewApp creates a new CLI application
func NewApp() *cli.Command {
	return &cli.Command{
		Name:    "gopin",
		Usage:   "Pin versions of go install commands",
		Version: Version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "config file path",
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "verbose output",
			},
		},
		Commands: []*cli.Command{
			runCommand(),
			checkCommand(),
			listCommand(),
			initCommand(),
		},
	}
}

func runCommand() *cli.Command {
	return &cli.Command{
		Name:  "run",
		Usage: "Pin go install versions in files",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "dry-run",
				Aliases: []string{"n"},
				Usage:   "show changes without applying",
			},
			&cli.BoolFlag{
				Name:    "diff",
				Aliases: []string{"d"},
				Usage:   "show diff output",
			},
			&cli.StringSliceFlag{
				Name:    "include",
				Aliases: []string{"i"},
				Usage:   "include modules matching regex",
			},
			&cli.StringSliceFlag{
				Name:    "exclude",
				Aliases: []string{"e"},
				Usage:   "exclude modules matching regex",
			},
		},
		Action: runAction,
	}
}

func runAction(ctx context.Context, c *cli.Command) error {
	logger := setupLogger(c.Bool("verbose"))

	// Load configuration
	cfg, err := config.Load(c.String("config"))
	if err != nil {
		logger.Warn("failed to load config, using defaults", "error", err)
		cfg = config.Default()
	}

	// Merge CLI options into configuration
	if includes := c.StringSlice("include"); len(includes) > 0 {
		cfg.IncludeModules = includes
	}
	if excludes := c.StringSlice("exclude"); len(excludes) > 0 {
		cfg.ExcludeModules = excludes
	}

	// Get target files
	files, err := getTargetFiles(c.Args().Slice(), cfg)
	if err != nil {
		return fmt.Errorf("get target files: %w", err)
	}

	if len(files) == 0 {
		logger.Info("no target files found")
		return nil
	}

	// Create detector
	det, err := detector.NewDetector(cfg.IncludeModules, cfg.ExcludeModules)
	if err != nil {
		return fmt.Errorf("create detector: %w", err)
	}

	// Create resolver
	res := createResolver(cfg)

	// Create rewriter
	rew := rewriter.NewRewriter(rewriter.DefaultOptions())

	// Process files
	var results []*rewriter.Result

	for _, file := range files {
		result, err := processFile(ctx, logger, file, det, res, rew)
		if err != nil {
			logger.Error("failed to process file", "file", file, "error", err)
			continue
		}
		if result != nil {
			result.FilePath = file
			results = append(results, result)
		}
	}

	// Display/apply results
	dryRun := c.Bool("dry-run")
	showDiff := c.Bool("diff")

	for _, result := range results {
		if len(result.Changes) == 0 {
			continue
		}

		if showDiff {
			fmt.Println(rewriter.FormatDiff(result.FilePath, result.Changes))
		}

		if !dryRun {
			if err := os.WriteFile(result.FilePath, []byte(result.Content), 0644); err != nil {
				logger.Error("failed to write file", "file", result.FilePath, "error", err)
				continue
			}
			logger.Info("updated", "file", result.FilePath)
		} else {
			logger.Info("would update (dry-run)", "file", result.FilePath)
		}

		fmt.Printf("%s:\n%s", result.FilePath, rewriter.FormatChangeSummary(result.Changes))
	}

	if !rewriter.HasChanges(results) {
		logger.Info("no changes needed")
	}

	return nil
}

func checkCommand() *cli.Command {
	return &cli.Command{
		Name:  "check",
		Usage: "Check for unpinned go install commands",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "fix",
				Usage: "fix unpinned versions",
			},
			&cli.BoolFlag{
				Name:    "diff",
				Aliases: []string{"d"},
				Usage:   "show diff output",
			},
		},
		Action: checkAction,
	}
}

func checkAction(ctx context.Context, c *cli.Command) error {
	logger := setupLogger(c.Bool("verbose"))

	cfg, err := config.Load(c.String("config"))
	if err != nil {
		cfg = config.Default()
	}

	files, err := getTargetFiles(c.Args().Slice(), cfg)
	if err != nil {
		return fmt.Errorf("get target files: %w", err)
	}

	det, err := detector.NewDetector(cfg.IncludeModules, cfg.ExcludeModules)
	if err != nil {
		return fmt.Errorf("create detector: %w", err)
	}

	var unpinnedCount int
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			logger.Error("failed to read file", "file", file, "error", err)
			continue
		}

		matches, err := det.Detect(string(content))
		if err != nil {
			logger.Error("failed to detect", "file", file, "error", err)
			continue
		}

		for _, m := range matches {
			if detector.NeedsPin(m.Version) {
				unpinnedCount++
				fmt.Printf("%s:%d: %s@%s is not pinned\n", file, m.Line, m.ModulePath, m.Version)
			}
		}
	}

	if unpinnedCount > 0 {
		return fmt.Errorf("found %d unpinned go install commands", unpinnedCount)
	}

	logger.Info("all go install commands are pinned")
	return nil
}

func listCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List go install commands in files",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "unpinned",
				Usage: "show only unpinned versions",
			},
		},
		Action: listAction,
	}
}

func listAction(ctx context.Context, c *cli.Command) error {
	logger := setupLogger(c.Bool("verbose"))

	cfg, err := config.Load(c.String("config"))
	if err != nil {
		cfg = config.Default()
	}

	files, err := getTargetFiles(c.Args().Slice(), cfg)
	if err != nil {
		return fmt.Errorf("get target files: %w", err)
	}

	det, err := detector.NewDetector(cfg.IncludeModules, cfg.ExcludeModules)
	if err != nil {
		return fmt.Errorf("create detector: %w", err)
	}

	unpinnedOnly := c.Bool("unpinned")

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			logger.Error("failed to read file", "file", file, "error", err)
			continue
		}

		matches, err := det.Detect(string(content))
		if err != nil {
			logger.Error("failed to detect", "file", file, "error", err)
			continue
		}

		for _, m := range matches {
			if unpinnedOnly && !detector.NeedsPin(m.Version) {
				continue
			}

			version := m.Version
			if version == "" {
				version = "(none)"
			}
			status := "pinned"
			if detector.NeedsPin(m.Version) {
				status = "unpinned"
			}
			fmt.Printf("%s:%d: %s@%s [%s]\n", file, m.Line, m.ModulePath, version, status)
		}
	}

	return nil
}

func initCommand() *cli.Command {
	return &cli.Command{
		Name:  "init",
		Usage: "Create a config file",
		Action: func(ctx context.Context, c *cli.Command) error {
			path := c.Args().First()
			if path == "" {
				path = ".gopin.yaml"
			}

			if _, err := os.Stat(path); err == nil {
				return fmt.Errorf("config file already exists: %s", path)
			}

			content := `# gopin configuration
version: 1

# Target file patterns
files:
  - pattern: ".github/**/*.yml"
  - pattern: ".github/**/*.yaml"
  - pattern: "Makefile"
  - pattern: "makefile"
  - pattern: "GNUmakefile"
  - pattern: "*.mk"

# Modules to ignore (optional)
# ignore_modules:
#   - name: "github.com/internal/.*"
#     reason: "internal tools"
`
			if err := os.WriteFile(path, []byte(content), 0644); err != nil {
				return fmt.Errorf("write config file: %w", err)
			}

			fmt.Printf("Created %s\n", path)
			return nil
		},
	}
}

// getTargetFiles retrieves the list of target file paths
func getTargetFiles(args []string, cfg *config.Config) ([]string, error) {
	if len(args) > 0 {
		// Use files specified in command-line arguments
		var files []string
		for _, pattern := range args {
			matches, err := filepath.Glob(pattern)
			if err != nil {
				return nil, fmt.Errorf("glob %s: %w", pattern, err)
			}
			files = append(files, matches...)
		}
		return files, nil
	}

	// Use patterns from configuration
	var files []string
	for _, p := range cfg.Files {
		matches, err := filepath.Glob(p.Pattern)
		if err != nil {
			return nil, fmt.Errorf("glob %s: %w", p.Pattern, err)
		}
		files = append(files, matches...)
	}
	return files, nil
}

// createResolver creates a resolver based on configuration
func createResolver(cfg *config.Config) resolver.Resolver {
	var res resolver.Resolver

	switch cfg.Resolver.Type {
	case "golist":
		res = resolver.NewGoListResolver()
	default:
		res = resolver.NewProxyResolver(cfg.Resolver.ProxyURL, cfg.Resolver.Timeout)
	}

	if cfg.Resolver.Fallback {
		res = resolver.NewFallbackResolver(res, resolver.NewGoListResolver())
	}

	return resolver.NewCachedResolver(res)
}

// processFile processes a single file
func processFile(
	ctx context.Context,
	logger *slog.Logger,
	file string,
	det *detector.Detector,
	res resolver.Resolver,
	rew *rewriter.Rewriter,
) (*rewriter.Result, error) {
	content, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	matches, err := det.Detect(string(content))
	if err != nil {
		return nil, fmt.Errorf("detect: %w", err)
	}

	if len(matches) == 0 {
		return nil, nil
	}

	// Resolve versions
	versions := make(map[string]string)
	for _, m := range matches {
		// Skip if already resolved
		if _, ok := versions[m.ModulePath]; ok {
			continue
		}

		// Get root module and resolve version
		rootModule := detector.ExtractRootModule(m.ModulePath)
		version, err := res.LatestVersion(ctx, rootModule)
		if err != nil {
			logger.Warn("failed to resolve version", "module", m.ModulePath, "error", err)
			continue
		}

		versions[m.ModulePath] = version
		logger.Debug("resolved version", "module", m.ModulePath, "version", version)
	}

	// Rewrite
	result, err := rew.Rewrite(string(content), matches, versions)
	if err != nil {
		return nil, fmt.Errorf("rewrite: %w", err)
	}

	return result, nil
}

// setupLogger configures slog based on verbosity
func setupLogger(verbose bool) *slog.Logger {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}

	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	})

	return slog.New(handler)
}
