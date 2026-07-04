//go:build integration

package app_test

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"

	complianceapp "github.com/mohsinazam/banking/internal/compliance/app"
	fraudapp "github.com/mohsinazam/banking/internal/fraud/app"
	fraudredis "github.com/mohsinazam/banking/internal/fraud/adapters/redis"
	idempotencyapp "github.com/mohsinazam/banking/internal/idempotency/app"
	idempotencyredis "github.com/mohsinazam/banking/internal/idempotency/adapters/redis"
	ledgerdomain "github.com/mohsinazam/banking/internal/ledger/domain"
	outboxpostgres "github.com/mohsinazam/banking/internal/outbox/adapters/postgres"
	uowpostgres "github.com/mohsinazam/banking/internal/platform/uow/postgres"
	settlementapp "github.com/mohsinazam/banking/internal/settlement/app"
	shareddomain "github.com/mohsinazam/banking/internal/shared/domain"
	transferapp "github.com/mohsinazam/banking/internal/transfer/app"
	transferdomain "github.com/mohsinazam/banking/internal/transfer/domain"
	transferpostgres "github.com/mohsinazam/banking/internal/transfer/adapters/postgres"
	"github.com/mohsinazam/banking/tests/integration"
)

func TestSettlementProcessorCompletesSuccessfully(t *testing.T) {
	ledgerRepo, pool, cleanup := integration.LedgerRepo(t)
	defer cleanup()

	ctx := context.Background()
	accounts := integration.AccountReader{Repo: ledgerRepo}
	sagaRepo := transferpostgres.NewSagaRepository(pool)
	coordinator := uowpostgres.NewCoordinator(pool, ledgerRepo, outboxpostgres.NewRepository(pool), sagaRepo, testLogger(t))

	alice, bob, _ := fundWallets(t, ctx, ledgerRepo)
	svc := newTransferService(t, accounts, sagaRepo, coordinator)

	result, err := svc.Execute(ctx, transferdomain.Command{
		IdempotencyKey:  "settlement-success-001",
		FromAccountID:   alice.ID(),
		ToAccountID:     bob.ID(),
		Amount:          shareddomain.MustSAR(25, 0),
		BeneficiaryIBAN: "SA0380000000608010167519",
	})
	require.NoError(t, err)

	processor := settlementapp.NewSettlementProcessor(
		sagaRepo, coordinator,
		settlementapp.NewSarieMock(settlementapp.SarieConfig{FailRate: 0, MinDelay: time.Millisecond, MaxDelay: time.Millisecond}),
		accounts, testLogger(t),
	)
	processed, err := processor.ProcessOnce(ctx, 10)
	require.NoError(t, err)
	require.Equal(t, 1, processed)

	status, err := svc.GetStatusByTransactionID(ctx, result.TransactionID.String())
	require.NoError(t, err)
	require.Equal(t, transferdomain.SagaStateCompleted, status.SagaState)
	require.Equal(t, transferdomain.SettlementSettled, status.SettlementStatus)
}

func TestSettlementProcessorCompensatesOnFailure(t *testing.T) {
	ledgerRepo, pool, cleanup := integration.LedgerRepo(t)
	defer cleanup()

	ctx := context.Background()
	accounts := integration.AccountReader{Repo: ledgerRepo}
	sagaRepo := transferpostgres.NewSagaRepository(pool)
	coordinator := uowpostgres.NewCoordinator(pool, ledgerRepo, outboxpostgres.NewRepository(pool), sagaRepo, testLogger(t))

	alice, bob, _ := fundWallets(t, ctx, ledgerRepo)
	svc := newTransferService(t, accounts, sagaRepo, coordinator)

	amount := shareddomain.MustSAR(40, 0)
	result, err := svc.Execute(ctx, transferdomain.Command{
		IdempotencyKey:  "settlement-fail-001",
		FromAccountID:   alice.ID(),
		ToAccountID:     bob.ID(),
		Amount:          amount,
		BeneficiaryIBAN: "SA0380000000608010167519",
	})
	require.NoError(t, err)

	aliceBefore, _ := ledgerRepo.GetBalance(ctx, alice)
	bobBefore, _ := ledgerRepo.GetBalance(ctx, bob)

	processor := settlementapp.NewSettlementProcessor(
		sagaRepo, coordinator,
		settlementapp.NewSarieMock(settlementapp.SarieConfig{FailRate: 1, MinDelay: time.Millisecond, MaxDelay: time.Millisecond}),
		accounts, testLogger(t),
	)
	processed, err := processor.ProcessOnce(ctx, 10)
	require.NoError(t, err)
	require.Equal(t, 1, processed)

	status, err := svc.GetStatusByTransactionID(ctx, result.TransactionID.String())
	require.NoError(t, err)
	require.Equal(t, transferdomain.SagaStateCompensated, status.SagaState)

	aliceAfter, _ := ledgerRepo.GetBalance(ctx, alice)
	bobAfter, _ := ledgerRepo.GetBalance(ctx, bob)
	require.Equal(t, aliceBefore.String(), aliceAfter.String())
	require.Equal(t, bobBefore.String(), bobAfter.String())
}

func newTransferService(t *testing.T, accounts integration.AccountReader, sagaRepo *transferpostgres.SagaRepository, coordinator *uowpostgres.Coordinator) *transferapp.Service {
	t.Helper()
	mr := miniredis.RunT(t)
	redisClient := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	return transferapp.NewService(
		accounts,
		coordinator,
		idempotencyapp.NewGuard(idempotencyredis.NewStore(redisClient), idempotencyapp.DefaultConfig()),
		fraudapp.NewChecker(fraudredis.NewVelocityStore(redisClient), fraudapp.Config{MaxHourlyTransfers: 100}),
		complianceapp.NewChecker(),
		sagaRepo,
		testLogger(t),
	)
}

func fundWallets(t *testing.T, ctx context.Context, ledgerRepo interface {
	CreateAccount(context.Context, ledgerdomain.Account) error
	PostTransaction(context.Context, ledgerdomain.Transaction) error
}) (ledgerdomain.Account, ledgerdomain.Account, ledgerdomain.Account) {
	t.Helper()
	alice, _ := ledgerdomain.NewAccount(ledgerdomain.AccountID("wallet-alice"), ledgerdomain.AccountTypeLiability, "Alice")
	bob, _ := ledgerdomain.NewAccount(ledgerdomain.AccountID("wallet-bob"), ledgerdomain.AccountTypeLiability, "Bob")
	suspense, _ := ledgerdomain.NewAccount(ledgerdomain.AccountID("suspense-funding"), ledgerdomain.AccountTypeAsset, "Suspense")
	require.NoError(t, ledgerRepo.CreateAccount(ctx, alice))
	require.NoError(t, ledgerRepo.CreateAccount(ctx, bob))
	require.NoError(t, ledgerRepo.CreateAccount(ctx, suspense))

	fundAmount := shareddomain.MustSAR(100_000, 0)
	funding, err := ledgerdomain.NewTransaction(
		ledgerdomain.TransactionID("tx-fund-settlement-test"),
		"fund",
		[]ledgerdomain.Entry{
			integration.MustEntry(t, "fund-c", alice.ID().String(), ledgerdomain.SideCredit, fundAmount),
			integration.MustEntry(t, "fund-d", suspense.ID().String(), ledgerdomain.SideDebit, fundAmount),
		},
		time.Now().UTC(),
	)
	require.NoError(t, err)
	require.NoError(t, ledgerRepo.PostTransaction(ctx, funding))
	return alice, bob, suspense
}

func testLogger(t *testing.T) *slog.Logger {
	t.Helper()
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
