package app

import (
	"context"
	"strings"
)

// CustomerRegistry checks KYC status for outbound transfers.
type CustomerRegistry interface {
	IsKYCApproved(ctx context.Context, beneficiaryIBAN, beneficiaryName string) (bool, error)
}

// MockRegistry is an in-memory KYC registry for portfolio demos.
type MockRegistry struct {
	blocked map[string]struct{}
}

// NewMockRegistry returns a registry with demo blocked customers.
func NewMockRegistry() *MockRegistry {
	return &MockRegistry{
		blocked: map[string]struct{}{
			"KYC PENDING CUSTOMER": {},
			"UNVERIFIED USER":      {},
		},
	}
}

// IsKYCApproved returns false when the beneficiary name is on the blocked list.
func (r *MockRegistry) IsKYCApproved(_ context.Context, _ string, beneficiaryName string) (bool, error) {
	name := strings.ToUpper(strings.TrimSpace(beneficiaryName))
	if name == "" {
		return true, nil
	}
	_, blocked := r.blocked[name]
	return !blocked, nil
}
