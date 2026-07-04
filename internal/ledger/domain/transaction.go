package domain

import (
	"time"

	shareddomain "github.com/mohsinazam/banking/internal/shared/domain"
)

// TransactionID uniquely identifies a journal transaction.
type TransactionID string

func (id TransactionID) String() string {
	return string(id)
}

// Transaction is an append-only, double-entry journal record.
// Invariants enforced at construction:
//   - at least two entries
//   - every entry amount is positive
//   - total debits == total credits (per currency)
type Transaction struct {
	id          TransactionID
	description string
	entries     []Entry
	recordedAt  time.Time
}

// NewTransaction validates and constructs a balanced journal transaction.
func NewTransaction(
	id TransactionID,
	description string,
	entries []Entry,
	recordedAt time.Time,
) (Transaction, error) {
	if id == "" {
		return Transaction{}, shareddomain.NewDomainError(ErrCodeInvalidLedgerEntry, "transaction id is required")
	}
	if len(entries) < 2 {
		return Transaction{}, shareddomain.NewDomainError(
			ErrCodeInvalidLedgerEntry,
			"transaction requires at least two entries",
		)
	}

	copied := make([]Entry, len(entries))
	copy(copied, entries)

	if err := validateBalanced(copied); err != nil {
		return Transaction{}, err
	}

	if recordedAt.IsZero() {
		recordedAt = time.Now().UTC()
	}

	return Transaction{
		id:          id,
		description: description,
		entries:     copied,
		recordedAt:  recordedAt.UTC(),
	}, nil
}

func (t Transaction) ID() TransactionID   { return t.id }
func (t Transaction) Description() string { return t.description }
func (t Transaction) RecordedAt() time.Time { return t.recordedAt }

// Entries returns a defensive copy of journal lines.
func (t Transaction) Entries() []Entry {
	out := make([]Entry, len(t.entries))
	copy(out, t.entries)
	return out
}

// TotalDebits sums all debit line amounts.
func (t Transaction) TotalDebits() (shareddomain.Money, error) {
	return sumBySide(t.entries, SideDebit)
}

// TotalCredits sums all credit line amounts.
func (t Transaction) TotalCredits() (shareddomain.Money, error) {
	return sumBySide(t.entries, SideCredit)
}

// IsBalanced reports whether debits equal credits.
func (t Transaction) IsBalanced() bool {
	return validateBalanced(t.entries) == nil
}

func validateBalanced(entries []Entry) error {
	debits, err := sumBySide(entries, SideDebit)
	if err != nil {
		return shareddomain.WrapDomainError(ErrCodeUnbalancedTransaction, "failed to sum debits", err)
	}
	credits, err := sumBySide(entries, SideCredit)
	if err != nil {
		return shareddomain.WrapDomainError(ErrCodeUnbalancedTransaction, "failed to sum credits", err)
	}

	cmp, err := debits.Cmp(credits)
	if err != nil {
		return shareddomain.WrapDomainError(ErrCodeUnbalancedTransaction, "currency mismatch in entries", err)
	}
	if cmp != 0 {
		return shareddomain.NewDomainError(
			ErrCodeUnbalancedTransaction,
			"debits must equal credits",
		)
	}
	return nil
}

func sumBySide(entries []Entry, side EntrySide) (shareddomain.Money, error) {
	total := shareddomain.ZeroSAR()
	for _, e := range entries {
		if e.Side() != side {
			continue
		}
		var err error
		total, err = total.Add(e.Amount())
		if err != nil {
			return shareddomain.Money{}, err
		}
	}
	return total, nil
}
