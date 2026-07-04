package domain_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/mohsinazam/banking/internal/shared/domain"
)

func TestDomainErrorWrapping(t *testing.T) {
	t.Parallel()

	root := fmt.Errorf("postgres timeout")
	err := domain.WrapDomainError(domain.ErrCodeConflict, "duplicate transfer", root)

	var de *domain.DomainError
	if !errors.As(err, &de) {
		t.Fatal("expected DomainError in chain")
	}
	if de.Code != domain.ErrCodeConflict {
		t.Fatalf("Code = %q, want %q", de.Code, domain.ErrCodeConflict)
	}
	if !errors.Is(err, root) {
		t.Fatal("expected wrapped root error")
	}
}

func TestIsDomainCode(t *testing.T) {
	t.Parallel()

	err := domain.NewDomainError(domain.ErrCodeInvalidIBAN, "bad iban")
	if !domain.IsDomainCode(err, domain.ErrCodeInvalidIBAN) {
		t.Fatal("IsDomainCode should match")
	}
	if domain.IsDomainCode(err, domain.ErrCodeInvalidMoney) {
		t.Fatal("IsDomainCode should not match different code")
	}
}
