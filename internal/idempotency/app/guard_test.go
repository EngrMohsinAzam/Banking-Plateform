package app_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"

	idempotencyapp "github.com/mohsinazam/banking/internal/idempotency/app"
	idempotencydomain "github.com/mohsinazam/banking/internal/idempotency/domain"
	idempotencyredis "github.com/mohsinazam/banking/internal/idempotency/adapters/redis"
	shareddomain "github.com/mohsinazam/banking/internal/shared/domain"
)

func setupGuard(t *testing.T) (*idempotencyapp.Guard, *miniredis.Miniredis) {
	t.Helper()

	mr := miniredis.RunT(t)
	client := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	store := idempotencyredis.NewStore(client)
	guard := idempotencyapp.NewGuard(store, idempotencyapp.Config{
		ProcessingTTL: time.Minute,
		CompletedTTL:  time.Hour,
	})
	return guard, mr
}

func TestGuardRunsOnceAndReplays(t *testing.T) {
	guard, _ := setupGuard(t)
	ctx := context.Background()

	key, err := idempotencydomain.ParseKey("client-key-001")
	require.NoError(t, err)
	scope := idempotencydomain.Scope("transfer")
	fp := idempotencydomain.FingerprintFromPayload([]byte(`{"amount":"10.00"}`))

	var calls atomic.Int32
	fn := func(ctx context.Context) (idempotencydomain.Result, error) {
		calls.Add(1)
		return idempotencydomain.Result{
			ResourceID: "tx-123",
			Payload:    []byte(`{"status":"posted"}`),
		}, nil
	}

	result, replay, err := guard.Run(ctx, scope, key, fp, fn)
	require.NoError(t, err)
	require.False(t, replay)
	require.Equal(t, "tx-123", result.ResourceID)

	result2, replay2, err := guard.Run(ctx, scope, key, fp, fn)
	require.NoError(t, err)
	require.True(t, replay2)
	require.Equal(t, result.ResourceID, result2.ResourceID)
	require.Equal(t, int32(1), calls.Load())
}

func TestGuardRejectsDifferentPayloadSameKey(t *testing.T) {
	guard, _ := setupGuard(t)
	ctx := context.Background()

	key, _ := idempotencydomain.ParseKey("client-key-002")
	scope := idempotencydomain.Scope("transfer")

	_, _, err := guard.Run(ctx, scope, key,
		idempotencydomain.FingerprintFromPayload([]byte(`{"amount":"10.00"}`)),
		func(ctx context.Context) (idempotencydomain.Result, error) {
			return idempotencydomain.Result{ResourceID: "tx-1"}, nil
		},
	)
	require.NoError(t, err)

	_, _, err = guard.Run(ctx, scope, key,
		idempotencydomain.FingerprintFromPayload([]byte(`{"amount":"20.00"}`)),
		func(ctx context.Context) (idempotencydomain.Result, error) {
			return idempotencydomain.Result{ResourceID: "tx-2"}, nil
		},
	)
	require.Error(t, err)
	require.True(t, shareddomain.IsDomainCode(err, shareddomain.ErrCodeConflict))
}

func TestGuardConcurrentRequestsSingleExecution(t *testing.T) {
	guard, _ := setupGuard(t)
	ctx := context.Background()

	key, _ := idempotencydomain.ParseKey("client-key-003")
	scope := idempotencydomain.Scope("transfer")
	fp := idempotencydomain.FingerprintFromPayload([]byte(`{"amount":"50.00"}`))

	var calls atomic.Int32
	var wg sync.WaitGroup
	results := make(chan idempotencydomain.Result, 20)

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, _, err := guard.Run(ctx, scope, key, fp, func(ctx context.Context) (idempotencydomain.Result, error) {
				calls.Add(1)
				time.Sleep(50 * time.Millisecond)
				return idempotencydomain.Result{ResourceID: "tx-concurrent"}, nil
			})
			if err == nil {
				results <- result
			}
		}()
	}
	wg.Wait()
	close(results)

	require.Equal(t, int32(1), calls.Load())
	for result := range results {
		require.Equal(t, "tx-concurrent", result.ResourceID)
	}
}

func TestGuardValidationFailureAllowsRetry(t *testing.T) {
	guard, _ := setupGuard(t)
	ctx := context.Background()

	key, _ := idempotencydomain.ParseKey("client-key-004")
	scope := idempotencydomain.Scope("transfer")
	fp := idempotencydomain.FingerprintFromPayload([]byte(`{"iban":"bad"}`))

	_, _, err := guard.Run(ctx, scope, key, fp, func(ctx context.Context) (idempotencydomain.Result, error) {
		return idempotencydomain.Result{}, shareddomain.NewDomainError(shareddomain.ErrCodeValidation, "invalid iban")
	})
	require.Error(t, err)

	var calls atomic.Int32
	_, replay, err := guard.Run(ctx, scope, key, fp, func(ctx context.Context) (idempotencydomain.Result, error) {
		calls.Add(1)
		return idempotencydomain.Result{ResourceID: "tx-fixed"}, nil
	})
	require.NoError(t, err)
	require.False(t, replay)
	require.Equal(t, int32(1), calls.Load())
}

func TestGuardProcessingReturnsInProgress(t *testing.T) {
	guard, mr := setupGuard(t)
	ctx := context.Background()

	key, _ := idempotencydomain.ParseKey("client-key-005")
	scope := idempotencydomain.Scope("transfer")
	fp := idempotencydomain.FingerprintFromPayload([]byte(`{"amount":"1.00"}`))

	processing := idempotencydomain.Record{
		Scope:       scope,
		Key:         key,
		Status:      idempotencydomain.StatusProcessing,
		Fingerprint: fp,
		CreatedAt:   time.Now().UTC(),
	}
	data, err := processing.Marshal()
	require.NoError(t, err)
	require.NoError(t, mr.Set("idempotency:transfer:client-key-005", string(data)))

	_, _, err = guard.Run(ctx, scope, key, fp, func(ctx context.Context) (idempotencydomain.Result, error) {
		return idempotencydomain.Result{ResourceID: "tx-never"}, nil
	})
	require.Error(t, err)
	require.True(t, shareddomain.IsDomainCode(err, shareddomain.ErrCodeRequestInProgress))
}
