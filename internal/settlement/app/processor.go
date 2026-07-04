package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	ledgerdomain "github.com/mohsinazam/banking/internal/ledger/domain"
	shareddomain "github.com/mohsinazam/banking/internal/shared/domain"
	uowpostgres "github.com/mohsinazam/banking/internal/platform/uow/postgres"
	settlementports "github.com/mohsinazam/banking/internal/settlement/ports"
	"github.com/mohsinazam/banking/internal/transfer/adapters/postgres"
	"github.com/mohsinazam/banking/internal/transfer/domain"
	"github.com/mohsinazam/banking/internal/transfer/ports"
)

// SettlementProcessor drives mock sarie settlement and compensation.
type SettlementProcessor struct {
	sagaStore   ports.SagaStore
	coordinator *uowpostgres.Coordinator
	sarie       settlementports.SettlementRail
	accounts    ports.AccountReader
	logger      *slog.Logger
}

// NewSettlementProcessor constructs a settlement worker component.
func NewSettlementProcessor(
	sagaStore ports.SagaStore,
	coordinator *uowpostgres.Coordinator,
	sarie settlementports.SettlementRail,
	accounts ports.AccountReader,
	logger *slog.Logger,
) *SettlementProcessor {
	return &SettlementProcessor{
		sagaStore:   sagaStore,
		coordinator: coordinator,
		sarie:       sarie,
		accounts:    accounts,
		logger:      logger,
	}
}

// ProcessOnce claims pending settlements and drives sarie mock + compensation.
func (p *SettlementProcessor) ProcessOnce(ctx context.Context, limit int) (int, error) {
	pending, err := p.sagaStore.ClaimPendingSettlements(ctx, limit)
	if err != nil {
		return 0, err
	}

	processed := 0
	for _, stl := range pending {
		if err := p.sagaStore.UpdateSagaState(ctx, stl.SagaID, domain.SagaStateSettling, nil); err != nil {
			p.logger.Error("failed to update saga state", "saga_id", stl.SagaID, "state", domain.SagaStateSettling, "error", err)
		}

		result, err := p.sarie.Settle(ctx, stl.ID)
		if err != nil {
			return processed, err
		}
		if result.Success {
			if err := p.sagaStore.UpdateSettlementStatus(ctx, stl.ID, domain.SettlementSettled, ""); err != nil {
				return processed, err
			}
			if err := p.sagaStore.UpdateSagaState(ctx, stl.SagaID, domain.SagaStateCompleted, nil); err != nil {
				p.logger.Error("failed to mark saga completed", "saga_id", stl.SagaID, "error", err)
			}
			p.logger.Info("settlement completed", "settlement_id", stl.ID, "delay", result.Delay.String())
			processed++
			continue
		}

		if err := p.compensate(ctx, stl); err != nil {
			p.logger.Error("compensation failed", "settlement_id", stl.ID, "error", err)
			if err := p.sagaStore.UpdateSettlementStatus(ctx, stl.ID, domain.SettlementFailed, result.Error); err != nil {
				p.logger.Error("failed to update settlement status", "settlement_id", stl.ID, "error", err)
			}
			if err := p.sagaStore.UpdateSagaState(ctx, stl.SagaID, domain.SagaStateFailed, map[string]string{"failure_reason": result.Error}); err != nil {
				p.logger.Error("failed to mark saga failed", "saga_id", stl.SagaID, "error", err)
			}
			continue
		}
		if err := p.sagaStore.UpdateSettlementStatus(ctx, stl.ID, domain.SettlementFailed, result.Error); err != nil {
			p.logger.Error("failed to update settlement status", "settlement_id", stl.ID, "error", err)
		}
		p.logger.Warn("settlement failed; compensated", "settlement_id", stl.ID, "error", result.Error)
		processed++
	}
	return processed, nil
}

func (p *SettlementProcessor) compensate(ctx context.Context, stl domain.SettlementRecord) error {
	saga, err := p.sagaStore.GetSaga(ctx, stl.SagaID)
	if err != nil {
		return err
	}
	cmd, err := postgres.CommandFromJSON(saga.CommandJSON)
	if err != nil {
		return err
	}

	amount, err := shareddomain.HalalasFromMinorUnits(stl.AmountHalalas, stl.Currency)
	if err != nil {
		return err
	}

	compTxID := ledgerdomain.TransactionID(fmt.Sprintf("%s-comp", saga.TransactionID))
	debit, err := ledgerdomain.NewEntry(domain.EntryID(compTxID, 1), cmd.ToAccountID, ledgerdomain.SideDebit, amount)
	if err != nil {
		return err
	}
	credit, err := ledgerdomain.NewEntry(domain.EntryID(compTxID, 2), cmd.FromAccountID, ledgerdomain.SideCredit, amount)
	if err != nil {
		return err
	}

	tx, err := ledgerdomain.NewTransaction(compTxID, "compensate failed sarie settlement",
		[]ledgerdomain.Entry{debit, credit}, time.Now().UTC())
	if err != nil {
		return err
	}

	_ = p.sagaStore.UpdateSagaState(ctx, stl.SagaID, domain.SagaStateCompensating, nil)
	return p.coordinator.PostCompensation(ctx, tx, stl.SagaID)
}
