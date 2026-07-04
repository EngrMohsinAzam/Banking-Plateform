package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"

	outboxdomain "github.com/mohsinazam/banking/internal/outbox/domain"
)

// Publisher delivers outbox events to a Kafka topic.
type Publisher struct {
	writer *kafka.Writer
	logger *slog.Logger
}

// Config configures the Kafka outbox publisher.
type Config struct {
	Brokers []string
	Topic   string
}

// NewPublisher constructs a Kafka publisher. Caller must call Close on shutdown.
func NewPublisher(cfg Config, logger *slog.Logger) *Publisher {
	if logger == nil {
		logger = slog.Default()
	}
	return &Publisher{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(cfg.Brokers...),
			Topic:        cfg.Topic,
			Balancer:     &kafka.LeastBytes{},
			RequiredAcks: kafka.RequireOne,
			Async:        false,
		},
		logger: logger,
	}
}

// Publish writes the event payload to Kafka.
func (p *Publisher) Publish(ctx context.Context, event outboxdomain.Event) error {
	body, err := json.Marshal(map[string]any{
		"event_id":   event.ID(),
		"event_type": event.EventType(),
		"payload":    json.RawMessage(event.Payload()),
		"occurred_at": event.CreatedAt().UTC().Format(time.RFC3339Nano),
	})
	if err != nil {
		return fmt.Errorf("marshal kafka message: %w", err)
	}

	if err := p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(event.EventType()),
		Value: body,
	}); err != nil {
		return fmt.Errorf("kafka publish: %w", err)
	}

	p.logger.Info("event published to kafka",
		"event_id", event.ID(),
		"event_type", event.EventType(),
		"topic", p.writer.Topic,
	)
	return nil
}

// Close releases the Kafka writer.
func (p *Publisher) Close() error {
	return p.writer.Close()
}
