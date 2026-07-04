package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	fraudapp "github.com/mohsinazam/banking/internal/fraud/app"
	fraudredis "github.com/mohsinazam/banking/internal/fraud/adapters/redis"
	idempotencyapp "github.com/mohsinazam/banking/internal/idempotency/app"
	idempotencyredis "github.com/mohsinazam/banking/internal/idempotency/adapters/redis"
	complianceapp "github.com/mohsinazam/banking/internal/compliance/app"
	ledgerpostgres "github.com/mohsinazam/banking/internal/ledger/adapters/postgres"
	ledgerapp "github.com/mohsinazam/banking/internal/ledger/app"
	"github.com/mohsinazam/banking/internal/platform/audit"
	"github.com/mohsinazam/banking/internal/platform/auth"
	"github.com/mohsinazam/banking/internal/platform/config"
	platformhttp "github.com/mohsinazam/banking/internal/platform/http"
	platformpostgres "github.com/mohsinazam/banking/internal/platform/postgres"
	platformredis "github.com/mohsinazam/banking/internal/platform/redis"
	uowpostgres "github.com/mohsinazam/banking/internal/platform/uow/postgres"
	outboxpostgres "github.com/mohsinazam/banking/internal/outbox/adapters/postgres"
	transferapp "github.com/mohsinazam/banking/internal/transfer/app"
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

	ctx := context.Background()
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

	redisClient := platformredis.NewClient(cfg.RedisAddr)
	defer func() { _ = redisClient.Close() }()
	if err := platformredis.Ping(ctx, redisClient); err != nil {
		logger.Error("redis connection failed", "error", err)
		os.Exit(1)
	}
	logger.Info("dependencies ready", "postgres", "ok", "redis", "ok")

	ledgerRepo := ledgerpostgres.NewRepository(pool)
	ledgerService := ledgerapp.NewPoster(ledgerRepo)
	outboxRepo := outboxpostgres.NewRepository(pool)
	sagaRepo := transferpostgres.NewSagaRepository(pool)
	coordinator := uowpostgres.NewCoordinator(pool, ledgerRepo, outboxRepo, sagaRepo, logger)

	fraudCfg := fraudapp.Config{
		MaxSingleHalalas:   cfg.FraudMaxSingleHalalas,
		MaxHourlyTransfers: cfg.FraudMaxHourlyTransfers,
	}
	transferService := transferapp.NewService(
		ledgerRepo,
		coordinator,
		idempotencyapp.NewGuard(idempotencyredis.NewStore(redisClient), idempotencyapp.DefaultConfig()),
		fraudapp.NewChecker(fraudredis.NewVelocityStore(redisClient), fraudCfg),
		complianceapp.NewChecker(),
		sagaRepo,
		logger,
	)

	server := platformhttp.NewServer(platformhttp.Dependencies{
		Config:          cfg,
		Logger:          logger,
		Pool:            pool,
		RedisClient:     redisClient,
		LedgerService:   ledgerService,
		TransferService: transferService,
		AuditLogger:     audit.NewLogger(pool),
		APIKeyAuth:      auth.NewAPIKeyValidator(cfg.AuthEnabled, cfg.APIKeys),
	})

	errCh := make(chan error, 1)
	go func() { errCh <- server.ListenAndServe() }()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		logger.Info("shutdown signal received", "signal", sig.String())
	case err := <-errCh:
		if err != nil {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
		return
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}
	logger.Info("server stopped")
}
