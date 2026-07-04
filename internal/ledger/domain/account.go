package domain

import shareddomain "github.com/mohsinazam/banking/internal/shared/domain"

// AccountID uniquely identifies a ledger account.
type AccountID string

func (id AccountID) String() string {
	return string(id)
}

// AccountType determines how debits and credits affect the derived balance.
// Banking customer wallets are LIABILITIES (the bank owes the customer).
type AccountType string

const (
	AccountTypeAsset     AccountType = "ASSET"
	AccountTypeLiability AccountType = "LIABILITY"
	AccountTypeEquity    AccountType = "EQUITY"
	AccountTypeRevenue   AccountType = "REVENUE"
	AccountTypeExpense   AccountType = "EXPENSE"
)

// Account is a ledger account metadata record.
// Balance is never stored here — it is derived from append-only entries.
type Account struct {
	id   AccountID
	typ  AccountType
	name string
}

// NewAccount constructs an account with validated type.
func NewAccount(id AccountID, typ AccountType, name string) (Account, error) {
	if id == "" {
		return Account{}, shareddomain.NewDomainError(ErrCodeInvalidAccount, "account id is required")
	}
	if !typ.isValid() {
		return Account{}, shareddomain.NewDomainError(ErrCodeInvalidAccount, "invalid account type")
	}
	if name == "" {
		return Account{}, shareddomain.NewDomainError(ErrCodeInvalidAccount, "account name is required")
	}
	return Account{id: id, typ: typ, name: name}, nil
}

func (t AccountType) isValid() bool {
	switch t {
	case AccountTypeAsset, AccountTypeLiability, AccountTypeEquity, AccountTypeRevenue, AccountTypeExpense:
		return true
	default:
		return false
	}
}

func (a Account) ID() AccountID       { return a.id }
func (a Account) Type() AccountType   { return a.typ }
func (a Account) Name() string        { return a.name }

// IsDebitNormal reports whether debits increase this account's balance.
func (a Account) IsDebitNormal() bool {
	switch a.typ {
	case AccountTypeAsset, AccountTypeExpense:
		return true
	default:
		return false
	}
}
