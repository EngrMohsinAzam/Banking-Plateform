package domain

import (
	"encoding/json"
	"time"
)

// SagaRecord is the persisted saga metadata.
type SagaRecord struct {
	ID             string
	State          SagaState
	IdempotencyKey string
	CommandJSON    []byte
	TransactionID  string
	SettlementID   string
	FailureReason  string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// SettlementRecord tracks mock sarie settlement for a saga.
type SettlementRecord struct {
	ID              string
	SagaID          string
	BeneficiaryIBAN string
	AmountHalalas   int64
	Currency        string
	Status          SettlementStatus
	Attempts        int
	LastError       string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// CommandJSON marshals a transfer command for saga persistence.
func CommandJSON(cmd Command) ([]byte, error) {
	payload := struct {
		IdempotencyKey  string `json:"idempotency_key"`
		FromAccountID   string `json:"from_account_id"`
		ToAccountID     string `json:"to_account_id"`
		Amount          string `json:"amount"`
		BeneficiaryIBAN string `json:"beneficiary_iban"`
		Description     string `json:"description"`
		BeneficiaryName string `json:"beneficiary_name"`
	}{
		IdempotencyKey:  cmd.IdempotencyKey,
		FromAccountID:   cmd.FromAccountID.String(),
		ToAccountID:     cmd.ToAccountID.String(),
		Amount:          cmd.Amount.String(),
		BeneficiaryIBAN: cmd.BeneficiaryIBAN,
		Description:     cmd.Description,
		BeneficiaryName: cmd.BeneficiaryName,
	}
	return json.Marshal(payload)
}
