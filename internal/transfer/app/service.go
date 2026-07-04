package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"time"

	idempotencydomain "github.com/mohsinazam/banking/internal/idempotency/domain"
	ledgerdomain "github.com/mohsinazam/banking/internal/ledger/domain"
	outboxdomain "github.com/mohsinazam/banking/internal/outbox/domain"
	shareddomain "github.com/mohsinazam/banking/internal/shared/domain"
	uowpostgres "github.com/mohsinazam/banking/internal/platform/uow/postgres"
	"github.com/mohsinazam/banking/internal/transfer/adapters/postgres"
	"github.com/mohsinazam/banking/internal/transfer/domain"
	"github.com/mohsinazam/banking/internal/transfer/ports"
)

// Service orchestrates the full transfer saga: fraud → compliance → ledger → outbox → settlement.
type Service struct {
	accounts    ports.AccountReader
	coordinator ports.TransferCoordinator
	idempotency ports.Idempotency
	fraud       ports.FraudChecker
	compliance  ports.ComplianceChecker
	sagaStore   ports.SagaStore
	logger      *slog.Logger
	clock       func() time.Time
}

// NewService constructs the saga-aware transfer service.
func NewService(
	accounts ports.AccountReader,
	coordinator ports.TransferCoordinator,
	idempotency ports.Idempotency,
	fraud ports.FraudChecker,
	compliance ports.ComplianceChecker,
	sagaStore ports.SagaStore,
	logger *slog.Logger,
) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	return &Service{
		accounts:    accounts,
		coordinator: coordinator,
		idempotency: idempotency,
		fraud:       fraud,
		compliance:  compliance,
		sagaStore:   sagaStore,
		logger:      logger,
		clock:       time.Now,
	}
}

// Execute runs the transfer saga behind idempotency protection.
func (s *Service) Execute(ctx context.Context, cmd domain.Command) (domain.Result, error) {
	if err := cmd.Validate(); err != nil {
		return domain.Result{}, err
	}

	key, err := cmd.ParsedKey()
	if err != nil {
		return domain.Result{}, err
	}

	fingerprint := cmd.Fingerprint()
	s.logger.Info("transfer requested",
		"idempotency_key", cmd.IdempotencyKey,
		"from_account", cmd.FromAccountID.String(),
		"to_account", cmd.ToAccountID.String(),
		"amount", cmd.Amount.String(),
	)

	idemResult, replay, err := s.idempotency.Run(ctx, domain.TransferScope, key, fingerprint, func(ctx context.Context) (idempotencydomain.Result, error) {
		return s.runSaga(ctx, cmd)
	})
	if err != nil {
		s.logger.Warn("transfer failed", "idempotency_key", cmd.IdempotencyKey, "error", err)
		return domain.Result{}, err
	}

	result, err := decodeTransferResult(idemResult)
	if err != nil {
		return domain.Result{}, err
	}
	result.Replayed = replay

	s.logger.Info("transfer completed",
		"idempotency_key", cmd.IdempotencyKey,
		"transaction_id", result.TransactionID.String(),
		"saga_id", result.SagaID,
		"replayed", replay,
	)
	return result, nil
}

// GetStatusByTransactionID returns saga + settlement state for a posted transfer.
func (s *Service) GetStatusByTransactionID(ctx context.Context, transactionID string) (domain.Status, error) {
	saga, err := s.sagaStore.GetSagaByTransactionID(ctx, transactionID)
	if err != nil {
		return domain.Status{}, err
	}
	return s.statusFromSaga(ctx, saga)
}

// GetStatusByIdempotencyKey returns saga + settlement state by client idempotency key.
func (s *Service) GetStatusByIdempotencyKey(ctx context.Context, key string) (domain.Status, error) {
	saga, err := s.sagaStore.GetSagaByIdempotencyKey(ctx, key)
	if err != nil {
		return domain.Status{}, err
	}
	return s.statusFromSaga(ctx, saga)
}

func (s *Service) statusFromSaga(ctx context.Context, saga domain.SagaRecord) (domain.Status, error) {
	cmd, err := postgres.CommandFromJSON(saga.CommandJSON)
	if err != nil {
		return domain.Status{}, fmt.Errorf("decode saga command: %w", err)
	}

	status := domain.Status{
		SagaID:          saga.ID,
		SagaState:       saga.State,
		IdempotencyKey:  saga.IdempotencyKey,
		TransactionID:   saga.TransactionID,
		SettlementID:    saga.SettlementID,
		FailureReason:   saga.FailureReason,
		FromAccountID:   cmd.FromAccountID,
		ToAccountID:     cmd.ToAccountID,
		Amount:          cmd.Amount,
		BeneficiaryIBAN: cmd.BeneficiaryIBAN,
		CreatedAt:       saga.CreatedAt,
		UpdatedAt:       saga.UpdatedAt,
	}

	settlement, err := s.sagaStore.GetSettlementBySagaID(ctx, saga.ID)
	if err == nil {
		status.SettlementStatus = settlement.Status
		status.SettlementError = settlement.LastError
	} else if !shareddomain.IsDomainCode(err, shareddomain.ErrCodeNotFound) {
		return domain.Status{}, err
	}

	return status, nil
}

func (s *Service) runSaga(ctx context.Context, cmd domain.Command) (idempotencydomain.Result, error) {
	sagaID := sagaIDFromKey(cmd.IdempotencyKey)
	now := s.clock().UTC()

	s.logger.Debug("saga fraud check", "saga_id", sagaID)
	if err := s.fraud.Check(ctx, ports.FraudCheckInput{
		FromAccountID: cmd.FromAccountID.String(),
		Amount:        cmd.Amount,
	}); err != nil {
		return idempotencydomain.Result{}, err
	}

	s.logger.Debug("saga compliance check", "saga_id", sagaID)
	if err := s.compliance.Check(ctx, ports.ComplianceCheckInput{
		BeneficiaryIBAN: cmd.BeneficiaryIBAN,
		BeneficiaryName: cmd.BeneficiaryName,
	}); err != nil {
		return idempotencydomain.Result{}, err
	}

	fromAccount, err := s.accounts.GetAccount(ctx, cmd.FromAccountID)
	if err != nil {
		return idempotencydomain.Result{}, err
	}
	toAccount, err := s.accounts.GetAccount(ctx, cmd.ToAccountID)
	if err != nil {
		return idempotencydomain.Result{}, err
	}
	if fromAccount.Type() != ledgerdomain.AccountTypeLiability || toAccount.Type() != ledgerdomain.AccountTypeLiability {
		return idempotencydomain.Result{}, shareddomain.NewDomainError(shareddomain.ErrCodeValidation, "wallet transfers require liability accounts")
	}

	txID, err := cmd.LedgerTransactionID()
	if err != nil {
		return idempotencydomain.Result{}, err
	}

	debitEntry, err := ledgerdomain.NewEntry(domain.EntryID(txID, 1), cmd.FromAccountID, ledgerdomain.SideDebit, cmd.Amount)
	if err != nil {
		return idempotencydomain.Result{}, err
	}
	creditEntry, err := ledgerdomain.NewEntry(domain.EntryID(txID, 2), cmd.ToAccountID, ledgerdomain.SideCredit, cmd.Amount)
	if err != nil {
		return idempotencydomain.Result{}, err
	}

	description := cmd.Description
	if description == "" {
		description = fmt.Sprintf("transfer to %s", cmd.BeneficiaryIBAN)
	}

	ledgerTx, err := ledgerdomain.NewTransaction(txID, description, []ledgerdomain.Entry{debitEntry, creditEntry}, now)
	if err != nil {
		return idempotencydomain.Result{}, err
	}

	result := domain.Result{
		TransactionID:    txID,
		FromAccountID:    cmd.FromAccountID,
		ToAccountID:      cmd.ToAccountID,
		Amount:           cmd.Amount,
		SagaID:           sagaID,
		SettlementID:     fmt.Sprintf("stl_%s", txID),
		SagaState:        domain.SagaStatePosted,
		SettlementStatus: domain.SettlementPending,
	}

	event, err := domain.TransferPostedEvent(cmd, result)
	if err != nil {
		return idempotencydomain.Result{}, err
	}

	cmdJSON, err := domain.CommandJSON(cmd)
	if err != nil {
		return idempotencydomain.Result{}, err
	}

	sagaRec := domain.SagaRecord{
		ID:             sagaID,
		State:          domain.SagaStateComplianceOK,
		IdempotencyKey: cmd.IdempotencyKey,
		CommandJSON:    cmdJSON,
		TransactionID:  txID.String(),
		SettlementID:   result.SettlementID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	settlement := domain.SettlementRecord{
		ID:              result.SettlementID,
		SagaID:          sagaID,
		BeneficiaryIBAN: cmd.BeneficiaryIBAN,
		AmountHalalas:   cmd.Amount.Halalas(),
		Currency:        cmd.Amount.Currency(),
		Status:          domain.SettlementPending,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	s.logger.Info("posting transfer atomically",
		"saga_id", sagaID,
		"transaction_id", txID.String(),
		"settlement_id", result.SettlementID,
	)
	if err := s.coordinator.PostTransfer(ctx, uowpostgres.TransferCommit{
		LedgerTx:   ledgerTx,
		Events:     []outboxdomain.Event{event},
		Saga:       sagaRec,
		Settlement: settlement,
	}); err != nil {
		return idempotencydomain.Result{}, err
	}

	payload, err := encodeTransferResult(result)
	if err != nil {
		return idempotencydomain.Result{}, err
	}
	return idempotencydomain.Result{ResourceID: txID.String(), Payload: payload}, nil
}

func sagaIDFromKey(key string) string {
	sum := sha256.Sum256([]byte("saga:" + key))
	return "saga_" + hex.EncodeToString(sum[:12])
}
