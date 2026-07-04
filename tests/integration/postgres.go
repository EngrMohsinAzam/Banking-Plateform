//go:build integration

package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	ledgerdomain "github.com/mohsinazam/banking/internal/ledger/domain"
	ledgerpostgres "github.com/mohsinazam/banking/internal/ledger/adapters/postgres"
	platformpostgres "github.com/mohsinazam/banking/internal/platform/postgres"
	shareddomain "github.com/mohsinazam/banking/internal/shared/domain"
)

const defaultLocalDSN = "postgres://banking:banking@localhost:5432/banking?sslmode=disable"

// LedgerRepo returns a migrated postgres ledger repository for integration tests.
func LedgerRepo(t *testing.T) (*ledgerpostgres.Repository, *pgxpool.Pool, func()) {
	t.Helper()
	dsn, cleanup := resolvePostgresDSN(t, context.Background())
	require.NoError(t, platformpostgres.RunMigrations(dsn))
	pool, err := platformpostgres.NewPool(context.Background(), dsn)
	require.NoError(t, err)
	repo := ledgerpostgres.NewRepository(pool)
	return repo, pool, func() {
		pool.Close()
		cleanup()
	}
}

// AccountReader adapts a ledger repository for transfer account reads.
type AccountReader struct {
	Repo *ledgerpostgres.Repository
}

func (a AccountReader) GetAccount(ctx context.Context, id ledgerdomain.AccountID) (ledgerdomain.Account, error) {
	return a.Repo.GetAccount(ctx, id)
}

// MustEntry builds a ledger entry for tests.
func MustEntry(t *testing.T, id, accountID string, side ledgerdomain.EntrySide, amount shareddomain.Money) ledgerdomain.Entry {
	t.Helper()
	entry, err := ledgerdomain.NewEntry(ledgerdomain.EntryID(id), ledgerdomain.AccountID(accountID), side, amount)
	require.NoError(t, err)
	return entry
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
		return dsn, func() { _ = testcontainers.TerminateContainer(container) }
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
