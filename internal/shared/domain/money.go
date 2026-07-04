package domain

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
)

const (
	// CurrencySAR is the ISO 4217 code for Saudi Riyal.
	CurrencySAR = "SAR"
	// HalalasPerSAR is the number of minor units in one riyal.
	HalalasPerSAR = 100
)

// Money is an immutable monetary amount stored as integer halalas.
// Never use float64 for money — rounding errors become real SAR losses.
type Money struct {
	halalas  int64
	currency string
}

// HalalasFromMinorUnits constructs money from a signed halala count.
func HalalasFromMinorUnits(halalas int64, currency string) (Money, error) {
	if currency != CurrencySAR {
		return Money{}, NewDomainError(ErrCodeInvalidMoney, fmt.Sprintf("unsupported currency %q", currency))
	}
	return Money{halalas: halalas, currency: currency}, nil
}

// SAR constructs money from whole riyals and halalas (0–99).
// Example: SAR(1500, 50) → 1500.50 SAR.
func SAR(riyals int64, halalas int64) (Money, error) {
	if halalas < 0 || halalas >= HalalasPerSAR {
		return Money{}, NewDomainError(ErrCodeInvalidMoney, "halalas must be between 0 and 99")
	}
	if riyals > 0 && halalas < 0 {
		return Money{}, NewDomainError(ErrCodeInvalidMoney, "halalas must be non-negative")
	}
	if riyals < 0 && halalas > 0 {
		return Money{}, NewDomainError(ErrCodeInvalidMoney, "halalas must be non-negative")
	}

	total := riyals*HalalasPerSAR + halalas
	return Money{halalas: total, currency: CurrencySAR}, nil
}

// MustSAR is a test/helper constructor that panics on invalid input.
func MustSAR(riyals int64, halalas int64) Money {
	m, err := SAR(riyals, halalas)
	if err != nil {
		panic(err)
	}
	return m
}

// ZeroSAR returns 0.00 SAR.
func ZeroSAR() Money {
	return Money{halalas: 0, currency: CurrencySAR}
}

// ParseSAR parses a decimal string such as "1500.50" or "-10.25" without floats.
func ParseSAR(value string) (Money, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return Money{}, NewDomainError(ErrCodeInvalidMoney, "amount is required")
	}

	negative := false
	if strings.HasPrefix(value, "-") {
		negative = true
		value = strings.TrimPrefix(value, "-")
	} else if strings.HasPrefix(value, "+") {
		value = strings.TrimPrefix(value, "+")
	}

	parts := strings.Split(value, ".")
	if len(parts) > 2 {
		return Money{}, NewDomainError(ErrCodeInvalidMoney, fmt.Sprintf("invalid amount %q", value))
	}

	riyals, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return Money{}, NewDomainError(ErrCodeInvalidMoney, fmt.Sprintf("invalid amount %q", value))
	}

	var halalas int64
	if len(parts) == 2 {
		frac := parts[1]
		if frac == "" {
			return Money{}, NewDomainError(ErrCodeInvalidMoney, fmt.Sprintf("invalid amount %q", value))
		}
		if len(frac) > 2 {
			return Money{}, NewDomainError(ErrCodeInvalidMoney, "amount must have at most 2 decimal places")
		}
		if len(frac) == 1 {
			frac += "0"
		}
		halalas, err = strconv.ParseInt(frac, 10, 64)
		if err != nil {
			return Money{}, NewDomainError(ErrCodeInvalidMoney, fmt.Sprintf("invalid amount %q", value))
		}
	}

	m, err := SAR(riyals, halalas)
	if err != nil {
		return Money{}, err
	}
	if negative {
		return m.Negate(), nil
	}
	return m, nil
}

// Halalas returns the amount in minor units.
func (m Money) Halalas() int64 {
	return m.halalas
}

// Currency returns the ISO currency code.
func (m Money) Currency() string {
	return m.currency
}

// IsZero reports whether the amount is exactly zero.
func (m Money) IsZero() bool {
	return m.halalas == 0
}

// IsNegative reports whether the amount is below zero.
func (m Money) IsNegative() bool {
	return m.halalas < 0
}

// IsPositive reports whether the amount is above zero.
func (m Money) IsPositive() bool {
	return m.halalas > 0
}

// Add returns m + other. Currencies must match.
func (m Money) Add(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, NewDomainError(ErrCodeInvalidMoney, "currency mismatch")
	}
	sum, err := safeAdd(m.halalas, other.halalas)
	if err != nil {
		return Money{}, WrapDomainError(ErrCodeInvalidMoney, "amount overflow", err)
	}
	return Money{halalas: sum, currency: m.currency}, nil
}

// Sub returns m - other. Currencies must match.
func (m Money) Sub(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, NewDomainError(ErrCodeInvalidMoney, "currency mismatch")
	}
	diff, err := safeSub(m.halalas, other.halalas)
	if err != nil {
		return Money{}, WrapDomainError(ErrCodeInvalidMoney, "amount overflow", err)
	}
	return Money{halalas: diff, currency: m.currency}, nil
}

// Negate returns -m.
func (m Money) Negate() Money {
	return Money{halalas: -m.halalas, currency: m.currency}
}

// Abs returns the absolute value of m.
func (m Money) Abs() Money {
	if m.halalas < 0 {
		return m.Negate()
	}
	return m
}

// Cmp compares two amounts. Returns -1 if m < other, 0 if equal, 1 if m > other.
func (m Money) Cmp(other Money) (int, error) {
	if m.currency != other.currency {
		return 0, NewDomainError(ErrCodeInvalidMoney, "currency mismatch")
	}
	switch {
	case m.halalas < other.halalas:
		return -1, nil
	case m.halalas > other.halalas:
		return 1, nil
	default:
		return 0, nil
	}
}

// GreaterThan reports whether m > other.
func (m Money) GreaterThan(other Money) (bool, error) {
	cmp, err := m.Cmp(other)
	return cmp > 0, err
}

// LessThan reports whether m < other.
func (m Money) LessThan(other Money) (bool, error) {
	cmp, err := m.Cmp(other)
	return cmp < 0, err
}

// String formats the amount as a decimal SAR string (e.g. "1500.50").
func (m Money) String() string {
	sign := ""
	abs := m.halalas
	if abs < 0 {
		sign = "-"
		abs = -abs
	}
	riyals := abs / HalalasPerSAR
	halalas := abs % HalalasPerSAR
	return fmt.Sprintf("%s%d.%02d", sign, riyals, halalas)
}

type moneyJSON struct {
	Currency string `json:"currency"`
	Halalas  int64  `json:"halalas"`
}

// MarshalJSON encodes money as {"currency":"SAR","halalas":150050}.
func (m Money) MarshalJSON() ([]byte, error) {
	return json.Marshal(moneyJSON{Currency: m.currency, Halalas: m.halalas})
}

// UnmarshalJSON decodes money from {"currency":"SAR","halalas":150050}.
func (m *Money) UnmarshalJSON(data []byte) error {
	var payload moneyJSON
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}
	parsed, err := HalalasFromMinorUnits(payload.Halalas, payload.Currency)
	if err != nil {
		return err
	}
	*m = parsed
	return nil
}

func safeAdd(a, b int64) (int64, error) {
	if b > 0 && a > math.MaxInt64-b {
		return 0, fmt.Errorf("overflow adding %d and %d", a, b)
	}
	if b < 0 && a < math.MinInt64-b {
		return 0, fmt.Errorf("overflow adding %d and %d", a, b)
	}
	return a + b, nil
}

func safeSub(a, b int64) (int64, error) {
	if b < 0 && a > math.MaxInt64+b {
		return 0, fmt.Errorf("overflow subtracting %d and %d", a, b)
	}
	if b > 0 && a < math.MinInt64+b {
		return 0, fmt.Errorf("overflow subtracting %d and %d", a, b)
	}
	return a - b, nil
}
