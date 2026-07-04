package domain

import (
	"encoding/json"

	outboxdomain "github.com/mohsinazam/banking/internal/outbox/domain"
)

// TransferPostedPayload is the outbox payload for transfer.posted.
type TransferPostedPayload struct {
	TransactionID   string `json:"transaction_id"`
	FromAccountID   string `json:"from_account_id"`
	ToAccountID     string `json:"to_account_id"`
	Amount          string `json:"amount"`
	BeneficiaryIBAN string `json:"beneficiary_iban"`
}

// TransferPostedEvent builds the outbox event emitted after a successful ledger post.
func TransferPostedEvent(cmd Command, result Result) (outboxdomain.Event, error) {
	payload, err := json.Marshal(TransferPostedPayload{
		TransactionID:   result.TransactionID.String(),
		FromAccountID:   result.FromAccountID.String(),
		ToAccountID:     result.ToAccountID.String(),
		Amount:          result.Amount.String(),
		BeneficiaryIBAN: cmd.BeneficiaryIBAN,
	})
	if err != nil {
		return outboxdomain.Event{}, err
	}

	return outboxdomain.NewEvent(
		outboxdomain.TransferPostedID(result.TransactionID.String()),
		outboxdomain.AggregateTransfer,
		result.TransactionID.String(),
		outboxdomain.EventTransferPosted,
		payload,
	)
}
