package app_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"

	fraudapp "github.com/mohsinazam/banking/internal/fraud/app"
	fraudredis "github.com/mohsinazam/banking/internal/fraud/adapters/redis"
	complianceapp "github.com/mohsinazam/banking/internal/compliance/app"
	idempotencyapp "github.com/mohsinazam/banking/internal/idempotency/app"
	idempotencyredis "github.com/mohsinazam/banking/internal/idempotency/adapters/redis"
	ledgerdomain "github.com/mohsinazam/banking/internal/ledger/domain"
	outboxdomain "github.com/mohsinazam/banking/internal/outbox/domain"
	shareddomain "github.com/mohsinazam/banking/internal/shared/domain"
	uowpostgres "github.com/mohsinazam/banking/internal/platform/uow/postgres"
	transferapp "github.com/mohsinazam/banking/internal/transfer/app"
	transferdomain "github.com/mohsinazam/banking/internal/transfer/domain"
	"github.com/mohsinazam/banking/internal/transfer/ports"
)

type memoryLedger struct {
	mu       sync.Mutex
	accounts map[ledgerdomain.AccountID]ledgerdomain.Account
	entries  map[ledgerdomain.AccountID][]ledgerdomain.Entry
	posted   []ledgerdomain.TransactionID
	events   []outboxdomain.Event
	sagas    []transferdomain.SagaRecord
}

func newMemoryLedger() *memoryLedger {
	return &memoryLedger{
		accounts: make(map[ledgerdomain.AccountID]ledgerdomain.Account),
		entries:  make(map[ledgerdomain.AccountID][]ledgerdomain.Entry),
	}
}

func (m *memoryLedger) seed(account ledgerdomain.Account, funding shareddomain.Money) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.accounts[account.ID()] = account
	credit, err := ledgerdomain.NewEntry(
		ledgerdomain.EntryID(account.ID().String()+"-fund-c"),
		account.ID(),
		ledgerdomain.SideCredit,
		funding,
	)
	if err != nil {
		return err
	}
	debit, err := ledgerdomain.NewEntry(
		ledgerdomain.EntryID(account.ID().String()+"-fund-d"),
		ledgerdomain.AccountID("suspense"),
		ledgerdomain.SideDebit,
		funding,
	)
	if err != nil {
		return err
	}
	m.entries[account.ID()] = append(m.entries[account.ID()], credit)
	m.entries[ledgerdomain.AccountID("suspense")] = append(m.entries[ledgerdomain.AccountID("suspense")], debit)
	return nil
}

func (m *memoryLedger) GetAccount(_ context.Context, id ledgerdomain.AccountID) (ledgerdomain.Account, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	account, ok := m.accounts[id]
	if !ok {
		return ledgerdomain.Account{}, shareddomain.NewDomainError(shareddomain.ErrCodeNotFound, "account not found")
	}
	return account, nil
}

func (m *memoryLedger) PostTransaction(_ context.Context, tx ledgerdomain.Transaction) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, id := range m.posted {
		if id == tx.ID() {
			return shareddomain.NewDomainError(shareddomain.ErrCodeConflict, "transaction already posted")
		}
	}

	for _, account := range m.accounts {
		existing := m.entries[account.ID()]
		projected, err := ledgerdomain.ApplyEntries(account, balanceOf(account, existing), tx.Entries())
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

func (m *memoryLedger) PostTransactionWithEvents(_ context.Context, tx ledgerdomain.Transaction, events []outboxdomain.Event) error {
	if err := m.PostTransaction(context.Background(), tx); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, events...)
	return nil
}

type memoryCoordinator struct {
	ledger *memoryLedger
}

func (m *memoryCoordinator) PostTransfer(_ context.Context, commit uowpostgres.TransferCommit) error {
	if err := m.ledger.PostTransaction(context.Background(), commit.LedgerTx); err != nil {
		return err
	}
	m.ledger.mu.Lock()
	defer m.ledger.mu.Unlock()
	m.ledger.events = append(m.ledger.events, commit.Events...)
	m.ledger.sagas = append(m.ledger.sagas, commit.Saga)
	return nil
}

type memorySagaStore struct {
	mu sync.Mutex
}

func (m *memorySagaStore) CreateSaga(context.Context, transferdomain.SagaRecord) error { return nil }
func (m *memorySagaStore) UpdateSagaState(context.Context, string, transferdomain.SagaState, map[string]string) error {
	return nil
}
func (m *memorySagaStore) GetSagaByIdempotencyKey(context.Context, string) (transferdomain.SagaRecord, error) {
	return transferdomain.SagaRecord{}, nil
}
func (m *memorySagaStore) GetSagaByTransactionID(context.Context, string) (transferdomain.SagaRecord, error) {
	return transferdomain.SagaRecord{}, shareddomain.NewDomainError(shareddomain.ErrCodeNotFound, "saga not found")
}
func (m *memorySagaStore) GetSettlementBySagaID(context.Context, string) (transferdomain.SettlementRecord, error) {
	return transferdomain.SettlementRecord{}, shareddomain.NewDomainError(shareddomain.ErrCodeNotFound, "settlement not found")
}
func (m *memorySagaStore) CreateSettlement(context.Context, transferdomain.SettlementRecord) error {
	return nil
}
func (m *memorySagaStore) ClaimPendingSettlements(context.Context, int) ([]transferdomain.SettlementRecord, error) {
	return nil, nil
}
func (m *memorySagaStore) UpdateSettlementStatus(context.Context, string, transferdomain.SettlementStatus, string) error {
	return nil
}
func (m *memorySagaStore) GetSaga(context.Context, string) (transferdomain.SagaRecord, error) {
	return transferdomain.SagaRecord{}, nil
}

func balanceOf(account ledgerdomain.Account, entries []ledgerdomain.Entry) shareddomain.Money {
	bal, _ := ledgerdomain.BalanceFromEntries(account, entries)
	return bal
}

func setupTransferService(t *testing.T) (*transferapp.Service, *memoryLedger) {
	t.Helper()

	mr := miniredis.RunT(t)
	client := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	store := idempotencyredis.NewStore(client)
	guard := idempotencyapp.NewGuard(store, idempotencyapp.Config{
		ProcessingTTL: time.Minute,
		CompletedTTL:  time.Hour,
	})

	ledger := newMemoryLedger()
	alice, _ := ledgerdomain.NewAccount(ledgerdomain.AccountID("wallet-alice"), ledgerdomain.AccountTypeLiability, "Alice")
	bob, _ := ledgerdomain.NewAccount(ledgerdomain.AccountID("wallet-bob"), ledgerdomain.AccountTypeLiability, "Bob")
	ledger.accounts[alice.ID()] = alice
	ledger.accounts[bob.ID()] = bob
	require.NoError(t, ledger.seed(alice, shareddomain.MustSAR(1000, 0)))

	coordinator := &memoryCoordinator{ledger: ledger}
	fraud := fraudapp.NewChecker(fraudredis.NewVelocityStore(client), fraudapp.Config{MaxHourlyTransfers: 100})
	svc := transferapp.NewService(ledger, coordinator, guard, fraud, complianceapp.NewChecker(), &memorySagaStore{}, nil)
	return svc, ledger
}

func sampleCommand(key string) transferdomain.Command {
	return transferdomain.Command{
		IdempotencyKey:  key,
		FromAccountID:   ledgerdomain.AccountID("wallet-alice"),
		ToAccountID:     ledgerdomain.AccountID("wallet-bob"),
		Amount:          shareddomain.MustSAR(100, 0),
		BeneficiaryIBAN: "SA0380000000608010167519",
		Description:     "p2p transfer",
	}
}

func TestTransferHappyPath(t *testing.T) {
	svc, ledger := setupTransferService(t)
	ctx := context.Background()

	result, err := svc.Execute(ctx, sampleCommand("transfer-key-001"))
	require.NoError(t, err)
	require.False(t, result.Replayed)
	require.NotEmpty(t, result.TransactionID)

	alice, _ := ledger.GetAccount(ctx, ledgerdomain.AccountID("wallet-alice"))
	aliceBal := balanceOf(alice, ledger.entries[alice.ID()])
	require.Equal(t, "900.00", aliceBal.String())

	bob, _ := ledger.GetAccount(ctx, ledgerdomain.AccountID("wallet-bob"))
	bobBal := balanceOf(bob, ledger.entries[bob.ID()])
	require.Equal(t, "100.00", bobBal.String())
	require.Len(t, ledger.events, 1)
	require.Equal(t, outboxdomain.EventTransferPosted, ledger.events[0].EventType())
}

func TestTransferIdempotencyReplay(t *testing.T) {
	svc, ledger := setupTransferService(t)
	ctx := context.Background()
	cmd := sampleCommand("transfer-key-002")

	first, err := svc.Execute(ctx, cmd)
	require.NoError(t, err)

	second, err := svc.Execute(ctx, cmd)
	require.NoError(t, err)
	require.True(t, second.Replayed)
	require.Equal(t, first.TransactionID, second.TransactionID)
	require.Len(t, ledger.posted, 1)
}

func TestTransferRejectsInvalidIBAN(t *testing.T) {
	svc, _ := setupTransferService(t)
	ctx := context.Background()

	cmd := sampleCommand("transfer-key-003")
	cmd.BeneficiaryIBAN = "SA9980000000608010167519"

	_, err := svc.Execute(ctx, cmd)
	require.Error(t, err)
	require.True(t, shareddomain.IsDomainCode(err, shareddomain.ErrCodeInvalidIBAN))
}

func TestTransferRejectsInsufficientFunds(t *testing.T) {
	svc, _ := setupTransferService(t)
	ctx := context.Background()

	cmd := sampleCommand("transfer-key-004")
	cmd.Amount = shareddomain.MustSAR(2000, 0)

	_, err := svc.Execute(ctx, cmd)
	require.Error(t, err)
	require.True(t, shareddomain.IsDomainCode(err, shareddomain.ErrCodeInsufficientFunds))
}

func TestTransferRejectsSanctionedIBAN(t *testing.T) {
	svc, _ := setupTransferService(t)
	ctx := context.Background()

	cmd := sampleCommand("transfer-key-005")
	cmd.BeneficiaryIBAN = "SA4420000001234567891234"

	_, err := svc.Execute(ctx, cmd)
	require.Error(t, err)
	require.True(t, shareddomain.IsDomainCode(err, shareddomain.ErrCodeForbidden))
}

func TestTransferRejectsFraudVelocity(t *testing.T) {
	mr := miniredis.RunT(t)
	client := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	store := idempotencyredis.NewStore(client)
	guard := idempotencyapp.NewGuard(store, idempotencyapp.Config{ProcessingTTL: time.Minute, CompletedTTL: time.Hour})

	ledger := newMemoryLedger()
	alice, _ := ledgerdomain.NewAccount(ledgerdomain.AccountID("wallet-alice"), ledgerdomain.AccountTypeLiability, "Alice")
	bob, _ := ledgerdomain.NewAccount(ledgerdomain.AccountID("wallet-bob"), ledgerdomain.AccountTypeLiability, "Bob")
	ledger.accounts[alice.ID()] = alice
	ledger.accounts[bob.ID()] = bob
	require.NoError(t, ledger.seed(alice, shareddomain.MustSAR(1000, 0)))

	fraud := fraudapp.NewChecker(fraudredis.NewVelocityStore(client), fraudapp.Config{MaxHourlyTransfers: 1})
	svc := transferapp.NewService(ledger, &memoryCoordinator{ledger: ledger}, guard, fraud, complianceapp.NewChecker(), &memorySagaStore{}, nil)

	ctx := context.Background()
	_, err := svc.Execute(ctx, sampleCommand("velocity-1"))
	require.NoError(t, err)
	_, err = svc.Execute(ctx, sampleCommand("velocity-2"))
	require.Error(t, err)
	require.True(t, shareddomain.IsDomainCode(err, shareddomain.ErrCodeForbidden))
}

// Ensure memoryLedger implements transfer ports at compile time.
var (
	_ ports.AccountReader     = (*memoryLedger)(nil)
	_ ports.TransactionWriter = (*memoryLedger)(nil)
	_ ports.Idempotency       = (*idempotencyapp.Guard)(nil)
)
