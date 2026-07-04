package domain

import (
	"fmt"
	"strings"
	"unicode"
)

const (
	saIBANLength  = 24
	saCountryCode = "SA"
)

// IBAN is a validated Saudi Arabia IBAN value object.
type IBAN struct {
	value string // normalized, no spaces, uppercase
}

// ParseSAIBAN validates and constructs a Saudi IBAN.
// Accepts spaced or unspaced input (e.g. "SA03 8000 0000 6080 1016 7519").
func ParseSAIBAN(raw string) (IBAN, error) {
	normalized := normalizeIBAN(raw)

	if len(normalized) != saIBANLength {
		return IBAN{}, NewDomainError(
			ErrCodeInvalidIBAN,
			fmt.Sprintf("SA IBAN must be %d characters, got %d", saIBANLength, len(normalized)),
		)
	}

	if !strings.HasPrefix(normalized, saCountryCode) {
		return IBAN{}, NewDomainError(ErrCodeInvalidIBAN, "IBAN must start with SA")
	}

	for _, r := range normalized {
		if !unicode.IsDigit(r) && !unicode.IsUpper(r) {
			return IBAN{}, NewDomainError(ErrCodeInvalidIBAN, "IBAN contains invalid characters")
		}
	}

	if ibanMod97(normalized) != 1 {
		return IBAN{}, NewDomainError(ErrCodeInvalidIBAN, "IBAN check digits are invalid")
	}

	return IBAN{value: normalized}, nil
}

// String returns the normalized IBAN without spaces.
func (i IBAN) String() string {
	return i.value
}

// Formatted returns the IBAN in groups of four characters for display.
func (i IBAN) Formatted() string {
	var b strings.Builder
	for idx, r := range i.value {
		if idx > 0 && idx%4 == 0 {
			b.WriteByte(' ')
		}
		b.WriteRune(r)
	}
	return b.String()
}

// BankCode returns the two-digit Saudi bank identifier (positions 5–6).
// Example: SA03 80... → "80" (Al Rajhi).
func (i IBAN) BankCode() string {
	if len(i.value) < 6 {
		return ""
	}
	return i.value[4:6]
}

// AccountNumber returns the 18-digit domestic account segment.
func (i IBAN) AccountNumber() string {
	if len(i.value) < saIBANLength {
		return ""
	}
	return i.value[6:]
}

func normalizeIBAN(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.ReplaceAll(raw, " ", "")
	return strings.ToUpper(raw)
}

// ibanMod97 implements ISO 13616 mod-97 check without big integers.
func ibanMod97(iban string) int {
	rearranged := iban[4:] + iban[:4]

	remainder := 0
	for _, r := range rearranged {
		if r >= '0' && r <= '9' {
			remainder = (remainder*10 + int(r-'0')) % 97
			continue
		}
		val := int(r-'A') + 10
		remainder = (remainder*10 + val/10) % 97
		remainder = (remainder*10 + val%10) % 97
	}
	return remainder
}
