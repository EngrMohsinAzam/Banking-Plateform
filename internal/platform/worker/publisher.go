package worker

import (
	"fmt"
	"log/slog"

	"github.com/mohsinazam/banking/internal/outbox/adapters/kafka"
	outboxlog "github.com/mohsinazam/banking/internal/outbox/adapters/logpublisher"
	"github.com/mohsinazam/banking/internal/outbox/ports"
	"github.com/mohsinazam/banking/internal/platform/config"
)

// PublisherHandle wraps an outbox publisher with optional cleanup.
type PublisherHandle struct {
	ports.Publisher
	close func() error
}

// Close releases publisher resources when applicable.
func (h PublisherHandle) Close() error {
	if h.close != nil {
		return h.close()
	}
	return nil
}

// NewEventPublisher selects log or Kafka publisher from config.
func NewEventPublisher(cfg config.Config, logger *slog.Logger) (PublisherHandle, error) {
	switch cfg.EventPublisher {
	case "", "log":
		return PublisherHandle{Publisher: outboxlog.NewPublisher(logger)}, nil
	case "kafka":
		if len(cfg.KafkaBrokers) == 0 || cfg.KafkaTopic == "" {
			return PublisherHandle{}, fmt.Errorf("kafka publisher requires KAFKA_BROKERS and KAFKA_TOPIC")
		}
		pub := kafka.NewPublisher(kafka.Config{Brokers: cfg.KafkaBrokers, Topic: cfg.KafkaTopic}, logger)
		return PublisherHandle{Publisher: pub, close: pub.Close}, nil
	default:
		return PublisherHandle{}, fmt.Errorf("unknown EVENT_PUBLISHER: %s", cfg.EventPublisher)
	}
}
