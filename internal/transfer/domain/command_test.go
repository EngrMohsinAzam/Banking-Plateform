package domain_test

import (
	"testing"

	shareddomain "github.com/mohsinazam/banking/internal/shared/domain"
	ledgerdomain "github.com/mohsinazam/banking/internal/ledger/domain"
	"github.com/mohsinazam/banking/internal/transfer/domain"
)

func TestCommandValidate(t *testing.T) {
	t.Parallel()

	cmd := domain.Command{
		IdempotencyKey:  "transfer-valid-01",
		FromAccountID:   ledgerdomain.AccountID("wallet-alice"),
		ToAccountID:     ledgerdomain.AccountID("wallet-bob"),
		Amount:          shareddomain.MustSAR(10, 0),
		BeneficiaryIBAN: "SA0380000000608010167519",
	}
	if err := cmd.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestCommandValidateRejectsSameAccount(t *testing.T) {
	t.Parallel()

	cmd := domain.Command{
		IdempotencyKey:  "transfer-valid-02",
		FromAccountID:   ledgerdomain.AccountID("wallet-alice"),
		ToAccountID:     ledgerdomain.AccountID("wallet-alice"),
		Amount:          shareddomain.MustSAR(10, 0),
		BeneficiaryIBAN: "SA0380000000608010167519",
	}
	if err := cmd.Validate(); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestCommandFingerprintStable(t *testing.T) {
	t.Parallel()

	cmd := domain.Command{
		IdempotencyKey:  "transfer-valid-03",
		FromAccountID:   ledgerdomain.AccountID("wallet-alice"),
		ToAccountID:     ledgerdomain.AccountID("wallet-bob"),
		Amount:          shareddomain.MustSAR(10, 0),
		BeneficiaryIBAN: "SA03 8000 0000 6080 1016 7519",
	}
	a := cmd.Fingerprint()
	b := cmd.Fingerprint()
	if a != b {
		t.Fatalf("fingerprints differ: %s vs %s", a, b)
	}
}
