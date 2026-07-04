package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	shareddomain "github.com/mohsinazam/banking/internal/shared/domain"
	"github.com/mohsinazam/banking/internal/transfer/ports"
)

const (
	defaultMaxSingleHalalas   int64 = 500_000_00 // 500,000 SAR
	defaultMaxHourlyTransfers       = 20
)

// VelocityStore tracks transfer frequency per account.
type VelocityStore interface {
	Increment(ctx context.Context, accountID string, window time.Duration) (int64, error)
}

// Checker applies fraud rules and Redis-backed velocity limits.
type Checker struct {
	velocity           VelocityStore
	maxSingleHalalas   int64
	maxHourlyTransfers int64
	velocityWindow     time.Duration
	logger             *slog.Logger
}

// Config tunes fraud thresholds.
type Config struct {
	MaxSingleHalalas   int64
	MaxHourlyTransfers int64
	VelocityWindow     time.Duration
}

// NewChecker constructs a fraud checker.
func NewChecker(store VelocityStore, cfg Config) *Checker {
	if cfg.MaxSingleHalalas == 0 {
		cfg.MaxSingleHalalas = defaultMaxSingleHalalas
	}
	if cfg.MaxHourlyTransfers == 0 {
		cfg.MaxHourlyTransfers = defaultMaxHourlyTransfers
	}
	if cfg.VelocityWindow == 0 {
		cfg.VelocityWindow = time.Hour
	}
	return &Checker{
		velocity:           store,
		maxSingleHalalas:   cfg.MaxSingleHalalas,
		maxHourlyTransfers: cfg.MaxHourlyTransfers,
		velocityWindow:     cfg.VelocityWindow,
		logger:             slog.Default(),
	}
}

// Check returns an error if the transfer fails fraud rules.
func (c *Checker) Check(ctx context.Context, input ports.FraudCheckInput) error {
	if input.Amount.Halalas() > c.maxSingleHalalas {
		c.logger.Warn("fraud blocked transfer",
			"reason", "amount_limit",
			"from_account", input.FromAccountID,
			"amount_halalas", input.Amount.Halalas(),
		)
		return shareddomain.NewDomainError(
			shareddomain.ErrCodeForbidden,
			fmt.Sprintf("transfer exceeds single-transaction limit of %d halalas", c.maxSingleHalalas),
		)
	}

	count, err := c.velocity.Increment(ctx, input.FromAccountID, c.velocityWindow)
	if err != nil {
		c.logger.Error("fraud velocity store failed", "from_account", input.FromAccountID, "error", err)
		return fmt.Errorf("velocity check: %w", err)
	}
	if count > c.maxHourlyTransfers {
		c.logger.Warn("fraud blocked transfer",
			"reason", "velocity_limit",
			"from_account", input.FromAccountID,
			"count", count,
		)
		return shareddomain.NewDomainError(
			shareddomain.ErrCodeForbidden,
			"transfer velocity limit exceeded",
		)
	}
	return nil
}
