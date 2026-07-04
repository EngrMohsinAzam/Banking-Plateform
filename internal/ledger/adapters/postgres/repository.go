package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	shareddomain "github.com/mohsinazam/banking/internal/shared/domain"
	"github.com/mohsinazam/banking/internal/ledger/domain"
)

// Repository is the PostgreSQL adapter for the append-only ledger.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository constructs a ledger repository backed by Postgres.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// CreateAccount inserts a new ledger account.
func (r *Repository) CreateAccount(ctx context.Context, account domain.Account) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO ledger_accounts (id, account_type, name)
		VALUES ($1, $2, $3)
	`, account.ID().String(), string(account.Type()), account.Name())
	if err != nil {
		if isUniqueViolation(err) {
			return shareddomain.NewDomainError(shareddomain.ErrCodeConflict, "account already exists")
		}
		return fmt.Errorf("insert account: %w", err)
	}
	return nil
}

// GetAccount loads account metadata by ID.
func (r *Repository) GetAccount(ctx context.Context, id domain.AccountID) (domain.Account, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT account_type, name
		FROM ledger_accounts
		WHERE id = $1
	`, id.String())

	var accountType string
	var name string
	if err := row.Scan(&accountType, &name); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Account{}, shareddomain.NewDomainError(shareddomain.ErrCodeNotFound, "account not found")
		}
		return domain.Account{}, fmt.Errorf("select account: %w", err)
	}

	return domain.NewAccount(id, domain.AccountType(accountType), name)
}

// PostTransaction inserts the journal header and all lines in a single DB transaction.
func (r *Repository) PostTransaction(ctx context.Context, tx domain.Transaction) error {
	dbTx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		_ = dbTx.Rollback(ctx)
	}()

	if err := r.PostTransactionInTx(ctx, dbTx, tx); err != nil {
		return err
	}

	if err := dbTx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

// PostTransactionInTx posts a journal inside an existing database transaction.
func (r *Repository) PostTransactionInTx(ctx context.Context, dbTx pgx.Tx, tx domain.Transaction) error {
	if !tx.IsBalanced() {
		return shareddomain.NewDomainError(domain.ErrCodeUnbalancedTransaction, "transaction is not balanced")
	}

	accountIDs := uniqueAccountIDs(tx)
	accounts, err := lockAccounts(ctx, dbTx, accountIDs)
	if err != nil {
		return err
	}

	if err := validateSufficientFunds(ctx, dbTx, accounts, tx); err != nil {
		return err
	}

	_, err = dbTx.Exec(ctx, `
		INSERT INTO ledger_transactions (id, description, recorded_at)
		VALUES ($1, $2, $3)
	`, tx.ID().String(), tx.Description(), tx.RecordedAt())
	if err != nil {
		if isUniqueViolation(err) {
			return shareddomain.NewDomainError(shareddomain.ErrCodeConflict, "transaction already posted")
		}
		return fmt.Errorf("insert transaction: %w", err)
	}

	for _, entry := range tx.Entries() {
		_, err = dbTx.Exec(ctx, `
			INSERT INTO ledger_entries (id, transaction_id, account_id, side, amount_halalas, currency)
			VALUES ($1, $2, $3, $4, $5, $6)
		`,
			entry.ID().String(),
			tx.ID().String(),
			entry.AccountID().String(),
			string(entry.Side()),
			entry.Amount().Halalas(),
			entry.Amount().Currency(),
		)
		if err != nil {
			if isUniqueViolation(err) {
				return shareddomain.NewDomainError(shareddomain.ErrCodeConflict, "entry already posted")
			}
			if isForeignKeyViolation(err) {
				return shareddomain.NewDomainError(shareddomain.ErrCodeValidation, "account does not exist")
			}
			return fmt.Errorf("insert entry %s: %w", entry.ID(), err)
		}
	}

	return nil
}

// ListEntriesByAccount returns all journal lines for an account in insertion order.
func (r *Repository) ListEntriesByAccount(ctx context.Context, accountID domain.AccountID) ([]domain.Entry, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, account_id, side, amount_halalas, currency
		FROM ledger_entries
		WHERE account_id = $1
		ORDER BY created_at ASC, id ASC
	`, accountID.String())
	if err != nil {
		return nil, fmt.Errorf("list entries: %w", err)
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
		return nil, fmt.Errorf("iterate entries: %w", err)
	}
	return entries, nil
}

// GetBalance derives the account balance from persisted entries using domain rules.
func (r *Repository) GetBalance(ctx context.Context, account domain.Account) (shareddomain.Money, error) {
	entries, err := r.ListEntriesByAccount(ctx, account.ID())
	if err != nil {
		return shareddomain.Money{}, err
	}
	return domain.BalanceFromEntries(account, entries)
}

// VerifyGlobalLedgerBalanced asserts the fundamental double-entry invariant in storage.
func (r *Repository) VerifyGlobalLedgerBalanced(ctx context.Context) error {
	row := r.pool.QueryRow(ctx, `
		SELECT
			COALESCE(SUM(CASE WHEN side = 'DEBIT' THEN amount_halalas ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN side = 'CREDIT' THEN amount_halalas ELSE 0 END), 0)
		FROM ledger_entries
	`)

	var debits, credits int64
	if err := row.Scan(&debits, &credits); err != nil {
		return fmt.Errorf("sum ledger: %w", err)
	}
	if debits != credits {
		return shareddomain.NewDomainError(
			domain.ErrCodeUnbalancedTransaction,
			fmt.Sprintf("stored debits (%d) != credits (%d)", debits, credits),
		)
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func isForeignKeyViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23503"
}
