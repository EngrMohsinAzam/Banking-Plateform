package app_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	complianceapp "github.com/mohsinazam/banking/internal/compliance/app"
	shareddomain "github.com/mohsinazam/banking/internal/shared/domain"
	"github.com/mohsinazam/banking/internal/transfer/ports"
)

func TestComplianceRejectsKYCNotApproved(t *testing.T) {
	checker := complianceapp.NewChecker()
	err := checker.Check(context.Background(), ports.ComplianceCheckInput{
		BeneficiaryIBAN: "SA0380000000608010167519",
		BeneficiaryName: "KYC Pending Customer",
	})
	require.Error(t, err)
	require.True(t, shareddomain.IsDomainCode(err, shareddomain.ErrCodeForbidden))
}

func TestComplianceAllowsApprovedKYC(t *testing.T) {
	checker := complianceapp.NewChecker()
	err := checker.Check(context.Background(), ports.ComplianceCheckInput{
		BeneficiaryIBAN: "SA0380000000608010167519",
		BeneficiaryName: "Verified Customer",
	})
	require.NoError(t, err)
}
