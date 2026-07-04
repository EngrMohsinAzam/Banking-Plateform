//go:build integration

package postgres_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	ledgerpostgres "github.com/mohsinazam/banking/internal/ledger/adapters/postgres"
	"github.com/mohsinazam/banking/internal/ledger/domain"
	platformpostgres "github.com/mohsinazam/banking/internal/platform/postgres"
	shareddomain "github.com/mohsinazam/banking/internal/shared/domain"
)

const defaultLocalDSN = "postgres://banking:banking@localhost:5432/banking?sslmode=disable"

func setupRepository(t *testing.T) (*ledgerpostgres.Repository, func()) {
	t.Helper()
	ctx := context.Background()

	dsn, containerCleanup := resolvePostgresDSN(t, ctx)
	require.NoError(t, platformpostgres.RunMigrations(dsn))

	pool, err := platformpostgres.NewPool(ctx, dsn)
	require.NoError(t, err)

	repo := ledgerpostgres.NewRepository(pool)
	cleanup := func() {
		pool.Close()
		containerCleanup()
	}
	return repo, cleanup
}

func resolvePostgresDSN(t *testing.T, ctx context.Context) (string, func()) {
	t.Helper()

	if dsn := os.Getenv("TEST_POSTGRES_DSN"); dsn != "" {
		requirePostgres(t, ctx, dsn)
		return dsn, func() {}
	}

	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("banking"),
		postgres.WithUsername("banking"),
		postgres.WithPassword("banking"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err == nil {
		dsn, connErr := container.ConnectionString(ctx, "sslmode=disable")
		require.NoError(t, connErr)
		return dsn, func() {
			_ = testcontainers.TerminateContainer(container)
		}
	}

	t.Logf("testcontainers unavailable (%v); falling back to docker-compose Postgres", err)
	requirePostgres(t, ctx, defaultLocalDSN)
	return defaultLocalDSN, func() {}
}

func requirePostgres(t *testing.T, ctx context.Context, dsn string) {
	t.Helper()
	pool, err := platformpostgres.NewPool(ctx, dsn)
	if err != nil {
		t.Skipf("postgres not reachable at %s: %v (run `make up` or set TEST_POSTGRES_DSN)", dsn, err)
	}
	pool.Close()
}

func TestPostTransactionAndDerivedBalance(t *testing.T) {
	repo, cleanup := setupRepository(t)
	defer cleanup()

	ctx := context.Background()

	alice, err := domain.NewAccount(domain.AccountID("wallet-alice"), domain.AccountTypeLiability, "Alice")
	require.NoError(t, err)
	bob, err := domain.NewAccount(domain.AccountID("wallet-bob"), domain.AccountTypeLiability, "Bob")
	require.NoError(t, err)

	require.NoError(t, repo.CreateAccount(ctx, alice))
	require.NoError(t, repo.CreateAccount(ctx, bob))

	amount, err := shareddomain.SAR(100, 0)
	require.NoError(t, err)

	funding, err := domain.NewTransaction(
		domain.TransactionID("tx-fund-alice"),
		"initial funding",
		[]domain.Entry{
			mustEntry(t, "e1", "wallet-alice", domain.SideCredit, amount),
			mustEntry(t, "e2", "suspense-funding", domain.SideDebit, amount),
		},
		time.Now().UTC(),
	)
	require.NoError(t, err)

	suspense, err := domain.NewAccount(domain.AccountID("suspense-funding"), domain.AccountTypeAsset, "Funding Suspense")
	require.NoError(t, err)
	require.NoError(t, repo.CreateAccount(ctx, suspense))
	require.NoError(t, repo.PostTransaction(ctx, funding))

	transferAmount, err := shareddomain.SAR(40, 0)
	require.NoError(t, err)
	transfer, err := domain.NewTransaction(
		domain.TransactionID("tx-transfer"),
		"p2p transfer",
		[]domain.Entry{
			mustEntry(t, "e3", "wallet-alice", domain.SideDebit, transferAmount),
			mustEntry(t, "e4", "wallet-bob", domain.SideCredit, transferAmount),
		},
		time.Now().UTC(),
	)
	require.NoError(t, err)
	require.NoError(t, repo.PostTransaction(ctx, transfer))

	aliceBal, err := repo.GetBalance(ctx, alice)
	require.NoError(t, err)
	require.Equal(t, "60.00", aliceBal.String())

	bobBal, err := repo.GetBalance(ctx, bob)
	require.NoError(t, err)
	require.Equal(t, "40.00", bobBal.String())
}

func TestPostTransactionIsAtomicOnFailure(t *testing.T) {
	repo, cleanup := setupRepository(t)
	defer cleanup()

	ctx := context.Background()

	alice, err := domain.NewAccount(domain.AccountID("wallet-alice"), domain.AccountTypeLiability, "Alice")
	require.NoError(t, err)
	require.NoError(t, repo.CreateAccount(ctx, alice))

	amount, err := shareddomain.SAR(10, 0)
	require.NoError(t, err)

	badTx, err := domain.NewTransaction(
		domain.TransactionID("tx-bad-fk"),
		"missing counterparty account",
		[]domain.Entry{
			mustEntry(t, "e1", "wallet-alice", domain.SideDebit, amount),
			mustEntry(t, "e2", "wallet-missing", domain.SideCredit, amount),
		},
		time.Now().UTC(),
	)
	require.NoError(t, err)

	err = repo.PostTransaction(ctx, badTx)
	require.Error(t, err)
	require.True(t, shareddomain.IsDomainCode(err, shareddomain.ErrCodeValidation))

	entries, err := repo.ListEntriesByAccount(ctx, alice.ID())
	require.NoError(t, err)
	require.Empty(t, entries)
}

func TestPostTransactionDuplicateIsConflict(t *testing.T) {
	repo, cleanup := setupRepository(t)
	defer cleanup()

	ctx := context.Background()

	alice, _ := domain.NewAccount(domain.AccountID("wallet-alice"), domain.AccountTypeLiability, "Alice")
	bob, _ := domain.NewAccount(domain.AccountID("wallet-bob"), domain.AccountTypeLiability, "Bob")
	require.NoError(t, repo.CreateAccount(ctx, alice))
	require.NoError(t, repo.CreateAccount(ctx, bob))

	amount, _ := shareddomain.SAR(5, 0)
	tx, err := domain.NewTransaction(
		domain.TransactionID("tx-dup"),
		"dup test",
		[]domain.Entry{
			mustEntry(t, "e1", "wallet-alice", domain.SideDebit, amount),
			mustEntry(t, "e2", "wallet-bob", domain.SideCredit, amount),
		},
		time.Now().UTC(),
	)
	require.NoError(t, err)

	require.NoError(t, repo.PostTransaction(ctx, tx))

	err = repo.PostTransaction(ctx, tx)
	require.Error(t, err)
	require.True(t, shareddomain.IsDomainCode(err, shareddomain.ErrCodeConflict))
}

func mustEntry(
	t *testing.T,
	id string,
	accountID string,
	side domain.EntrySide,
	amount shareddomain.Money,
) domain.Entry {
	t.Helper()
	entry, err := domain.NewEntry(domain.EntryID(id), domain.AccountID(accountID), side, amount)
	require.NoError(t, err)
	return entry
}
