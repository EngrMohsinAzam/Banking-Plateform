package domain

import (
	"encoding/json"
	"fmt"
	"time"

	shareddomain "github.com/mohsinazam/banking/internal/shared/domain"
)

const (
	AggregateTransfer = "transfer"

	EventTransferPosted = "transfer.posted"
)

// Status is the publication lifecycle of an outbox row.
type Status string

const (
	StatusPending   Status = "PENDING"
	StatusPublished Status = "PUBLISHED"
	StatusFailed    Status = "FAILED"
)

// Event is a domain event staged for reliable publication.
type Event struct {
	id            string
	aggregateType string
	aggregateID   string
	eventType     string
	payload       []byte
	status        Status
	createdAt     time.Time
	publishedAt   time.Time
}

// NewEvent constructs a pending outbox event.
func NewEvent(id, aggregateType, aggregateID, eventType string, payload []byte) (Event, error) {
	if id == "" {
		return Event{}, shareddomain.NewDomainError(shareddomain.ErrCodeValidation, "event id is required")
	}
	if aggregateType == "" || aggregateID == "" || eventType == "" {
		return Event{}, shareddomain.NewDomainError(shareddomain.ErrCodeValidation, "aggregate metadata is required")
	}
	if len(payload) == 0 {
		return Event{}, shareddomain.NewDomainError(shareddomain.ErrCodeValidation, "event payload is required")
	}
	if !json.Valid(payload) {
		return Event{}, shareddomain.NewDomainError(shareddomain.ErrCodeValidation, "event payload must be valid JSON")
	}
	return Event{
		id:            id,
		aggregateType: aggregateType,
		aggregateID:   aggregateID,
		eventType:     eventType,
		payload:       append([]byte(nil), payload...),
		status:        StatusPending,
		createdAt:     time.Now().UTC(),
	}, nil
}

// RehydrateEvent rebuilds an event loaded from storage.
func RehydrateEvent(
	id, aggregateType, aggregateID, eventType string,
	payload []byte,
	status Status,
	createdAt, publishedAt time.Time,
) Event {
	return Event{
		id:            id,
		aggregateType: aggregateType,
		aggregateID:   aggregateID,
		eventType:     eventType,
		payload:       append([]byte(nil), payload...),
		status:        status,
		createdAt:     createdAt,
		publishedAt:   publishedAt,
	}
}

func (e Event) ID() string              { return e.id }
func (e Event) AggregateType() string   { return e.aggregateType }
func (e Event) AggregateID() string     { return e.aggregateID }
func (e Event) EventType() string       { return e.eventType }
func (e Event) Payload() []byte         { return append([]byte(nil), e.payload...) }
func (e Event) Status() Status          { return e.status }
func (e Event) CreatedAt() time.Time    { return e.createdAt }
func (e Event) PublishedAt() time.Time  { return e.publishedAt }

// TransferPostedID builds a deterministic outbox id for a transfer journal id.
func TransferPostedID(transactionID string) string {
	return fmt.Sprintf("evt_transfer_posted_%s", transactionID)
}
