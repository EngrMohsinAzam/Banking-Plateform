package app

import (
	"context"
	"log/slog"
	"strings"

	shareddomain "github.com/mohsinazam/banking/internal/shared/domain"
	"github.com/mohsinazam/banking/internal/transfer/ports"
)

var blockedIBANs = map[string]struct{}{
	"SA4420000001234567891234": {},
}

var blockedNames = map[string]struct{}{
	"SANCTIONED ENTITY": {},
	"BLOCKED CUSTOMER":  {},
}

// Checker performs compliance screening before money movement.
type Checker struct {
	kyc CustomerRegistry
}

// NewChecker constructs a compliance checker with built-in mock lists and KYC registry.
func NewChecker() *Checker {
	return &Checker{kyc: NewMockRegistry()}
}

// Check returns an error when the beneficiary matches sanctions or KYC deny lists.
func (c *Checker) Check(ctx context.Context, input ports.ComplianceCheckInput) error {
	iban, err := shareddomain.ParseSAIBAN(input.BeneficiaryIBAN)
	if err != nil {
		return err
	}
	if _, blocked := blockedIBANs[iban.String()]; blocked {
		slog.Warn("compliance blocked transfer", "reason", "sanctioned_iban", "iban", iban.String())
		return shareddomain.NewDomainError(shareddomain.ErrCodeForbidden, "beneficiary IBAN is sanctioned")
	}

	name := strings.ToUpper(strings.TrimSpace(input.BeneficiaryName))
	if name != "" {
		if _, blocked := blockedNames[name]; blocked {
			slog.Warn("compliance blocked transfer", "reason", "sanctioned_name", "name", name)
			return shareddomain.NewDomainError(shareddomain.ErrCodeForbidden, "beneficiary is sanctioned")
		}
	}

	approved, err := c.kyc.IsKYCApproved(ctx, iban.String(), input.BeneficiaryName)
	if err != nil {
		return err
	}
	if !approved {
		slog.Warn("compliance blocked transfer", "reason", "kyc_not_approved", "name", input.BeneficiaryName)
		return shareddomain.NewDomainError(shareddomain.ErrCodeForbidden, "beneficiary KYC not approved")
	}
	return nil
}
