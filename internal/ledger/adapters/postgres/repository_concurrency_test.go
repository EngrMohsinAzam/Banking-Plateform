//go:build integration

package postgres_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	ledgerpostgres "github.com/mohsinazam/banking/internal/ledger/adapters/postgres"
	"github.com/mohsinazam/banking/internal/ledger/domain"
	shareddomain "github.com/mohsinazam/banking/internal/shared/domain"
)

func TestInsufficientFundsRejectedUnderLock(t *testing.T) {
	repo, cleanup := setupRepository(t)
	defer cleanup()
	ctx := context.Background()

	alice, bob, suspense := createWalletAccounts(t)
	require.NoError(t, repo.CreateAccount(ctx, alice))
	require.NoError(t, repo.CreateAccount(ctx, bob))
	require.NoError(t, repo.CreateAccount(ctx, suspense))

	fundAccount(t, repo, ctx, alice, suspense, shareddomain.MustSAR(50, 0), "tx-fund")

	tooMuch := shareddomain.MustSAR(51, 0)
	transfer, err := domain.NewTransaction(
		domain.TransactionID("tx-overdraft"),
		"overdraft attempt",
		[]domain.Entry{
			mustEntry(t, "e-over-1", alice.ID().String(), domain.SideDebit, tooMuch),
			mustEntry(t, "e-over-2", bob.ID().String(), domain.SideCredit, tooMuch),
		},
		time.Now().UTC(),
	)
	require.NoError(t, err)

	err = repo.PostTransaction(ctx, transfer)
	require.Error(t, err)
	require.True(t, shareddomain.IsDomainCode(err, shareddomain.ErrCodeInsufficientFunds))

	aliceBal, err := repo.GetBalance(ctx, alice)
	require.NoError(t, err)
	require.Equal(t, "50.00", aliceBal.String())
}

func TestConcurrentTransfersBooksRemainBalanced(t *testing.T) {
	repo, cleanup := setupRepository(t)
	defer cleanup()
	ctx := context.Background()

	alice, bob, suspense := createWalletAccounts(t)
	require.NoError(t, repo.CreateAccount(ctx, alice))
	require.NoError(t, repo.CreateAccount(ctx, bob))
	require.NoError(t, repo.CreateAccount(ctx, suspense))

	const (
		workers       = 100
		transferCents = int64(10000) // 100.00 SAR per transfer
	)
	funding := shareddomain.MustSAR(workers*transferCents/100, 0) // 10000.00 SAR
	fundAccount(t, repo, ctx, alice, suspense, funding, "tx-fund-concurrent")

	var wg sync.WaitGroup
	errs := make(chan error, workers)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			amount, err := shareddomain.HalalasFromMinorUnits(transferCents, shareddomain.CurrencySAR)
			if err != nil {
				errs <- err
				return
			}

			tx, err := domain.NewTransaction(
				domain.TransactionID(fmt.Sprintf("tx-concurrent-%03d", i)),
				"parallel transfer",
				[]domain.Entry{
					mustEntry(t, fmt.Sprintf("e-d-%03d", i), alice.ID().String(), domain.SideDebit, amount),
					mustEntry(t, fmt.Sprintf("e-c-%03d", i), bob.ID().String(), domain.SideCredit, amount),
				},
				time.Now().UTC(),
			)
			if err != nil {
				errs <- err
				return
			}

			if err := repo.PostTransaction(ctx, tx); err != nil {
				errs <- err
			}
		}(i)
	}

	wg.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}

	aliceBal, err := repo.GetBalance(ctx, alice)
	require.NoError(t, err)
	require.True(t, aliceBal.IsZero(), "alice balance = %s", aliceBal.String())

	bobBal, err := repo.GetBalance(ctx, bob)
	require.NoError(t, err)
	require.Equal(t, "10000.00", bobBal.String())

	require.NoError(t, repo.VerifyGlobalLedgerBalanced(ctx))
}

func TestConcurrentTransfersRejectExcessOverdraft(t *testing.T) {
	repo, cleanup := setupRepository(t)
	defer cleanup()
	ctx := context.Background()

	alice, bob, suspense := createWalletAccounts(t)
	require.NoError(t, repo.CreateAccount(ctx, alice))
	require.NoError(t, repo.CreateAccount(ctx, bob))
	require.NoError(t, repo.CreateAccount(ctx, suspense))

	fundAccount(t, repo, ctx, alice, suspense, shareddomain.MustSAR(1000, 0), "tx-fund-contention")

	const (
		workers         = 50
		transferAmount  = int64(10000) // 100.00 SAR
		expectedSuccess = 10
	)

	var success atomic.Int32
	var rejected atomic.Int32
	var unexpected atomic.Int32

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			amount, err := shareddomain.HalalasFromMinorUnits(transferAmount, shareddomain.CurrencySAR)
			if err != nil {
				unexpected.Add(1)
				return
			}

			tx, err := domain.NewTransaction(
				domain.TransactionID(fmt.Sprintf("tx-contention-%03d", i)),
				"contention transfer",
				[]domain.Entry{
					mustEntry(t, fmt.Sprintf("e-cd-%03d", i), alice.ID().String(), domain.SideDebit, amount),
					mustEntry(t, fmt.Sprintf("e-cc-%03d", i), bob.ID().String(), domain.SideCredit, amount),
				},
				time.Now().UTC(),
			)
			if err != nil {
				unexpected.Add(1)
				return
			}

			err = repo.PostTransaction(ctx, tx)
			switch {
			case err == nil:
				success.Add(1)
			case shareddomain.IsDomainCode(err, shareddomain.ErrCodeInsufficientFunds):
				rejected.Add(1)
			default:
				unexpected.Add(1)
			}
		}(i)
	}
	wg.Wait()

	require.Equal(t, int32(0), unexpected.Load())
	require.Equal(t, int32(expectedSuccess), success.Load())
	require.Equal(t, int32(workers-expectedSuccess), rejected.Load())

	aliceBal, err := repo.GetBalance(ctx, alice)
	require.NoError(t, err)
	require.True(t, aliceBal.IsZero(), "alice balance = %s", aliceBal.String())

	bobBal, err := repo.GetBalance(ctx, bob)
	require.NoError(t, err)
	require.Equal(t, "1000.00", bobBal.String())

	require.NoError(t, repo.VerifyGlobalLedgerBalanced(ctx))
}

func createWalletAccounts(t *testing.T) (alice, bob, suspense domain.Account) {
	t.Helper()
	alice, err := domain.NewAccount(domain.AccountID("wallet-alice"), domain.AccountTypeLiability, "Alice")
	require.NoError(t, err)
	bob, err = domain.NewAccount(domain.AccountID("wallet-bob"), domain.AccountTypeLiability, "Bob")
	require.NoError(t, err)
	suspense, err = domain.NewAccount(domain.AccountID("suspense-funding"), domain.AccountTypeAsset, "Funding Suspense")
	require.NoError(t, err)
	return alice, bob, suspense
}

func fundAccount(
	t *testing.T,
	repo *ledgerpostgres.Repository,
	ctx context.Context,
	wallet, suspense domain.Account,
	amount shareddomain.Money,
	txID string,
) {
	t.Helper()

	funding, err := domain.NewTransaction(
		domain.TransactionID(txID),
		"fund wallet",
		[]domain.Entry{
			mustEntry(t, txID+"-credit", wallet.ID().String(), domain.SideCredit, amount),
			mustEntry(t, txID+"-debit", suspense.ID().String(), domain.SideDebit, amount),
		},
		time.Now().UTC(),
	)
	require.NoError(t, err)
	require.NoError(t, repo.PostTransaction(ctx, funding))
}
