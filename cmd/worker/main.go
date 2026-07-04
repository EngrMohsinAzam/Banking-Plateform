package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	ledgerpostgres "github.com/mohsinazam/banking/internal/ledger/adapters/postgres"
	"github.com/mohsinazam/banking/internal/platform/config"
	platformpostgres "github.com/mohsinazam/banking/internal/platform/postgres"
	platformworker "github.com/mohsinazam/banking/internal/platform/worker"
	outboxpostgres "github.com/mohsinazam/banking/internal/outbox/adapters/postgres"
	uowpostgres "github.com/mohsinazam/banking/internal/platform/uow/postgres"
	transferpostgres "github.com/mohsinazam/banking/internal/transfer/adapters/postgres"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := platformpostgres.NewPool(ctx, cfg.PostgresDSN)
	if err != nil {
		logger.Error("postgres connection failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if cfg.AutoMigrate {
		if err := platformpostgres.RunMigrations(cfg.PostgresDSN); err != nil {
			logger.Error("migrations failed", "error", err)
			os.Exit(1)
		}
		logger.Info("migrations applied")
	}

	ledgerRepo := ledgerpostgres.NewRepository(pool)
	outboxRepo := outboxpostgres.NewRepository(pool)
	sagaRepo := transferpostgres.NewSagaRepository(pool)
	coordinator := uowpostgres.NewCoordinator(pool, ledgerRepo, outboxRepo, sagaRepo, logger)

	runner, err := platformworker.NewRunner(cfg, ledgerRepo, outboxRepo, sagaRepo, coordinator, logger)
	if err != nil {
		logger.Error("worker wiring failed", "error", err)
		os.Exit(1)
	}
	logger.Info("worker starting",
		"jobs", "outbox,settlement,reconciliation",
		"interval", cfg.WorkerInterval.String(),
		"outbox_interval", cfg.OutboxInterval.String(),
		"sarie_fail_rate", cfg.SarieFailRate,
	)
	runner.Run(ctx)
	logger.Info("worker stopped")
}
