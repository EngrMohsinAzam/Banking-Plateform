package domain_test

import (
	"testing"

	"github.com/mohsinazam/banking/internal/idempotency/domain"
)

func TestParseKey(t *testing.T) {
	t.Parallel()

	key, err := domain.ParseKey("transfer-abc-123")
	if err != nil {
		t.Fatalf("ParseKey() error = %v", err)
	}
	if key.String() != "transfer-abc-123" {
		t.Fatalf("key = %q", key)
	}
}

func TestParseKeyRejectsEmpty(t *testing.T) {
	t.Parallel()

	if _, err := domain.ParseKey(""); err == nil {
		t.Fatal("expected error")
	}
}

func TestLedgerTransactionIDIsDeterministic(t *testing.T) {
	t.Parallel()

	scope := domain.Scope("transfer")
	key, _ := domain.ParseKey("client-key-001")

	a := domain.LedgerTransactionID(scope, key)
	b := domain.LedgerTransactionID(scope, key)
	if a != b {
		t.Fatalf("ids differ: %s vs %s", a, b)
	}
}

func TestFingerprintDetectsDifferentPayloads(t *testing.T) {
	t.Parallel()

	a := domain.FingerprintFromPayload([]byte(`{"amount":"10.00"}`))
	b := domain.FingerprintFromPayload([]byte(`{"amount":"20.00"}`))
	if a == b {
		t.Fatal("expected different fingerprints")
	}
}
