package audit

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Entry is an immutable audit record.
type Entry struct {
	RequestID    string
	Action       string
	Actor        string
	ResourceType string
	ResourceID   string
	Metadata     map[string]any
}

// Logger persists audit events.
type Logger struct {
	pool *pgxpool.Pool
}

// NewLogger constructs a Postgres audit logger.
func NewLogger(pool *pgxpool.Pool) *Logger {
	return &Logger{pool: pool}
}

// Log writes an audit row.
func (l *Logger) Log(ctx context.Context, entry Entry) error {
	meta, err := json.Marshal(entry.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}
	_, err = l.pool.Exec(ctx, `
		INSERT INTO audit_log (request_id, action, actor, resource_type, resource_id, metadata)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, entry.RequestID, entry.Action, null(entry.Actor), null(entry.ResourceType), null(entry.ResourceID), meta)
	return err
}

func null(v string) any {
	if v == "" {
		return nil
	}
	return v
}
