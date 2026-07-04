package domain

import (
	"time"

	shareddomain "github.com/mohsinazam/banking/internal/shared/domain"
	ledgerdomain "github.com/mohsinazam/banking/internal/ledger/domain"
)

// Status is the read model for transfer + saga + settlement state.
type Status struct {
	SagaID           string
	SagaState        SagaState
	IdempotencyKey   string
	TransactionID    string
	SettlementID     string
	SettlementStatus SettlementStatus
	SettlementError  string
	FailureReason    string
	FromAccountID    ledgerdomain.AccountID
	ToAccountID      ledgerdomain.AccountID
	Amount           shareddomain.Money
	BeneficiaryIBAN  string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
