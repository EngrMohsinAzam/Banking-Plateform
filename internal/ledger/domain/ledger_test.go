package domain_test

import (
	"testing"
	"time"

	shareddomain "github.com/mohsinazam/banking/internal/shared/domain"
	"github.com/mohsinazam/banking/internal/ledger/domain"
)

func mustMoney(t *testing.T, riyals, halalas int64) shareddomain.Money {
	t.Helper()
	m, err := shareddomain.SAR(riyals, halalas)
	if err != nil {
		t.Fatalf("SAR() error = %v", err)
	}
	return m
}

func mustEntry(
	t *testing.T,
	id string,
	accountID string,
	side domain.EntrySide,
	riyals, halalas int64,
) domain.Entry {
	t.Helper()
	e, err := domain.NewEntry(
		domain.EntryID(id),
		domain.AccountID(accountID),
		side,
		mustMoney(t, riyals, halalas),
	)
	if err != nil {
		t.Fatalf("NewEntry() error = %v", err)
	}
	return e
}

func TestNewTransactionBalanced(t *testing.T) {
	t.Parallel()

	// P2P transfer between two customer wallets (both liabilities).
	entries := []domain.Entry{
		mustEntry(t, "e1", "wallet-alice", domain.SideDebit, 100, 0),
		mustEntry(t, "e2", "wallet-bob", domain.SideCredit, 100, 0),
	}

	tx, err := domain.NewTransaction(
		domain.TransactionID("tx-1"),
		"transfer alice to bob",
		entries,
		time.Date(2026, 6, 25, 12, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("NewTransaction() error = %v", err)
	}
	if !tx.IsBalanced() {
		t.Fatal("expected balanced transaction")
	}
}

func TestNewTransactionRejectsUnbalanced(t *testing.T) {
	t.Parallel()

	entries := []domain.Entry{
		mustEntry(t, "e1", "wallet-alice", domain.SideDebit, 100, 0),
		mustEntry(t, "e2", "wallet-bob", domain.SideCredit, 99, 0),
	}

	_, err := domain.NewTransaction(
		domain.TransactionID("tx-bad"),
		"unbalanced",
		entries,
		time.Now().UTC(),
	)
	if err == nil {
		t.Fatal("expected error for unbalanced transaction")
	}
	if !shareddomain.IsDomainCode(err, domain.ErrCodeUnbalancedTransaction) {
		t.Fatalf("expected UNBALANCED_TRANSACTION, got %v", err)
	}
}

func TestNewTransactionRejectsSingleEntry(t *testing.T) {
	t.Parallel()

	entries := []domain.Entry{
		mustEntry(t, "e1", "wallet-alice", domain.SideDebit, 100, 0),
	}

	_, err := domain.NewTransaction(
		domain.TransactionID("tx-single"),
		"single entry",
		entries,
		time.Now().UTC(),
	)
	if err == nil {
		t.Fatal("expected error for single-entry transaction")
	}
}

func TestNewEntryRejectsNonPositiveAmount(t *testing.T) {
	t.Parallel()

	zero := shareddomain.ZeroSAR()
	_, err := domain.NewEntry(
		domain.EntryID("e-zero"),
		domain.AccountID("wallet-alice"),
		domain.SideDebit,
		zero,
	)
	if err == nil {
		t.Fatal("expected error for zero amount entry")
	}
}

func TestBalanceFromEntriesCustomerWallet(t *testing.T) {
	t.Parallel()

	alice, err := domain.NewAccount(domain.AccountID("wallet-alice"), domain.AccountTypeLiability, "Alice Wallet")
	if err != nil {
		t.Fatalf("NewAccount() error = %v", err)
	}

	// Deposit 500 SAR (credit liability), then transfer out 100 SAR (debit liability).
	entries := []domain.Entry{
		mustEntry(t, "e1", "wallet-alice", domain.SideCredit, 500, 0),
		mustEntry(t, "e2", "wallet-alice", domain.SideDebit, 100, 0),
	}

	balance, err := domain.BalanceFromEntries(alice, entries)
	if err != nil {
		t.Fatalf("BalanceFromEntries() error = %v", err)
	}
	if balance.String() != "400.00" {
		t.Fatalf("balance = %s, want 400.00", balance.String())
	}
}

func TestBooksBalanceAfterTransfer(t *testing.T) {
	t.Parallel()

	alice, _ := domain.NewAccount(domain.AccountID("wallet-alice"), domain.AccountTypeLiability, "Alice")
	bob, _ := domain.NewAccount(domain.AccountID("wallet-bob"), domain.AccountTypeLiability, "Bob")

	entries := []domain.Entry{
		mustEntry(t, "e1", "wallet-alice", domain.SideDebit, 250, 50),
		mustEntry(t, "e2", "wallet-bob", domain.SideCredit, 250, 50),
	}

	tx, err := domain.NewTransaction(
		domain.TransactionID("tx-transfer"),
		"p2p transfer",
		entries,
		time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("NewTransaction() error = %v", err)
	}

	allEntries := tx.Entries()
	aliceBal, err := domain.BalanceFromEntries(alice, allEntries)
	if err != nil {
		t.Fatalf("alice balance error = %v", err)
	}
	bobBal, err := domain.BalanceFromEntries(bob, allEntries)
	if err != nil {
		t.Fatalf("bob balance error = %v", err)
	}

	if aliceBal.String() != "-250.50" {
		t.Fatalf("alice = %s, want -250.50", aliceBal.String())
	}
	if bobBal.String() != "250.50" {
		t.Fatalf("bob = %s, want 250.50", bobBal.String())
	}

	// Liability-to-liability transfer: customer funds move, but the banking system nets to zero.
	net, err := aliceBal.Add(bobBal)
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	if !net.IsZero() {
		t.Fatalf("p2p net = %s, want 0.00", net.String())
	}

	if err := domain.VerifyEntriesBalanced(allEntries); err != nil {
		t.Fatalf("VerifyEntriesBalanced() error = %v", err)
	}
}

func TestFourLegSettlementTransaction(t *testing.T) {
	t.Parallel()

	// Customer debit → settlement suspense credit → settlement debit → nostro credit.
	// Models internal routing before sarie settlement (Step 9).
	customer, _ := domain.NewAccount(domain.AccountID("wallet-alice"), domain.AccountTypeLiability, "Alice")
	suspense, _ := domain.NewAccount(domain.AccountID("suspense"), domain.AccountTypeLiability, "Suspense")
	nostro, _ := domain.NewAccount(domain.AccountID("nostro"), domain.AccountTypeAsset, "Nostro")

	entries := []domain.Entry{
		mustEntry(t, "e1", "wallet-alice", domain.SideDebit, 1000, 0),
		mustEntry(t, "e2", "suspense", domain.SideCredit, 1000, 0),
		mustEntry(t, "e3", "suspense", domain.SideDebit, 1000, 0),
		mustEntry(t, "e4", "nostro", domain.SideCredit, 1000, 0),
	}

	tx, err := domain.NewTransaction(domain.TransactionID("tx-4leg"), "outbound transfer", entries, time.Now().UTC())
	if err != nil {
		t.Fatalf("NewTransaction() error = %v", err)
	}
	if !tx.IsBalanced() {
		t.Fatal("expected balanced transaction")
	}

	allEntries := tx.Entries()
	if err := domain.VerifyEntriesBalanced(allEntries); err != nil {
		t.Fatalf("VerifyEntriesBalanced() error = %v", err)
	}

	suspenseBal, err := domain.BalanceFromEntries(suspense, allEntries)
	if err != nil {
		t.Fatalf("suspense balance error = %v", err)
	}
	if !suspenseBal.IsZero() {
		t.Fatalf("suspense = %s, want 0.00 (cleared)", suspenseBal.String())
	}

	customerBal, err := domain.BalanceFromEntries(customer, allEntries)
	if err != nil {
		t.Fatalf("customer balance error = %v", err)
	}
	if customerBal.String() != "-1000.00" {
		t.Fatalf("customer = %s, want -1000.00", customerBal.String())
	}

	nostroBal, err := domain.BalanceFromEntries(nostro, allEntries)
	if err != nil {
		t.Fatalf("nostro balance error = %v", err)
	}
	if nostroBal.String() != "-1000.00" {
		t.Fatalf("nostro = %s, want -1000.00", nostroBal.String())
	}
}

func TestTransactionDefensiveCopy(t *testing.T) {
	t.Parallel()

	entries := []domain.Entry{
		mustEntry(t, "e1", "wallet-alice", domain.SideDebit, 10, 0),
		mustEntry(t, "e2", "wallet-bob", domain.SideCredit, 10, 0),
	}

	tx, err := domain.NewTransaction(domain.TransactionID("tx-copy"), "copy test", entries, time.Now().UTC())
	if err != nil {
		t.Fatalf("NewTransaction() error = %v", err)
	}

	mutated := tx.Entries()
	mutated[0] = mustEntry(t, "eX", "wallet-alice", domain.SideDebit, 999, 0)

	if len(tx.Entries()) != 2 {
		t.Fatal("transaction entries should remain unchanged")
	}
	if !tx.IsBalanced() {
		t.Fatal("transaction should still be balanced after external slice mutation")
	}
}
