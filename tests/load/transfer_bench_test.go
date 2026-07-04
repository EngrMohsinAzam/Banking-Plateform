package load_test

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"

	complianceapp "github.com/mohsinazam/banking/internal/compliance/app"
	fraudapp "github.com/mohsinazam/banking/internal/fraud/app"
	fraudredis "github.com/mohsinazam/banking/internal/fraud/adapters/redis"
	idempotencyapp "github.com/mohsinazam/banking/internal/idempotency/app"
	idempotencyredis "github.com/mohsinazam/banking/internal/idempotency/adapters/redis"
	ledgerdomain "github.com/mohsinazam/banking/internal/ledger/domain"
	shareddomain "github.com/mohsinazam/banking/internal/shared/domain"
	uowpostgres "github.com/mohsinazam/banking/internal/platform/uow/postgres"
	transferapp "github.com/mohsinazam/banking/internal/transfer/app"
	transferdomain "github.com/mohsinazam/banking/internal/transfer/domain"
)

// BenchmarkTransferExecute measures in-memory transfer throughput (no Docker).
// Run: go test -bench=. -benchmem ./tests/load/...
func BenchmarkTransferExecute(b *testing.B) {
	svc := newBenchService(b)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := svc.Execute(ctx, transferdomain.Command{
			IdempotencyKey:  fmt.Sprintf("bench-key-%d", i),
			FromAccountID:   ledgerdomain.AccountID("wallet-alice"),
			ToAccountID:     ledgerdomain.AccountID("wallet-bob"),
			Amount:          shareddomain.MustSAR(1, 0),
			BeneficiaryIBAN: "SA0380000000608010167519",
			Description:     "load benchmark",
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func newBenchService(b *testing.B) *transferapp.Service {
	b.Helper()

	mr := miniredis.RunT(b)
	client := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	guard := idempotencyapp.NewGuard(idempotencyredis.NewStore(client), idempotencyapp.DefaultConfig())
	fraud := fraudapp.NewChecker(fraudredis.NewVelocityStore(client), fraudapp.Config{MaxHourlyTransfers: 1_000_000})

	ledger := &benchLedger{
		accounts: map[ledgerdomain.AccountID]ledgerdomain.Account{},
		entries:  map[ledgerdomain.AccountID][]ledgerdomain.Entry{},
	}
	alice, _ := ledgerdomain.NewAccount(ledgerdomain.AccountID("wallet-alice"), ledgerdomain.AccountTypeLiability, "Alice")
	bob, _ := ledgerdomain.NewAccount(ledgerdomain.AccountID("wallet-bob"), ledgerdomain.AccountTypeLiability, "Bob")
	ledger.accounts[alice.ID()] = alice
	ledger.accounts[bob.ID()] = bob
	credit, _ := ledgerdomain.NewEntry("seed-c", alice.ID(), ledgerdomain.SideCredit, shareddomain.MustSAR(1_000_000, 0))
	debit, _ := ledgerdomain.NewEntry("seed-d", ledgerdomain.AccountID("suspense"), ledgerdomain.SideDebit, shareddomain.MustSAR(1_000_000, 0))
	ledger.entries[alice.ID()] = []ledgerdomain.Entry{credit}
	ledger.entries[ledgerdomain.AccountID("suspense")] = []ledgerdomain.Entry{debit}

	coordinator := &benchCoordinator{ledger: ledger}
	return transferapp.NewService(ledger, coordinator, guard, fraud, complianceapp.NewChecker(), &benchSagaStore{}, nil)
}

type benchLedger struct {
	mu       sync.Mutex
	accounts map[ledgerdomain.AccountID]ledgerdomain.Account
	entries  map[ledgerdomain.AccountID][]ledgerdomain.Entry
	posted   []ledgerdomain.TransactionID
}

func (m *benchLedger) GetAccount(_ context.Context, id ledgerdomain.AccountID) (ledgerdomain.Account, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	account, ok := m.accounts[id]
	if !ok {
		return ledgerdomain.Account{}, shareddomain.NewDomainError(shareddomain.ErrCodeNotFound, "account not found")
	}
	return account, nil
}

func (m *benchLedger) PostTransaction(_ context.Context, tx ledgerdomain.Transaction) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, account := range m.accounts {
		existing := m.entries[account.ID()]
		projected, err := ledgerdomain.ApplyEntries(account, benchBalance(account, existing), tx.Entries())
		if err != nil {
			return err
		}
		if projected.IsNegative() {
			return shareddomain.NewDomainError(shareddomain.ErrCodeInsufficientFunds, "insufficient funds")
		}
	}
	for _, entry := range tx.Entries() {
		m.entries[entry.AccountID()] = append(m.entries[entry.AccountID()], entry)
	}
	m.posted = append(m.posted, tx.ID())
	return nil
}

func benchBalance(account ledgerdomain.Account, entries []ledgerdomain.Entry) shareddomain.Money {
	bal, _ := ledgerdomain.BalanceFromEntries(account, entries)
	return bal
}

type benchCoordinator struct {
	ledger *benchLedger
}

func (c *benchCoordinator) PostTransfer(_ context.Context, commit uowpostgres.TransferCommit) error {
	return c.ledger.PostTransaction(context.Background(), commit.LedgerTx)
}

type benchSagaStore struct{}

func (benchSagaStore) CreateSaga(context.Context, transferdomain.SagaRecord) error { return nil }
func (benchSagaStore) UpdateSagaState(context.Context, string, transferdomain.SagaState, map[string]string) error {
	return nil
}
func (benchSagaStore) GetSagaByIdempotencyKey(context.Context, string) (transferdomain.SagaRecord, error) {
	return transferdomain.SagaRecord{}, nil
}
func (benchSagaStore) GetSagaByTransactionID(context.Context, string) (transferdomain.SagaRecord, error) {
	return transferdomain.SagaRecord{}, nil
}
func (benchSagaStore) GetSettlementBySagaID(context.Context, string) (transferdomain.SettlementRecord, error) {
	return transferdomain.SettlementRecord{}, nil
}
func (benchSagaStore) CreateSettlement(context.Context, transferdomain.SettlementRecord) error { return nil }
func (benchSagaStore) ClaimPendingSettlements(context.Context, int) ([]transferdomain.SettlementRecord, error) {
	return nil, nil
}
func (benchSagaStore) UpdateSettlementStatus(context.Context, string, transferdomain.SettlementStatus, string) error {
	return nil
}
func (benchSagaStore) GetSaga(context.Context, string) (transferdomain.SagaRecord, error) {
	return transferdomain.SagaRecord{}, nil
}
