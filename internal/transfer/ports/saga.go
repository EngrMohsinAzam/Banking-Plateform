package ports

import (
	"context"

	shareddomain "github.com/mohsinazam/banking/internal/shared/domain"
	transferdomain "github.com/mohsinazam/banking/internal/transfer/domain"
)

// FraudChecker screens transfers before ledger effects.
type FraudChecker interface {
	Check(ctx context.Context, input FraudCheckInput) error
}

// FraudCheckInput is the fraud port DTO.
type FraudCheckInput struct {
	FromAccountID string
	Amount        shareddomain.Money
}

// ComplianceChecker screens beneficiaries against sanctions lists.
type ComplianceChecker interface {
	Check(ctx context.Context, input ComplianceCheckInput) error
}

// ComplianceCheckInput is the compliance port DTO.
type ComplianceCheckInput struct {
	BeneficiaryIBAN string
	BeneficiaryName string
}

// SagaStore persists saga and settlement state.
type SagaStore interface {
	CreateSaga(ctx context.Context, saga transferdomain.SagaRecord) error
	UpdateSagaState(ctx context.Context, sagaID string, state transferdomain.SagaState, fields map[string]string) error
	GetSagaByIdempotencyKey(ctx context.Context, key string) (transferdomain.SagaRecord, error)
	GetSagaByTransactionID(ctx context.Context, transactionID string) (transferdomain.SagaRecord, error)
	GetSaga(ctx context.Context, id string) (transferdomain.SagaRecord, error)
	GetSettlementBySagaID(ctx context.Context, sagaID string) (transferdomain.SettlementRecord, error)
	CreateSettlement(ctx context.Context, settlement transferdomain.SettlementRecord) error
	ClaimPendingSettlements(ctx context.Context, limit int) ([]transferdomain.SettlementRecord, error)
	UpdateSettlementStatus(ctx context.Context, id string, status transferdomain.SettlementStatus, lastError string) error
}
