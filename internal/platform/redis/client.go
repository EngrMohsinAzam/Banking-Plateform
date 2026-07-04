package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// NewClient creates a Redis client for the given address.
func NewClient(addr string) *goredis.Client {
	return goredis.NewClient(&goredis.Options{
		Addr:         addr,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})
}

// Ping verifies Redis connectivity.
func Ping(ctx context.Context, client *goredis.Client) error {
	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping: %w", err)
	}
	return nil
}
