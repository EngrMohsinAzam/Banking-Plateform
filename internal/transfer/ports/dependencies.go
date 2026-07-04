package ports

import (
	"context"

	idempotencydomain "github.com/mohsinazam/banking/internal/idempotency/domain"
	ledgerdomain "github.com/mohsinazam/banking/internal/ledger/domain"
	outboxdomain "github.com/mohsinazam/banking/internal/outbox/domain"
)

// AccountReader loads ledger account metadata.
type AccountReader interface {
	GetAccount(ctx context.Context, id ledgerdomain.AccountID) (ledgerdomain.Account, error)
}

// TransactionWriter atomically persists journal rows and outbox events.
type TransactionWriter interface {
	PostTransactionWithEvents(ctx context.Context, tx ledgerdomain.Transaction, events []outboxdomain.Event) error
}

// Idempotency runs a money-moving operation at most once per key.
type Idempotency interface {
	Run(
		ctx context.Context,
		scope idempotencydomain.Scope,
		key idempotencydomain.Key,
		fingerprint idempotencydomain.Fingerprint,
		fn func(ctx context.Context) (idempotencydomain.Result, error),
	) (idempotencydomain.Result, bool, error)
}
