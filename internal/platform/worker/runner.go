package worker

import (
	"context"
	"log/slog"
	"sync"
	"time"

	ledgerpostgres "github.com/mohsinazam/banking/internal/ledger/adapters/postgres"
	"github.com/mohsinazam/banking/internal/platform/config"
	"github.com/mohsinazam/banking/internal/platform/observability"
	outboxapp "github.com/mohsinazam/banking/internal/outbox/app"
	outboxpostgres "github.com/mohsinazam/banking/internal/outbox/adapters/postgres"
	uowpostgres "github.com/mohsinazam/banking/internal/platform/uow/postgres"
	settlementapp "github.com/mohsinazam/banking/internal/settlement/app"
	transferpostgres "github.com/mohsinazam/banking/internal/transfer/adapters/postgres"
)

// Runner executes outbox relay, settlement, and reconciliation loops.
type Runner struct {
	relay      *outboxapp.Relay
	settlement *settlementapp.SettlementProcessor
	reconcile  *settlementapp.Reconciler
	publisher  PublisherHandle
	logger     *slog.Logger
	interval   time.Duration
}

// NewRunner wires background jobs.
func NewRunner(
	cfg config.Config,
	ledgerRepo *ledgerpostgres.Repository,
	outboxRepo *outboxpostgres.Repository,
	sagaRepo *transferpostgres.SagaRepository,
	coordinator *uowpostgres.Coordinator,
	logger *slog.Logger,
) (*Runner, error) {
	pub, err := NewEventPublisher(cfg, logger)
	if err != nil {
		return nil, err
	}

	relayCfg := outboxapp.RelayConfig{BatchSize: 20, Interval: cfg.OutboxInterval}
	relay := outboxapp.NewRelay(outboxRepo, pub.Publisher, logger, relayCfg)
	sarieCfg := settlementapp.SarieConfig{
		FailRate: cfg.SarieFailRate,
		MinDelay: 50 * time.Millisecond,
		MaxDelay: 200 * time.Millisecond,
	}
	processor := settlementapp.NewSettlementProcessor(
		sagaRepo, coordinator, settlementapp.NewSarieMock(sarieCfg), ledgerRepo, logger,
	)
	reconciler := settlementapp.NewReconciler(ledgerRepo, logger)
	return &Runner{
		relay:      relay,
		settlement: processor,
		reconcile:  reconciler,
		publisher:  pub,
		logger:     logger,
		interval:   cfg.WorkerInterval,
	}, nil
}

// Run blocks until ctx is cancelled.
func (r *Runner) Run(ctx context.Context) {
	defer func() { _ = r.publisher.Close() }()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := r.relay.Run(ctx); err != nil && ctx.Err() == nil {
			r.logger.Error("outbox relay stopped", "error", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(r.interval)
		defer ticker.Stop()
		for {
			processed, err := r.settlement.ProcessOnce(ctx, 20)
			if err != nil {
				r.logger.Error("settlement batch failed", "error", err)
			} else if processed > 0 {
				observability.RecordSettlementsProcessed(processed)
				r.logger.Info("settlement batch processed", "count", processed)
			}
			if err := r.reconcile.Run(ctx); err != nil {
				r.logger.Error("reconciliation failed", "error", err)
			}
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()

	wg.Wait()
}
