package domain_test

import (
	"testing"

	"github.com/mohsinazam/banking/internal/shared/domain"
)

func TestParseSAIBANValid(t *testing.T) {
	t.Parallel()

	// ECBS / IBAN.com reference examples for Saudi Arabia.
	valid := []string{
		"SA0380000000608010167519",
		"SA03 8000 0000 6080 1016 7519",
		"SA4420000001234567891234",
	}

	for _, raw := range valid {
		iban, err := domain.ParseSAIBAN(raw)
		if err != nil {
			t.Fatalf("ParseSAIBAN(%q) error = %v", raw, err)
		}
		if iban.BankCode() == "" {
			t.Fatalf("BankCode() empty for %q", raw)
		}
	}
}

func TestParseSAIBANInvalid(t *testing.T) {
	t.Parallel()

	invalid := []string{
		"",
		"SA038000000060801016751",  // too short
		"SA03800000006080101675190", // too long
		"AE070331234567890123456",   // wrong country
		"SA9980000000608010167519",  // bad check digits
		"SA03-8000-0000-6080-1016-7519",
	}

	for _, raw := range invalid {
		if _, err := domain.ParseSAIBAN(raw); err == nil {
			t.Fatalf("ParseSAIBAN(%q) expected error", raw)
		}
	}
}

func TestIBANFormatted(t *testing.T) {
	t.Parallel()

	iban, err := domain.ParseSAIBAN("SA0380000000608010167519")
	if err != nil {
		t.Fatalf("ParseSAIBAN() error = %v", err)
	}

	want := "SA03 8000 0000 6080 1016 7519"
	if got := iban.Formatted(); got != want {
		t.Fatalf("Formatted() = %q, want %q", got, want)
	}
}

func TestIBANBankCodeAndAccount(t *testing.T) {
	t.Parallel()

	iban, err := domain.ParseSAIBAN("SA0380000000608010167519")
	if err != nil {
		t.Fatalf("ParseSAIBAN() error = %v", err)
	}

	if got := iban.BankCode(); got != "80" {
		t.Fatalf("BankCode() = %q, want 80", got)
	}
	if got := iban.AccountNumber(); got != "000000608010167519" {
		t.Fatalf("AccountNumber() = %q, want 000000608010167519", got)
	}
}
