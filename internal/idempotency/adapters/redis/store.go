package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"

	shareddomain "github.com/mohsinazam/banking/internal/shared/domain"
	"github.com/mohsinazam/banking/internal/idempotency/domain"
)

// Store implements idempotency persistence in Redis.
type Store struct {
	client *goredis.Client
}

// NewStore constructs a Redis-backed idempotency store.
func NewStore(client *goredis.Client) *Store {
	return &Store{client: client}
}

// Get returns an existing record or ErrNotFound.
func (s *Store) Get(ctx context.Context, scope domain.Scope, key domain.Key) (domain.Record, error) {
	data, err := s.client.Get(ctx, redisKey(scope, key)).Bytes()
	if errors.Is(err, goredis.Nil) {
		return domain.Record{}, shareddomain.NewDomainError(shareddomain.ErrCodeNotFound, "idempotency record not found")
	}
	if err != nil {
		return domain.Record{}, fmt.Errorf("redis get: %w", err)
	}

	record, err := domain.UnmarshalRecord(data)
	if err != nil {
		return domain.Record{}, fmt.Errorf("decode record: %w", err)
	}
	return record, nil
}

// TryAcquire creates a PROCESSING record using SET NX.
func (s *Store) TryAcquire(
	ctx context.Context,
	scope domain.Scope,
	key domain.Key,
	fingerprint domain.Fingerprint,
	ttl time.Duration,
) (bool, error) {
	record := domain.Record{
		Scope:       scope,
		Key:         key,
		Status:      domain.StatusProcessing,
		Fingerprint: fingerprint,
		CreatedAt:   time.Now().UTC(),
	}
	data, err := record.Marshal()
	if err != nil {
		return false, err
	}

	ok, err := s.client.SetNX(ctx, redisKey(scope, key), data, ttl).Result()
	if err != nil {
		return false, fmt.Errorf("redis setnx: %w", err)
	}
	return ok, nil
}

// Complete marks a record as COMPLETED and stores the response payload.
func (s *Store) Complete(
	ctx context.Context,
	scope domain.Scope,
	key domain.Key,
	result domain.Result,
	ttl time.Duration,
) error {
	existing, err := s.Get(ctx, scope, key)
	if err != nil {
		return err
	}

	existing.Status = domain.StatusCompleted
	existing.ResourceID = result.ResourceID
	existing.Payload = append([]byte(nil), result.Payload...)
	existing.CompletedAt = time.Now().UTC()

	data, err := existing.Marshal()
	if err != nil {
		return err
	}

	if err := s.client.Set(ctx, redisKey(scope, key), data, ttl).Err(); err != nil {
		return fmt.Errorf("redis set complete: %w", err)
	}
	return nil
}

// Fail marks a record as FAILED so clients must use a new key for money movement.
func (s *Store) Fail(ctx context.Context, scope domain.Scope, key domain.Key, ttl time.Duration) error {
	existing, err := s.Get(ctx, scope, key)
	if err != nil {
		return err
	}

	existing.Status = domain.StatusFailed
	data, err := existing.Marshal()
	if err != nil {
		return err
	}

	if err := s.client.Set(ctx, redisKey(scope, key), data, ttl).Err(); err != nil {
		return fmt.Errorf("redis set failed: %w", err)
	}
	return nil
}

// Delete removes a record (used when validation fails before money moves).
func (s *Store) Delete(ctx context.Context, scope domain.Scope, key domain.Key) error {
	if err := s.client.Del(ctx, redisKey(scope, key)).Err(); err != nil {
		return fmt.Errorf("redis del: %w", err)
	}
	return nil
}

func redisKey(scope domain.Scope, key domain.Key) string {
	return fmt.Sprintf("idempotency:%s:%s", scope, key)
}
