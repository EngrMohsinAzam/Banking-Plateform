package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// VelocityStore counts transfers per account in a sliding window using Redis INCR + EXPIRE.
type VelocityStore struct {
	client *goredis.Client
}

// NewVelocityStore constructs a Redis velocity adapter.
func NewVelocityStore(client *goredis.Client) *VelocityStore {
	return &VelocityStore{client: client}
}

// Increment returns the new count for the account in the given window.
func (s *VelocityStore) Increment(ctx context.Context, accountID string, window time.Duration) (int64, error) {
	key := fmt.Sprintf("velocity:%s", accountID)
	pipe := s.client.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, window)
	if _, err := pipe.Exec(ctx); err != nil {
		return 0, fmt.Errorf("redis velocity: %w", err)
	}
	return incr.Val(), nil
}
