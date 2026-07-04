package ports

import (
	"context"

	"github.com/mohsinazam/banking/internal/outbox/domain"
)

// Repository persists outbox rows and supports transactional inserts.
type Repository interface {
	InsertInTx(ctx context.Context, tx any, event domain.Event) error
	FetchPending(ctx context.Context, limit int) ([]domain.Event, error)
	MarkPublished(ctx context.Context, eventID string) error
	MarkFailed(ctx context.Context, eventID string) error
}

// Publisher delivers staged events to downstream consumers (mock bus, Kafka, etc.).
type Publisher interface {
	Publish(ctx context.Context, event domain.Event) error
}
