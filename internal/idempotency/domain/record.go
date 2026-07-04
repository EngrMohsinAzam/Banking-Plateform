package domain

import (
	"encoding/json"
	"time"
)

// Status is the lifecycle state of an idempotency record.
type Status string

const (
	StatusProcessing Status = "PROCESSING"
	StatusCompleted  Status = "COMPLETED"
	StatusFailed     Status = "FAILED"
)

// Record is the persisted idempotency state for a single key + scope.
type Record struct {
	Scope       Scope       `json:"scope"`
	Key         Key         `json:"key"`
	Status      Status      `json:"status"`
	Fingerprint Fingerprint `json:"fingerprint"`
	ResourceID  string      `json:"resource_id,omitempty"`
	Payload     []byte      `json:"payload,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
	CompletedAt time.Time   `json:"completed_at,omitempty"`
}

// Result is returned to callers after a successful operation.
type Result struct {
	ResourceID string
	Payload    []byte
}

func (r Record) ToResult() Result {
	return Result{
		ResourceID: r.ResourceID,
		Payload:    append([]byte(nil), r.Payload...),
	}
}

func (r Record) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalRecord(data []byte) (Record, error) {
	var record Record
	if err := json.Unmarshal(data, &record); err != nil {
		return Record{}, err
	}
	return record, nil
}
