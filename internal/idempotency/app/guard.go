package app

import (
	"context"
	"errors"
	"time"

	shareddomain "github.com/mohsinazam/banking/internal/shared/domain"
	"github.com/mohsinazam/banking/internal/idempotency/domain"
	"github.com/mohsinazam/banking/internal/idempotency/ports"
)

// Guard orchestrates idempotent execution of money-moving operations.
type Guard struct {
	store          ports.Store
	processingTTL  time.Duration
	completedTTL   time.Duration
}

// Config tunes idempotency retention windows.
type Config struct {
	ProcessingTTL time.Duration
	CompletedTTL  time.Duration
}

// DefaultConfig returns production-like TTL defaults.
func DefaultConfig() Config {
	return Config{
		ProcessingTTL: 15 * time.Minute,
		CompletedTTL:  24 * time.Hour,
	}
}

// NewGuard constructs an idempotency guard.
func NewGuard(store ports.Store, cfg Config) *Guard {
	if cfg.ProcessingTTL == 0 {
		cfg.ProcessingTTL = DefaultConfig().ProcessingTTL
	}
	if cfg.CompletedTTL == 0 {
		cfg.CompletedTTL = DefaultConfig().CompletedTTL
	}
	return &Guard{
		store:         store,
		processingTTL: cfg.ProcessingTTL,
		completedTTL:  cfg.CompletedTTL,
	}
}

// Run executes fn at most once per (scope, key). Returns replay=true when a cached result is returned.
func (g *Guard) Run(
	ctx context.Context,
	scope domain.Scope,
	key domain.Key,
	fingerprint domain.Fingerprint,
	fn func(ctx context.Context) (domain.Result, error),
) (result domain.Result, replay bool, err error) {
	if existing, getErr := g.store.Get(ctx, scope, key); getErr == nil {
		if err := validateFingerprint(existing, fingerprint); err != nil {
			return domain.Result{}, false, err
		}
		switch existing.Status {
		case domain.StatusCompleted:
			return existing.ToResult(), true, nil
		case domain.StatusProcessing:
			return domain.Result{}, false, shareddomain.NewDomainError(
				shareddomain.ErrCodeRequestInProgress,
				"request with this idempotency key is still processing",
			)
		case domain.StatusFailed:
			return domain.Result{}, false, shareddomain.NewDomainError(
				shareddomain.ErrCodeConflict,
				"idempotency key already used for a failed operation; use a new key",
			)
		}
	} else if !shareddomain.IsDomainCode(getErr, shareddomain.ErrCodeNotFound) {
		return domain.Result{}, false, getErr
	}

	acquired, err := g.store.TryAcquire(ctx, scope, key, fingerprint, g.processingTTL)
	if err != nil {
		return domain.Result{}, false, err
	}
	if !acquired {
		return g.waitForExisting(ctx, scope, key, fingerprint)
	}

	result, err = fn(ctx)
	if err != nil {
		g.handleFailure(ctx, scope, key, err)
		return domain.Result{}, false, err
	}

	if err := g.store.Complete(ctx, scope, key, result, g.completedTTL); err != nil {
		return domain.Result{}, false, err
	}
	return result, false, nil
}

func (g *Guard) waitForExisting(
	ctx context.Context,
	scope domain.Scope,
	key domain.Key,
	fingerprint domain.Fingerprint,
) (domain.Result, bool, error) {
	existing, err := g.store.Get(ctx, scope, key)
	if err != nil {
		return domain.Result{}, false, err
	}
	if err := validateFingerprint(existing, fingerprint); err != nil {
		return domain.Result{}, false, err
	}
	switch existing.Status {
	case domain.StatusCompleted:
		return existing.ToResult(), true, nil
	case domain.StatusProcessing:
		return domain.Result{}, false, shareddomain.NewDomainError(
			shareddomain.ErrCodeRequestInProgress,
			"request with this idempotency key is still processing",
		)
	default:
		return domain.Result{}, false, shareddomain.NewDomainError(
			shareddomain.ErrCodeConflict,
			"idempotency key already used for a failed operation; use a new key",
		)
	}
}

func (g *Guard) handleFailure(ctx context.Context, scope domain.Scope, key domain.Key, err error) {
	if isSafeToRetry(err) {
		_ = g.store.Delete(ctx, scope, key)
		return
	}
	_ = g.store.Fail(ctx, scope, key, g.completedTTL)
}

func validateFingerprint(existing domain.Record, fingerprint domain.Fingerprint) error {
	if existing.Fingerprint == "" || fingerprint == "" {
		return nil
	}
	if existing.Fingerprint != fingerprint {
		return shareddomain.NewDomainError(
			shareddomain.ErrCodeConflict,
			"idempotency key reused with a different request payload",
		)
	}
	return nil
}

// isSafeToRetry reports whether the client may retry with the same key (validation failed before side effects).
func isSafeToRetry(err error) bool {
	var de *shareddomain.DomainError
	if !errors.As(err, &de) {
		return false
	}
	switch de.Code {
	case shareddomain.ErrCodeValidation,
		shareddomain.ErrCodeInvalidMoney,
		shareddomain.ErrCodeInvalidIBAN,
		shareddomain.ErrCodeForbidden:
		return true
	default:
		return false
	}
}
