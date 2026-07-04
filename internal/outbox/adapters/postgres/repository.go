package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	shareddomain "github.com/mohsinazam/banking/internal/shared/domain"
	"github.com/mohsinazam/banking/internal/outbox/domain"
)

// Repository is the PostgreSQL outbox adapter.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository constructs an outbox repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// InsertInTx stages an event in the same database transaction as business writes.
func (r *Repository) InsertInTx(ctx context.Context, tx any, event domain.Event) error {
	dbTx, ok := tx.(pgx.Tx)
	if !ok {
		return fmt.Errorf("expected pgx.Tx, got %T", tx)
	}

	_, err := dbTx.Exec(ctx, `
		INSERT INTO outbox_events (id, aggregate_type, aggregate_id, event_type, payload, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`,
		event.ID(),
		event.AggregateType(),
		event.AggregateID(),
		event.EventType(),
		event.Payload(),
		string(domain.StatusPending),
		event.CreatedAt(),
	)
	if err != nil {
		if isUniqueViolation(err) {
			return shareddomain.NewDomainError(shareddomain.ErrCodeConflict, "outbox event already exists")
		}
		return fmt.Errorf("insert outbox event: %w", err)
	}
	return nil
}

// ProcessPendingBatch claims, publishes, and marks events in one DB transaction.
func (r *Repository) ProcessPendingBatch(
	ctx context.Context,
	limit int,
	publish func(context.Context, domain.Event) error,
) (int, error) {
	if limit <= 0 {
		limit = 10
	}

	dbTx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = dbTx.Rollback(ctx)
	}()

	events, err := claimPendingInTx(ctx, dbTx, limit)
	if err != nil {
		return 0, err
	}

	processed := 0
	for _, event := range events {
		if err := publish(ctx, event); err != nil {
			if markErr := markFailedInTx(ctx, dbTx, event.ID()); markErr != nil {
				return processed, markErr
			}
			continue
		}
		if err := markPublishedInTx(ctx, dbTx, event.ID()); err != nil {
			return processed, err
		}
		processed++
	}

	if err := dbTx.Commit(ctx); err != nil {
		return processed, fmt.Errorf("commit outbox batch: %w", err)
	}
	return processed, nil
}

// FetchPending returns pending events without claiming (diagnostics/tests).
func (r *Repository) FetchPending(ctx context.Context, limit int) ([]domain.Event, error) {
	if limit <= 0 {
		limit = 10
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, aggregate_type, aggregate_id, event_type, payload, status, created_at, published_at
		FROM outbox_events
		WHERE status = 'PENDING'
		ORDER BY created_at ASC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("fetch pending outbox: %w", err)
	}
	defer rows.Close()

	return scanEvents(rows)
}

// MarkPublished marks an event published outside a batch (legacy port support).
func (r *Repository) MarkPublished(ctx context.Context, eventID string) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE outbox_events
		SET status = 'PUBLISHED', published_at = NOW()
		WHERE id = $1 AND status = 'PENDING'
	`, eventID)
	if err != nil {
		return fmt.Errorf("mark published: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return shareddomain.NewDomainError(shareddomain.ErrCodeNotFound, "outbox event not found")
	}
	return nil
}

// MarkFailed marks an event failed outside a batch (legacy port support).
func (r *Repository) MarkFailed(ctx context.Context, eventID string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE outbox_events
		SET status = 'FAILED'
		WHERE id = $1 AND status = 'PENDING'
	`, eventID)
	if err != nil {
		return fmt.Errorf("mark failed: %w", err)
	}
	return nil
}

func claimPendingInTx(ctx context.Context, dbTx pgx.Tx, limit int) ([]domain.Event, error) {
	rows, err := dbTx.Query(ctx, `
		SELECT id, aggregate_type, aggregate_id, event_type, payload, status, created_at, published_at
		FROM outbox_events
		WHERE status = 'PENDING'
		ORDER BY created_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("claim pending outbox: %w", err)
	}
	defer rows.Close()
	return scanEvents(rows)
}

func markPublishedInTx(ctx context.Context, dbTx pgx.Tx, eventID string) error {
	_, err := dbTx.Exec(ctx, `
		UPDATE outbox_events
		SET status = 'PUBLISHED', published_at = NOW()
		WHERE id = $1
	`, eventID)
	return err
}

func markFailedInTx(ctx context.Context, dbTx pgx.Tx, eventID string) error {
	_, err := dbTx.Exec(ctx, `
		UPDATE outbox_events
		SET status = 'FAILED'
		WHERE id = $1
	`, eventID)
	return err
}

func scanEvents(rows pgx.Rows) ([]domain.Event, error) {
	var events []domain.Event
	for rows.Next() {
		var (
			id, aggregateType, aggregateID, eventType, status string
			payload                                           []byte
			createdAt                                         time.Time
			publishedAt                                       *time.Time
		)
		if err := rows.Scan(&id, &aggregateType, &aggregateID, &eventType, &payload, &status, &createdAt, &publishedAt); err != nil {
			return nil, fmt.Errorf("scan outbox: %w", err)
		}
		var published time.Time
		if publishedAt != nil {
			published = *publishedAt
		}
		events = append(events, domain.RehydrateEvent(
			id, aggregateType, aggregateID, eventType, payload,
			domain.Status(status), createdAt, published,
		))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate outbox: %w", err)
	}
	return events, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
