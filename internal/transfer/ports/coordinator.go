package ports

import (
	"context"

	uowpostgres "github.com/mohsinazam/banking/internal/platform/uow/postgres"
)

// TransferCoordinator atomically commits transfer-side effects.
type TransferCoordinator interface {
	PostTransfer(ctx context.Context, commit uowpostgres.TransferCommit) error
}
