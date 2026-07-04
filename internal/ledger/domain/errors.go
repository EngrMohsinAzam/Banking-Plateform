package domain

import shareddomain "github.com/mohsinazam/banking/internal/shared/domain"

// Ledger-specific error codes. Reuses the shared DomainError type.
const (
	ErrCodeUnbalancedTransaction shareddomain.ErrorCode = "UNBALANCED_TRANSACTION"
	ErrCodeInvalidLedgerEntry    shareddomain.ErrorCode = "INVALID_LEDGER_ENTRY"
	ErrCodeInvalidAccount        shareddomain.ErrorCode = "INVALID_ACCOUNT"
)
