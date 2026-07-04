package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"unicode/utf8"

	shareddomain "github.com/mohsinazam/banking/internal/shared/domain"
)

const (
	minKeyLength = 8
	maxKeyLength = 128
)

// Key is a client-supplied idempotency token for money-moving operations.
type Key string

// ParseKey validates an idempotency key from a request header or body field.
func ParseKey(raw string) (Key, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", shareddomain.NewDomainError(shareddomain.ErrCodeValidation, "idempotency key is required")
	}
	if n := utf8.RuneCountInString(raw); n < minKeyLength || n > maxKeyLength {
		return "", shareddomain.NewDomainError(
			shareddomain.ErrCodeValidation,
			fmt.Sprintf("idempotency key must be %d–%d characters", minKeyLength, maxKeyLength),
		)
	}
	return Key(raw), nil
}

func (k Key) String() string {
	return string(k)
}

// Scope namespaces keys per operation type (transfer, withdrawal, etc.).
type Scope string

func (s Scope) String() string {
	return string(s)
}

// Fingerprint is a stable hash of the request body used to detect key reuse with different payloads.
type Fingerprint string

// FingerprintFromPayload hashes canonical request bytes.
func FingerprintFromPayload(payload []byte) Fingerprint {
	sum := sha256.Sum256(payload)
	return Fingerprint(hex.EncodeToString(sum[:]))
}

func (f Fingerprint) String() string {
	return string(f)
}

// LedgerTransactionID derives a deterministic ledger transaction id from an idempotency key.
// Defense in depth: even if Redis expires, the ledger PK prevents double posting.
func LedgerTransactionID(scope Scope, key Key) string {
	sum := sha256.Sum256([]byte(scope.String() + ":" + key.String()))
	return "tx_idem_" + hex.EncodeToString(sum[:16])
}
