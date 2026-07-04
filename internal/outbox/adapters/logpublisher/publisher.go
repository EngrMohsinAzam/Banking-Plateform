package logpublisher

import (
	"context"
	"log/slog"

	"github.com/mohsinazam/banking/internal/outbox/domain"
)

// Publisher logs events to stdout — realistic mock for local dev and tests.
type Publisher struct {
	logger *slog.Logger
}

// NewPublisher constructs a logging event publisher.
func NewPublisher(logger *slog.Logger) *Publisher {
	return &Publisher{logger: logger}
}

// Publish emits the event to structured logs (stand-in for Kafka/NATS).
func (p *Publisher) Publish(ctx context.Context, event domain.Event) error {
	_ = ctx
	p.logger.Info("event published",
		"event_id", event.ID(),
		"event_type", event.EventType(),
		"aggregate_type", event.AggregateType(),
		"aggregate_id", event.AggregateID(),
		"payload", string(event.Payload()),
	)
	return nil
}
