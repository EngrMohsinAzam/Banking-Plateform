package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/mohsinazam/banking/internal/platform/config"
	platformpostgres "github.com/mohsinazam/banking/internal/platform/postgres"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("load config", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if len(os.Args) < 2 {
		logger.Error("usage: go run ./cmd/migrate [up|down]")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "up":
		if err := platformpostgres.RunMigrations(cfg.PostgresDSN); err != nil {
			logger.Error("migrate up failed", "error", err)
			os.Exit(1)
		}
		logger.Info("migrations applied")
	case "down":
		if err := platformpostgres.ResetMigrations(cfg.PostgresDSN); err != nil {
			logger.Error("migrate down failed", "error", err)
			os.Exit(1)
		}
		logger.Info("migrations rolled back")
	default:
		logger.Error("unknown command", "command", os.Args[1])
		os.Exit(1)
	}

	_ = ctx
}
