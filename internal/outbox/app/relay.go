package app

import (
	"context"
	"log/slog"
	"time"

	outboxpostgres "github.com/mohsinazam/banking/internal/outbox/adapters/postgres"
	"github.com/mohsinazam/banking/internal/outbox/ports"
	"github.com/mohsinazam/banking/internal/platform/observability"
)

// Relay polls the outbox and publishes pending events.
type Relay struct {
	repo      *outboxpostgres.Repository
	publisher ports.Publisher
	logger    *slog.Logger
	batchSize int
	interval  time.Duration
}

// RelayConfig configures the background relay loop.
type RelayConfig struct {
	BatchSize int
	Interval  time.Duration
}

// DefaultRelayConfig returns sensible worker defaults.
func DefaultRelayConfig() RelayConfig {
	return RelayConfig{BatchSize: 20, Interval: 2 * time.Second}
}

// NewRelay constructs an outbox relay.
func NewRelay(repo *outboxpostgres.Repository, publisher ports.Publisher, logger *slog.Logger, cfg RelayConfig) *Relay {
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = DefaultRelayConfig().BatchSize
	}
	if cfg.Interval <= 0 {
		cfg.Interval = DefaultRelayConfig().Interval
	}
	return &Relay{
		repo:      repo,
		publisher: publisher,
		logger:    logger,
		batchSize: cfg.BatchSize,
		interval:  cfg.Interval,
	}
}

// Run blocks until ctx is cancelled, polling and publishing outbox rows.
func (r *Relay) Run(ctx context.Context) error {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	r.logger.Info("outbox relay started", "interval", r.interval.String(), "batch_size", r.batchSize)

	for {
		processed, err := r.repo.ProcessPendingBatch(ctx, r.batchSize, r.publisher.Publish)
		if err != nil {
			r.logger.Error("outbox batch failed", "error", err)
		} else if processed > 0 {
			observability.RecordOutboxRelayed(processed)
			r.logger.Info("outbox batch relayed", "count", processed)
		}

		select {
		case <-ctx.Done():
			r.logger.Info("outbox relay stopping")
			return nil
		case <-ticker.C:
		}
	}
}

// ProcessOnce processes a single batch (useful in tests).
func (r *Relay) ProcessOnce(ctx context.Context) (int, error) {
	return r.repo.ProcessPendingBatch(ctx, r.batchSize, r.publisher.Publish)
}
