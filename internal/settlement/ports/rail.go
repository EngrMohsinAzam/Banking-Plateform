package ports

import (
	"context"
	"time"
)

// SettlementResult is the outcome of an external rail settlement attempt.
type SettlementResult struct {
	Success   bool
	Reference string
	Delay     time.Duration
	Error     string
}

// SettlementRail settles transfers on an external payment rail (e.g. SARIE).
type SettlementRail interface {
	Settle(ctx context.Context, reference string) (SettlementResult, error)
}
