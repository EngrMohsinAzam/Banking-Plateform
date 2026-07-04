package app

import (
	"encoding/json"
	"fmt"

	idempotencydomain "github.com/mohsinazam/banking/internal/idempotency/domain"
	ledgerdomain "github.com/mohsinazam/banking/internal/ledger/domain"
	shareddomain "github.com/mohsinazam/banking/internal/shared/domain"
	"github.com/mohsinazam/banking/internal/transfer/domain"
)

func encodeTransferResult(result domain.Result) ([]byte, error) {
	payload := struct {
		TransactionID    string `json:"transaction_id"`
		FromAccountID    string `json:"from_account_id"`
		ToAccountID      string `json:"to_account_id"`
		Amount           string `json:"amount"`
		SagaID           string `json:"saga_id"`
		SettlementID     string `json:"settlement_id"`
		SagaState        string `json:"saga_state"`
		SettlementStatus string `json:"settlement_status"`
	}{
		TransactionID:    result.TransactionID.String(),
		FromAccountID:    result.FromAccountID.String(),
		ToAccountID:      result.ToAccountID.String(),
		Amount:           result.Amount.String(),
		SagaID:           result.SagaID,
		SettlementID:     result.SettlementID,
		SagaState:        string(result.SagaState),
		SettlementStatus: string(result.SettlementStatus),
	}
	return json.Marshal(payload)
}

func decodeTransferResult(idem idempotencydomain.Result) (domain.Result, error) {
	var payload struct {
		TransactionID    string `json:"transaction_id"`
		FromAccountID    string `json:"from_account_id"`
		ToAccountID      string `json:"to_account_id"`
		Amount           string `json:"amount"`
		SagaID           string `json:"saga_id"`
		SettlementID     string `json:"settlement_id"`
		SagaState        string `json:"saga_state"`
		SettlementStatus string `json:"settlement_status"`
	}
	if err := json.Unmarshal(idem.Payload, &payload); err != nil {
		return domain.Result{}, fmt.Errorf("decode transfer result: %w", err)
	}

	amount, err := shareddomain.ParseSAR(payload.Amount)
	if err != nil {
		return domain.Result{}, err
	}

	return domain.Result{
		TransactionID:    ledgerdomain.TransactionID(payload.TransactionID),
		FromAccountID:    ledgerdomain.AccountID(payload.FromAccountID),
		ToAccountID:      ledgerdomain.AccountID(payload.ToAccountID),
		Amount:           amount,
		SagaID:           payload.SagaID,
		SettlementID:     payload.SettlementID,
		SagaState:        domain.SagaState(payload.SagaState),
		SettlementStatus: domain.SettlementStatus(payload.SettlementStatus),
	}, nil
}
