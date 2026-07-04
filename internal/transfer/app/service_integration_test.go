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
	"github.com/mohsinazam/banking/internal/outbox/adapters/logpublisher"
	outboxpostgres "github.com/mohsinazam/banking/internal/outbox/adapters/postgres"
	outboxapp "github.com/mohsinazam/banking/internal/outbox/app"
	outboxdomain "github.com/mohsinazam/banking/internal/outbox/domain"
	uowpostgres "github.com/mohsinazam/banking/internal/platform/uow/postgres"
	shareddomain "github.com/mohsinazam/banking/internal/shared/domain"
	transferapp "github.com/mohsinazam/banking/internal/transfer/app"
	transferdomain "github.com/mohsinazam/banking/internal/transfer/domain"
	transferpostgres "github.com/mohsinazam/banking/internal/transfer/adapters/postgres"
	"github.com/mohsinazam/banking/tests/integration"
)

func TestTransferIntegrationHappyPath(t *testing.T) {
	ledgerRepo, pool, cleanup := integration.LedgerRepo(t)
	defer cleanup()

	ctx := context.Background()
	accounts := integration.AccountReader{Repo: ledgerRepo}
	outboxRepo := outboxpostgres.NewRepository(pool)
	sagaRepo := transferpostgres.NewSagaRepository(pool)
	coordinator := uowpostgres.NewCoordinator(pool, ledgerRepo, outboxRepo, sagaRepo, testLogger(t))

	alice, _ := ledgerdomain.NewAccount(ledgerdomain.AccountID("wallet-alice"), ledgerdomain.AccountTypeLiability, "Alice")
	bob, _ := ledgerdomain.NewAccount(ledgerdomain.AccountID("wallet-bob"), ledgerdomain.AccountTypeLiability, "Bob")
	suspense, _ := ledgerdomain.NewAccount(ledgerdomain.AccountID("suspense-funding"), ledgerdomain.AccountTypeAsset, "Suspense")

	require.NoError(t, ledgerRepo.CreateAccount(ctx, alice))
	require.NoError(t, ledgerRepo.CreateAccount(ctx, bob))
	require.NoError(t, ledgerRepo.CreateAccount(ctx, suspense))

	fundAmount := shareddomain.MustSAR(500, 0)
	funding, err := ledgerdomain.NewTransaction(
		ledgerdomain.TransactionID("tx-fund-step6"),
		"fund",
		[]ledgerdomain.Entry{
			integration.MustEntry(t, "fund-c", alice.ID().String(), ledgerdomain.SideCredit, fundAmount),
			integration.MustEntry(t, "fund-d", suspense.ID().String(), ledgerdomain.SideDebit, fundAmount),
		},
		time.Now().UTC(),
	)
	require.NoError(t, err)
	require.NoError(t, ledgerRepo.PostTransaction(ctx, funding))

	mr := miniredis.RunT(t)
	redisClient := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	guard := idempotencyapp.NewGuard(idempotencyredis.NewStore(redisClient), idempotencyapp.DefaultConfig())
	fraud := fraudapp.NewChecker(fraudredis.NewVelocityStore(redisClient), fraudapp.Config{MaxHourlyTransfers: 100})

	svc := transferapp.NewService(
		accounts,
		coordinator,
		guard,
		fraud,
		complianceapp.NewChecker(),
		sagaRepo,
		testLogger(t),
	)

	result, err := svc.Execute(ctx, transferdomain.Command{
		IdempotencyKey:  "integration-transfer-001",
		FromAccountID:   alice.ID(),
		ToAccountID:     bob.ID(),
		Amount:          shareddomain.MustSAR(125, 50),
		BeneficiaryIBAN: "SA03 8000 0000 6080 1016 7519",
		Description:     "integration test transfer",
	})
	require.NoError(t, err)
	require.False(t, result.Replayed)

	aliceBal, err := ledgerRepo.GetBalance(ctx, alice)
	require.NoError(t, err)
	require.Equal(t, "374.50", aliceBal.String())

	bobBal, err := ledgerRepo.GetBalance(ctx, bob)
	require.NoError(t, err)
	require.Equal(t, "125.50", bobBal.String())

	pending, err := outboxRepo.FetchPending(ctx, 10)
	require.NoError(t, err)
	require.Len(t, pending, 1)
	require.Equal(t, outboxdomain.EventTransferPosted, pending[0].EventType())

	relay := outboxapp.NewRelay(outboxRepo, logpublisher.NewPublisher(testLogger(t)), testLogger(t), outboxapp.RelayConfig{BatchSize: 10, Interval: time.Second})
	processed, err := relay.ProcessOnce(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, processed)

	pending, err = outboxRepo.FetchPending(ctx, 10)
	require.NoError(t, err)
	require.Empty(t, pending)

	replayed, err := svc.Execute(ctx, transferdomain.Command{
		IdempotencyKey:  "integration-transfer-001",
		FromAccountID:   alice.ID(),
		ToAccountID:     bob.ID(),
		Amount:          shareddomain.MustSAR(125, 50),
		BeneficiaryIBAN: "SA03 8000 0000 6080 1016 7519",
	})
	require.NoError(t, err)
	require.True(t, replayed.Replayed)
	require.Equal(t, result.TransactionID, replayed.TransactionID)

	require.NoError(t, ledgerRepo.VerifyGlobalLedgerBalanced(ctx))
}

func testLogger(t *testing.T) *slog.Logger {
	t.Helper()
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
