package postgres

import (
	"context"
	"fmt"
	"sort"

	"github.com/jackc/pgx/v5"
	shareddomain "github.com/mohsinazam/banking/internal/shared/domain"
	"github.com/mohsinazam/banking/internal/ledger/domain"
)

// lockAccounts acquires row-level locks on all accounts touched by a transaction.
// IDs are locked in sorted order to prevent deadlocks when two transfers cross accounts.
func lockAccounts(ctx context.Context, dbTx pgx.Tx, accountIDs []string) ([]domain.Account, error) {
	sort.Strings(accountIDs)

	rows, err := dbTx.Query(ctx, `
		SELECT id, account_type, name
		FROM ledger_accounts
		WHERE id = ANY($1)
		ORDER BY id
		FOR UPDATE
	`, accountIDs)
	if err != nil {
		return nil, fmt.Errorf("lock accounts: %w", err)
	}
	defer rows.Close()

	accounts := make([]domain.Account, 0, len(accountIDs))
	found := make(map[string]struct{}, len(accountIDs))

	for rows.Next() {
		var id, accountType, name string
		if err := rows.Scan(&id, &accountType, &name); err != nil {
			return nil, fmt.Errorf("scan account: %w", err)
		}
		account, err := domain.NewAccount(domain.AccountID(id), domain.AccountType(accountType), name)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
		found[id] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate accounts: %w", err)
	}

	if len(found) != len(accountIDs) {
		return nil, shareddomain.NewDomainError(shareddomain.ErrCodeValidation, "account does not exist")
	}

	return accounts, nil
}

func uniqueAccountIDs(tx domain.Transaction) []string {
	seen := make(map[string]struct{})
	for _, entry := range tx.Entries() {
		seen[entry.AccountID().String()] = struct{}{}
	}
	ids := make([]string, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// validateSufficientFunds checks projected balances under row locks before posting.
func validateSufficientFunds(
	ctx context.Context,
	dbTx pgx.Tx,
	accounts []domain.Account,
	tx domain.Transaction,
) error {
	txEntries := tx.Entries()

	for _, account := range accounts {
		existing, err := listEntriesByAccountTx(ctx, dbTx, account.ID())
		if err != nil {
			return err
		}

		current, err := domain.BalanceFromEntries(account, existing)
		if err != nil {
			return err
		}

		projected, err := domain.ApplyEntries(account, current, txEntries)
		if err != nil {
			return err
		}

		if projected.IsNegative() {
			return shareddomain.NewDomainError(
				shareddomain.ErrCodeInsufficientFunds,
				fmt.Sprintf("account %s has insufficient funds", account.ID()),
			)
		}
	}
	return nil
}

func listEntriesByAccountTx(ctx context.Context, dbTx pgx.Tx, accountID domain.AccountID) ([]domain.Entry, error) {
	rows, err := dbTx.Query(ctx, `
		SELECT id, account_id, side, amount_halalas, currency
		FROM ledger_entries
		WHERE account_id = $1
		ORDER BY created_at ASC, id ASC
	`, accountID.String())
	if err != nil {
		return nil, fmt.Errorf("list entries in tx: %w", err)
	}
	defer rows.Close()

	var entries []domain.Entry
	for rows.Next() {
		var entryID, acctID, side, currency string
		var halalas int64
		if err := rows.Scan(&entryID, &acctID, &side, &halalas, &currency); err != nil {
			return nil, fmt.Errorf("scan entry: %w", err)
		}

		amount, err := shareddomain.HalalasFromMinorUnits(halalas, currency)
		if err != nil {
			return nil, err
		}

		entry, err := domain.NewEntry(
			domain.EntryID(entryID),
			domain.AccountID(acctID),
			domain.EntrySide(side),
			amount,
		)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate entries in tx: %w", err)
	}
	return entries, nil
}
