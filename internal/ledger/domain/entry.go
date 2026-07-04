package domain

import shareddomain "github.com/mohsinazam/banking/internal/shared/domain"

// EntryID uniquely identifies a single ledger line within a transaction.
type EntryID string

func (id EntryID) String() string {
	return string(id)
}

// EntrySide is the direction of a posting. Amounts are always stored as positive values.
type EntrySide string

const (
	SideDebit  EntrySide = "DEBIT"
	SideCredit EntrySide = "CREDIT"
)

func (s EntrySide) isValid() bool {
	return s == SideDebit || s == SideCredit
}

// Entry is one immutable debit or credit line in the append-only ledger.
type Entry struct {
	id        EntryID
	accountID AccountID
	side      EntrySide
	amount    shareddomain.Money
}

// NewEntry constructs a validated ledger line. Amount must be strictly positive.
func NewEntry(id EntryID, accountID AccountID, side EntrySide, amount shareddomain.Money) (Entry, error) {
	if id == "" {
		return Entry{}, shareddomain.NewDomainError(ErrCodeInvalidLedgerEntry, "entry id is required")
	}
	if accountID == "" {
		return Entry{}, shareddomain.NewDomainError(ErrCodeInvalidLedgerEntry, "account id is required")
	}
	if !side.isValid() {
		return Entry{}, shareddomain.NewDomainError(ErrCodeInvalidLedgerEntry, "invalid entry side")
	}
	if !amount.IsPositive() {
		return Entry{}, shareddomain.NewDomainError(ErrCodeInvalidLedgerEntry, "entry amount must be positive")
	}
	return Entry{
		id:        id,
		accountID: accountID,
		side:      side,
		amount:    amount,
	}, nil
}

func (e Entry) ID() EntryID                 { return e.id }
func (e Entry) AccountID() AccountID         { return e.accountID }
func (e Entry) Side() EntrySide              { return e.side }
func (e Entry) Amount() shareddomain.Money   { return e.amount }
