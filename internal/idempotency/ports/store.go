package ports

import (
	"context"
	"time"

	"github.com/mohsinazam/banking/internal/idempotency/domain"
)

// Store persists idempotency lifecycle state.
type Store interface {
	Get(ctx context.Context, scope domain.Scope, key domain.Key) (domain.Record, error)
	TryAcquire(ctx context.Context, scope domain.Scope, key domain.Key, fingerprint domain.Fingerprint, ttl time.Duration) (bool, error)
	Complete(ctx context.Context, scope domain.Scope, key domain.Key, result domain.Result, ttl time.Duration) error
	Fail(ctx context.Context, scope domain.Scope, key domain.Key, ttl time.Duration) error
	Delete(ctx context.Context, scope domain.Scope, key domain.Key) error
}
