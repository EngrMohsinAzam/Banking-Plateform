package app_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	complianceapp "github.com/mohsinazam/banking/internal/compliance/app"
	shareddomain "github.com/mohsinazam/banking/internal/shared/domain"
	"github.com/mohsinazam/banking/internal/transfer/ports"
)

func TestComplianceRejectsSanctionedIBAN(t *testing.T) {
	checker := complianceapp.NewChecker()
	err := checker.Check(context.Background(), ports.ComplianceCheckInput{
		BeneficiaryIBAN: "SA4420000001234567891234",
	})
	require.Error(t, err)
	require.True(t, shareddomain.IsDomainCode(err, shareddomain.ErrCodeForbidden))
}

func TestComplianceRejectsSanctionedName(t *testing.T) {
	checker := complianceapp.NewChecker()
	err := checker.Check(context.Background(), ports.ComplianceCheckInput{
		BeneficiaryIBAN: "SA0380000000608010167519",
		BeneficiaryName: "Sanctioned Entity",
	})
	require.Error(t, err)
	require.True(t, shareddomain.IsDomainCode(err, shareddomain.ErrCodeForbidden))
}

func TestComplianceAllowsCleanBeneficiary(t *testing.T) {
	checker := complianceapp.NewChecker()
	err := checker.Check(context.Background(), ports.ComplianceCheckInput{
		BeneficiaryIBAN: "SA0380000000608010167519",
		BeneficiaryName: "Mohsin Azam",
	})
	require.NoError(t, err)
}
