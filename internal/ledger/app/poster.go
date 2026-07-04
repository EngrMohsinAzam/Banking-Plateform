// Package app contains ledger application services (posting orchestration).
//
// Concurrency strategy (Step 4):
//   - Pessimistic row locks (SELECT ... FOR UPDATE) on all accounts in a transaction
//   - Accounts locked in sorted ID order to prevent deadlocks
//   - Balance checked under lock before any journal lines are inserted
//   - Optimistic retry is deferred to the transfer/idempotency layer (Step 5–6)
package app

import (
	"context"
	"fmt"

	shareddomain "github.com/mohsinazam/banking/internal/shared/domain"
	"github.com/mohsinazam/banking/internal/ledger/domain"
	"github.com/mohsinazam/banking/internal/ledger/ports"
)

// Poster is the application service for journal posting.
type Poster struct {
	repo ports.Repository
}

// NewPoster constructs a ledger posting service.
func NewPoster(repo ports.Repository) *Poster {
	return &Poster{repo: repo}
}

// Post persists a validated transaction through the repository concurrency guards.
func (p *Poster) Post(ctx context.Context, tx domain.Transaction) error {
	if !tx.IsBalanced() {
		return shareddomain.NewDomainError(domain.ErrCodeUnbalancedTransaction, "transaction is not balanced")
	}
	return p.repo.PostTransaction(ctx, tx)
}

// CreateAccount persists a new ledger account.
func (p *Poster) CreateAccount(ctx context.Context, account domain.Account) error {
	return p.repo.CreateAccount(ctx, account)
}

// GetBalanceForAccount returns the derived balance string for an account id.
func (p *Poster) GetBalanceForAccount(ctx context.Context, accountID domain.AccountID) (string, error) {
	account, err := p.repo.GetAccount(ctx, accountID)
	if err != nil {
		return "", err
	}
	return p.GetBalance(ctx, account)
}

// ListEntries returns journal lines for an account.
func (p *Poster) ListEntries(ctx context.Context, accountID domain.AccountID) ([]domain.Entry, error) {
	return p.repo.ListEntriesByAccount(ctx, accountID)
}

// GetBalance returns the derived balance for an account.
func (p *Poster) GetBalance(ctx context.Context, account domain.Account) (money string, err error) {
	bal, err := p.repo.GetBalance(ctx, account)
	if err != nil {
		return "", err
	}
	return bal.String(), nil
}

// VerifyBooks checks the global debits == credits invariant in storage.
func (p *Poster) VerifyBooks(ctx context.Context) error {
	if err := p.repo.VerifyGlobalLedgerBalanced(ctx); err != nil {
		return fmt.Errorf("books do not balance: %w", err)
	}
	return nil
}
