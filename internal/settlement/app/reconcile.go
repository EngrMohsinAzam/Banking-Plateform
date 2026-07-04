package app

import (
	"context"
	"fmt"

	"log/slog"

	shareddomain "github.com/mohsinazam/banking/internal/shared/domain"
)

// LedgerVerifier checks global double-entry invariants.
type LedgerVerifier interface {
	VerifyGlobalLedgerBalanced(ctx context.Context) error
}

// Reconciler runs periodic book-balance checks.
type Reconciler struct {
	ledger LedgerVerifier
	logger *slog.Logger
}

// NewReconciler constructs a reconciliation job.
func NewReconciler(ledger LedgerVerifier, logger *slog.Logger) *Reconciler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Reconciler{ledger: ledger, logger: logger}
}

// Run verifies debits equal credits across the ledger.
func (r *Reconciler) Run(ctx context.Context) error {
	if err := r.ledger.VerifyGlobalLedgerBalanced(ctx); err != nil {
		r.logger.Error("reconciliation failed", "error", err)
		return fmt.Errorf("reconciliation failed: %w", err)
	}
	r.logger.Debug("reconciliation ok")
	return nil
}

// IsBalanced is a helper for tests.
func IsBalanced(err error) bool {
	if err == nil {
		return true
	}
	return !shareddomain.IsDomainCode(err, shareddomain.ErrCodeValidation)
}
