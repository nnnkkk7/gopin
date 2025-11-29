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
		cancel()   // Ensure cancel is called before exit
		os.Exit(1) //nolint:gocritic // exitAfterDefer: cancel() is explicitly called before Exit
	}
}
