package domain

import (
	"encoding/json"
	"fmt"

	shareddomain "github.com/mohsinazam/banking/internal/shared/domain"
	ledgersdomain "github.com/mohsinazam/banking/internal/ledger/domain"
	idempotencydomain "github.com/mohsinazam/banking/internal/idempotency/domain"
)

const TransferScope = idempotencydomain.Scope("transfer")

// Command is the inbound transfer request for the happy path.
type Command struct {
	IdempotencyKey  string
	FromAccountID   ledgersdomain.AccountID
	ToAccountID     ledgersdomain.AccountID
	Amount          shareddomain.Money
	BeneficiaryIBAN string
	BeneficiaryName string
	Description     string
}

// Result is returned after a successful (or replayed) transfer.
type Result struct {
	TransactionID    ledgersdomain.TransactionID
	FromAccountID    ledgersdomain.AccountID
	ToAccountID      ledgersdomain.AccountID
	Amount           shareddomain.Money
	SagaID           string
	SettlementID     string
	SagaState        SagaState
	SettlementStatus SettlementStatus
	Replayed         bool
}

// Validate checks business rules before idempotency or ledger side effects.
func (c Command) Validate() error {
	key, err := idempotencydomain.ParseKey(c.IdempotencyKey)
	if err != nil {
		return err
	}
	_ = key

	if c.FromAccountID == "" || c.ToAccountID == "" {
		return shareddomain.NewDomainError(shareddomain.ErrCodeValidation, "from and to account are required")
	}
	if c.FromAccountID == c.ToAccountID {
		return shareddomain.NewDomainError(shareddomain.ErrCodeValidation, "cannot transfer to the same account")
	}
	if !c.Amount.IsPositive() {
		return shareddomain.NewDomainError(shareddomain.ErrCodeInvalidMoney, "transfer amount must be positive")
	}
	if c.BeneficiaryIBAN == "" {
		return shareddomain.NewDomainError(shareddomain.ErrCodeValidation, "beneficiary IBAN is required")
	}
	if _, err := shareddomain.ParseSAIBAN(c.BeneficiaryIBAN); err != nil {
		return err
	}
	return nil
}

// Fingerprint returns a stable hash of the transfer intent for idempotency.
func (c Command) Fingerprint() idempotencydomain.Fingerprint {
	payload := struct {
		FromAccountID   string `json:"from_account_id"`
		ToAccountID     string `json:"to_account_id"`
		AmountHalalas   int64  `json:"amount_halalas"`
		Currency        string `json:"currency"`
		BeneficiaryIBAN string `json:"beneficiary_iban"`
		BeneficiaryName string `json:"beneficiary_name"`
	}{
		FromAccountID:   c.FromAccountID.String(),
		ToAccountID:     c.ToAccountID.String(),
		AmountHalalas:   c.Amount.Halalas(),
		Currency:        c.Amount.Currency(),
		BeneficiaryIBAN: normalizeIBAN(c.BeneficiaryIBAN),
		BeneficiaryName: c.BeneficiaryName,
	}
	data, _ := json.Marshal(payload)
	return idempotencydomain.FingerprintFromPayload(data)
}

// ParsedKey parses the command idempotency key.
func (c Command) ParsedKey() (idempotencydomain.Key, error) {
	return idempotencydomain.ParseKey(c.IdempotencyKey)
}

// LedgerTransactionID derives the deterministic journal id for this transfer.
func (c Command) LedgerTransactionID() (ledgersdomain.TransactionID, error) {
	key, err := c.ParsedKey()
	if err != nil {
		return "", err
	}
	return ledgersdomain.TransactionID(idempotencydomain.LedgerTransactionID(TransferScope, key)), nil
}

func normalizeIBAN(raw string) string {
	iban, err := shareddomain.ParseSAIBAN(raw)
	if err != nil {
		return raw
	}
	return iban.String()
}

// EntryID builds a deterministic entry id for a transaction leg.
func EntryID(txID ledgersdomain.TransactionID, leg int) ledgersdomain.EntryID {
	return ledgersdomain.EntryID(fmt.Sprintf("%s-leg%d", txID, leg))
}
