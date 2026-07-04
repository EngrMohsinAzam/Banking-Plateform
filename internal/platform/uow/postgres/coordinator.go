package postgres

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	ledgerpostgres "github.com/mohsinazam/banking/internal/ledger/adapters/postgres"
	ledgerdomain "github.com/mohsinazam/banking/internal/ledger/domain"
	outboxdomain "github.com/mohsinazam/banking/internal/outbox/domain"
	outboxpostgres "github.com/mohsinazam/banking/internal/outbox/adapters/postgres"
	transferdomain "github.com/mohsinazam/banking/internal/transfer/domain"
	transferpostgres "github.com/mohsinazam/banking/internal/transfer/adapters/postgres"
)

// Coordinator commits ledger writes, outbox events, saga, and settlement atomically.
type Coordinator struct {
	pool   *pgxpool.Pool
	ledger *ledgerpostgres.Repository
	outbox *outboxpostgres.Repository
	saga   *transferpostgres.SagaRepository
	logger *slog.Logger
}

// NewCoordinator wires adapters against a shared pool.
func NewCoordinator(
	pool *pgxpool.Pool,
	ledger *ledgerpostgres.Repository,
	outbox *outboxpostgres.Repository,
	saga *transferpostgres.SagaRepository,
	logger *slog.Logger,
) *Coordinator {
	if logger == nil {
		logger = slog.Default()
	}
	return &Coordinator{pool: pool, ledger: ledger, outbox: outbox, saga: saga, logger: logger}
}

// TransferCommit bundles all writes for a successful transfer post.
type TransferCommit struct {
	LedgerTx   ledgerdomain.Transaction
	Events     []outboxdomain.Event
	Saga       transferdomain.SagaRecord
	Settlement transferdomain.SettlementRecord
}

// PostTransfer atomically persists ledger, outbox, saga, and settlement rows.
func (c *Coordinator) PostTransfer(ctx context.Context, commit TransferCommit) error {
	dbTx, err := c.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = dbTx.Rollback(ctx)
	}()

	if err := c.saga.InsertSagaInTx(ctx, dbTx, commit.Saga); err != nil {
		return err
	}
	if err := c.ledger.PostTransactionInTx(ctx, dbTx, commit.LedgerTx); err != nil {
		return err
	}
	for _, event := range commit.Events {
		if err := c.outbox.InsertInTx(ctx, dbTx, event); err != nil {
			return err
		}
	}
	if err := c.saga.InsertSettlementInTx(ctx, dbTx, commit.Settlement); err != nil {
		return err
	}
	if err := c.saga.UpdateSagaStateInTx(ctx, dbTx, commit.Saga.ID, transferdomain.SagaStatePosted, map[string]string{
		"transaction_id": commit.Saga.TransactionID,
		"settlement_id":  commit.Saga.SettlementID,
	}); err != nil {
		return err
	}

	if err := dbTx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	c.logger.Info("transfer committed",
		"saga_id", commit.Saga.ID,
		"transaction_id", commit.Saga.TransactionID,
		"settlement_id", commit.Saga.SettlementID,
	)
	return nil
}

// PostTransactionWithEvents posts journal + outbox only (legacy/tests).
func (c *Coordinator) PostTransactionWithEvents(
	ctx context.Context,
	tx ledgerdomain.Transaction,
	events []outboxdomain.Event,
) error {
	dbTx, err := c.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = dbTx.Rollback(ctx)
	}()

	if err := c.ledger.PostTransactionInTx(ctx, dbTx, tx); err != nil {
		return err
	}
	for _, event := range events {
		if err := c.outbox.InsertInTx(ctx, dbTx, event); err != nil {
			return err
		}
	}
	if err := dbTx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

// PostCompensation posts a reversing journal transaction atomically.
func (c *Coordinator) PostCompensation(ctx context.Context, tx ledgerdomain.Transaction, sagaID string) error {
	dbTx, err := c.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = dbTx.Rollback(ctx)
	}()

	if err := c.ledger.PostTransactionInTx(ctx, dbTx, tx); err != nil {
		return err
	}
	if err := c.saga.UpdateSagaStateInTx(ctx, dbTx, sagaID, transferdomain.SagaStateCompensated, map[string]string{
		"failure_reason": "sarie settlement failed; funds reversed",
	}); err != nil {
		return err
	}
	if err := dbTx.Commit(ctx); err != nil {
		return fmt.Errorf("commit compensation: %w", err)
	}
	c.logger.Warn("compensation committed", "saga_id", sagaID, "transaction_id", tx.ID().String())
	return nil
}
