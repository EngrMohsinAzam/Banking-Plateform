package ports

import (
	"context"

	shareddomain "github.com/mohsinazam/banking/internal/shared/domain"
	"github.com/mohsinazam/banking/internal/ledger/domain"
)

// Repository persists ledger accounts and append-only journal transactions.
type Repository interface {
	CreateAccount(ctx context.Context, account domain.Account) error
	GetAccount(ctx context.Context, id domain.AccountID) (domain.Account, error)

	// PostTransaction atomically persists a validated journal transaction and its entries.
	// Uses SELECT ... FOR UPDATE on touched accounts (sorted) and rejects postings that
	// would drive any account balance negative. Duplicate transaction IDs return ErrCodeConflict.
	PostTransaction(ctx context.Context, tx domain.Transaction) error

	ListEntriesByAccount(ctx context.Context, accountID domain.AccountID) ([]domain.Entry, error)
	GetBalance(ctx context.Context, account domain.Account) (shareddomain.Money, error)

	// VerifyGlobalLedgerBalanced checks total debits == total credits across all entries.
	VerifyGlobalLedgerBalanced(ctx context.Context) error
}
