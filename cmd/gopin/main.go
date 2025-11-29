// gopin is a CLI tool to pin versions of go install commands.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/nnnkkk7/gopin/pkg/cli"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	app := cli.NewApp()
	if err := app.Run(ctx, os.Args); err != nil {
		slog.Error("gopin failed", "error", err)
		os.Exit(1)
	}
}
