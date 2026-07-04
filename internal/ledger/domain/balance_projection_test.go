package domain_test

import (
	"testing"

	shareddomain "github.com/mohsinazam/banking/internal/shared/domain"
	"github.com/mohsinazam/banking/internal/ledger/domain"
)

func TestApplyEntriesProjectsLiabilityBalance(t *testing.T) {
	t.Parallel()

	account, err := domain.NewAccount(domain.AccountID("wallet-alice"), domain.AccountTypeLiability, "Alice")
	if err != nil {
		t.Fatalf("NewAccount() error = %v", err)
	}

	start := shareddomain.MustSAR(500, 0)
	amount := shareddomain.MustSAR(100, 0)
	entry, err := domain.NewEntry(domain.EntryID("e1"), account.ID(), domain.SideDebit, amount)
	if err != nil {
		t.Fatalf("NewEntry() error = %v", err)
	}

	projected, err := domain.ApplyEntries(account, start, []domain.Entry{entry})
	if err != nil {
		t.Fatalf("ApplyEntries() error = %v", err)
	}
	if projected.String() != "400.00" {
		t.Fatalf("projected = %s, want 400.00", projected.String())
	}
}
