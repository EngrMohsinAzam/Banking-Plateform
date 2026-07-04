package app

import (
	"context"
	"math/rand"
	"time"

	settlementports "github.com/mohsinazam/banking/internal/settlement/ports"
)

// SarieConfig configures mock rail behaviour.
type SarieConfig struct {
	FailRate float64
	MinDelay time.Duration
	MaxDelay time.Duration
}

// DefaultSarieConfig returns demo-friendly sarie settings.
func DefaultSarieConfig() SarieConfig {
	return SarieConfig{FailRate: 0.15, MinDelay: 50 * time.Millisecond, MaxDelay: 200 * time.Millisecond}
}

// SarieMock simulates the Saudi SARIE settlement rail with delay and intermittent failure.
type SarieMock struct {
	cfg SarieConfig
	rng *rand.Rand
}

// NewSarieMock constructs a mock sarie client.
func NewSarieMock(cfg SarieConfig) *SarieMock {
	if cfg.FailRate == 0 && cfg.MinDelay == 0 && cfg.MaxDelay == 0 {
		cfg = DefaultSarieConfig()
	}
	return &SarieMock{cfg: cfg, rng: rand.New(rand.NewSource(time.Now().UnixNano()))}
}

// Settle attempts mock settlement for a transfer reference.
func (s *SarieMock) Settle(ctx context.Context, reference string) (settlementports.SettlementResult, error) {
	delay := s.cfg.MinDelay
	if s.cfg.MaxDelay > s.cfg.MinDelay {
		delta := s.cfg.MaxDelay - s.cfg.MinDelay
		delay += time.Duration(s.rng.Int63n(int64(delta)))
	}

	select {
	case <-ctx.Done():
		return settlementports.SettlementResult{}, ctx.Err()
	case <-time.After(delay):
	}

	if s.rng.Float64() < s.cfg.FailRate {
		return settlementports.SettlementResult{
			Success:   false,
			Reference: reference,
			Delay:     delay,
			Error:     "sarie: settlement rejected by mock rail",
		}, nil
	}

	return settlementports.SettlementResult{
		Success:   true,
		Reference: reference,
		Delay:     delay,
	}, nil
}
