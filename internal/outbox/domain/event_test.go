package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/mohsinazam/banking/internal/outbox/domain"
)

func TestNewEvent(t *testing.T) {
	t.Parallel()

	payload, err := json.Marshal(map[string]string{"transaction_id": "tx-1"})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	event, err := domain.NewEvent("evt-1", domain.AggregateTransfer, "tx-1", domain.EventTransferPosted, payload)
	if err != nil {
		t.Fatalf("NewEvent() error = %v", err)
	}
	if event.Status() != domain.StatusPending {
		t.Fatalf("status = %s", event.Status())
	}
}

func TestNewEventRejectsInvalidJSON(t *testing.T) {
	t.Parallel()

	_, err := domain.NewEvent("evt-2", domain.AggregateTransfer, "tx-1", domain.EventTransferPosted, []byte("not-json"))
	if err == nil {
		t.Fatal("expected error")
	}
}
