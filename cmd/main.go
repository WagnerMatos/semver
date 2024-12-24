package main

import (
	"context"
	"log/slog"
	"os"
	"semver/internal/config"
	"semver/internal/tui"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()
	app := tui.New(cfg, logger)

	if err := app.Run(ctx); err != nil {
		logger.Error("application error", "error", err)
		os.Exit(1)
	}
}
