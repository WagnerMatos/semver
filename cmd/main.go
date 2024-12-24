package main

import (
	"context"
	"fmt"
	"github.com/WagnerMatos/semver/internal/config"
	"github.com/WagnerMatos/semver/internal/tui"
	"log/slog"
	"os"
)

func run(ctx context.Context, logger *slog.Logger, testing bool) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	var app *tui.App
	if testing {
		app = tui.NewTest(cfg, logger)
	} else {
		app = tui.New(cfg, logger)
	}

	if err := app.Run(ctx); err != nil {
		return fmt.Errorf("application error: %w", err)
	}

	return nil
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	ctx := context.Background()

	if err := run(ctx, logger, false); err != nil {
		logger.Error("error running application", "error", err)
		os.Exit(1)
	}
}
