package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/mohsinazam/banking/internal/shared/domain"
)

func TestSAR(t *testing.T) {
	t.Parallel()

	m, err := domain.SAR(1500, 50)
	if err != nil {
		t.Fatalf("SAR() error = %v", err)
	}
	if got := m.Halalas(); got != 150050 {
		t.Fatalf("Halalas() = %d, want 150050", got)
	}
	if got := m.String(); got != "1500.50" {
		t.Fatalf("String() = %q, want %q", got, "1500.50")
	}
}

func TestParseSAR(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		halalas int64
	}{
		{name: "integer", input: "100", halalas: 10000},
		{name: "one decimal padded", input: "100.5", halalas: 10050},
		{name: "two decimals", input: "100.05", halalas: 10005},
		{name: "negative", input: "-10.25", halalas: -1025},
		{name: "trimmed", input: "  1.00  ", halalas: 100},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m, err := domain.ParseSAR(tt.input)
			if err != nil {
				t.Fatalf("ParseSAR() error = %v", err)
			}
			if m.Halalas() != tt.halalas {
				t.Fatalf("Halalas() = %d, want %d", m.Halalas(), tt.halalas)
			}
		})
	}
}

func TestParseSARRejectsInvalid(t *testing.T) {
	t.Parallel()

	invalid := []string{"", "abc", "1.234", "1..0", "1."}
	for _, input := range invalid {
		if _, err := domain.ParseSAR(input); err == nil {
			t.Fatalf("ParseSAR(%q) expected error", input)
		}
	}
}

func TestMoneyArithmetic(t *testing.T) {
	t.Parallel()

	a := domain.MustSAR(100, 0)
	b := domain.MustSAR(25, 50)

	sum, err := a.Add(b)
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	if sum.String() != "125.50" {
		t.Fatalf("sum = %s, want 125.50", sum.String())
	}

	diff, err := sum.Sub(b)
	if err != nil {
		t.Fatalf("Sub() error = %v", err)
	}
	if diff.String() != "100.00" {
		t.Fatalf("diff = %s, want 100.00", diff.String())
	}
}

func TestMoneyJSON(t *testing.T) {
	t.Parallel()

	original := domain.MustSAR(10, 25)
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var decoded domain.Money
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if decoded.Halalas() != original.Halalas() {
		t.Fatalf("Halalas() = %d, want %d", decoded.Halalas(), original.Halalas())
	}
}

func TestMoneyNeverUsesFloat(t *testing.T) {
	t.Parallel()

	// Classic float bug: 0.1 + 0.2 != 0.3. Integer halalas avoid this entirely.
	a, _ := domain.ParseSAR("0.10")
	b, _ := domain.ParseSAR("0.20")
	sum, err := a.Add(b)
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	if sum.String() != "0.30" {
		t.Fatalf("sum = %s, want 0.30", sum.String())
	}
}
