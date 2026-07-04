package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	shareddomain "github.com/mohsinazam/banking/internal/shared/domain"
	ledgerdomain "github.com/mohsinazam/banking/internal/ledger/domain"
	"github.com/mohsinazam/banking/internal/transfer/domain"
)

// SagaRepository persists saga and settlement state.
type SagaRepository struct {
	pool *pgxpool.Pool
}

// NewSagaRepository constructs a saga repository.
func NewSagaRepository(pool *pgxpool.Pool) *SagaRepository {
	return &SagaRepository{pool: pool}
}

func (r *SagaRepository) CreateSaga(ctx context.Context, saga domain.SagaRecord) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO transfer_sagas (id, state, idempotency_key, command_json, transaction_id, settlement_id, failure_reason, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, saga.ID, string(saga.State), saga.IdempotencyKey, saga.CommandJSON, nullString(saga.TransactionID),
		nullString(saga.SettlementID), nullString(saga.FailureReason), saga.CreatedAt, saga.UpdatedAt)
	return err
}

func (r *SagaRepository) UpdateSagaState(ctx context.Context, sagaID string, state domain.SagaState, fields map[string]string) error {
	txID := fields["transaction_id"]
	settlementID := fields["settlement_id"]
	failure := fields["failure_reason"]
	_, err := r.pool.Exec(ctx, `
		UPDATE transfer_sagas
		SET state = $2,
		    transaction_id = COALESCE(NULLIF($3, ''), transaction_id),
		    settlement_id = COALESCE(NULLIF($4, ''), settlement_id),
		    failure_reason = COALESCE(NULLIF($5, ''), failure_reason),
		    updated_at = NOW()
		WHERE id = $1
	`, sagaID, string(state), txID, settlementID, failure)
	return err
}

func (r *SagaRepository) GetSagaByIdempotencyKey(ctx context.Context, key string) (domain.SagaRecord, error) {
	return r.scanSaga(r.pool.QueryRow(ctx, `
		SELECT id, state, idempotency_key, command_json, transaction_id, settlement_id, failure_reason, created_at, updated_at
		FROM transfer_sagas WHERE idempotency_key = $1
	`, key))
}

func (r *SagaRepository) GetSagaByTransactionID(ctx context.Context, transactionID string) (domain.SagaRecord, error) {
	return r.scanSaga(r.pool.QueryRow(ctx, `
		SELECT id, state, idempotency_key, command_json, transaction_id, settlement_id, failure_reason, created_at, updated_at
		FROM transfer_sagas WHERE transaction_id = $1
	`, transactionID))
}

func (r *SagaRepository) GetSaga(ctx context.Context, id string) (domain.SagaRecord, error) {
	return r.scanSaga(r.pool.QueryRow(ctx, `
		SELECT id, state, idempotency_key, command_json, transaction_id, settlement_id, failure_reason, created_at, updated_at
		FROM transfer_sagas WHERE id = $1
	`, id))
}

func (r *SagaRepository) GetSettlementBySagaID(ctx context.Context, sagaID string) (domain.SettlementRecord, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, saga_id, beneficiary_iban, amount_halalas, currency, status, attempts, last_error, created_at, updated_at
		FROM settlements WHERE saga_id = $1
	`, sagaID)
	var rec domain.SettlementRecord
	var status string
	var lastError *string
	if err := row.Scan(&rec.ID, &rec.SagaID, &rec.BeneficiaryIBAN, &rec.AmountHalalas, &rec.Currency,
		&status, &rec.Attempts, &lastError, &rec.CreatedAt, &rec.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return domain.SettlementRecord{}, shareddomain.NewDomainError(shareddomain.ErrCodeNotFound, "settlement not found")
		}
		return domain.SettlementRecord{}, err
	}
	rec.Status = domain.SettlementStatus(status)
	if lastError != nil {
		rec.LastError = *lastError
	}
	return rec, nil
}

func (r *SagaRepository) CreateSettlement(ctx context.Context, settlement domain.SettlementRecord) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO settlements (id, saga_id, beneficiary_iban, amount_halalas, currency, status, attempts, last_error, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, settlement.ID, settlement.SagaID, settlement.BeneficiaryIBAN, settlement.AmountHalalas,
		settlement.Currency, string(settlement.Status), settlement.Attempts, nullString(settlement.LastError),
		settlement.CreatedAt, settlement.UpdatedAt)
	return err
}

func (r *SagaRepository) ClaimPendingSettlements(ctx context.Context, limit int) ([]domain.SettlementRecord, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	rows, err := tx.Query(ctx, `
		SELECT id, saga_id, beneficiary_iban, amount_halalas, currency, status, attempts, last_error, created_at, updated_at
		FROM settlements
		WHERE status = 'PENDING'
		ORDER BY created_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.SettlementRecord
	for rows.Next() {
		rec, err := scanSettlement(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, rec)
		_, err = tx.Exec(ctx, `UPDATE settlements SET attempts = attempts + 1, updated_at = NOW() WHERE id = $1`, rec.ID)
		if err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *SagaRepository) UpdateSettlementStatus(ctx context.Context, id string, status domain.SettlementStatus, lastError string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE settlements SET status = $2, last_error = $3, updated_at = NOW() WHERE id = $1
	`, id, string(status), nullString(lastError))
	return err
}

func (r *SagaRepository) scanSaga(row pgx.Row) (domain.SagaRecord, error) {
	var rec domain.SagaRecord
	var state string
	var cmd []byte
	var txID, settlementID, failure *string
	if err := row.Scan(&rec.ID, &state, &rec.IdempotencyKey, &cmd, &txID, &settlementID, &failure, &rec.CreatedAt, &rec.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return domain.SagaRecord{}, shareddomain.NewDomainError(shareddomain.ErrCodeNotFound, "saga not found")
		}
		return domain.SagaRecord{}, err
	}
	rec.State = domain.SagaState(state)
	rec.CommandJSON = cmd
	if txID != nil {
		rec.TransactionID = *txID
	}
	if settlementID != nil {
		rec.SettlementID = *settlementID
	}
	if failure != nil {
		rec.FailureReason = *failure
	}
	return rec, nil
}

func scanSettlement(rows pgx.Rows) (domain.SettlementRecord, error) {
	var rec domain.SettlementRecord
	var status string
	var lastError *string
	if err := rows.Scan(&rec.ID, &rec.SagaID, &rec.BeneficiaryIBAN, &rec.AmountHalalas, &rec.Currency,
		&status, &rec.Attempts, &lastError, &rec.CreatedAt, &rec.UpdatedAt); err != nil {
		return domain.SettlementRecord{}, err
	}
	rec.Status = domain.SettlementStatus(status)
	if lastError != nil {
		rec.LastError = *lastError
	}
	return rec, nil
}

func nullString(v string) any {
	if v == "" {
		return nil
	}
	return v
}

// InsertSettlementInTx stages settlement in the same DB tx as ledger/outbox.
func (r *SagaRepository) InsertSettlementInTx(ctx context.Context, tx any, settlement domain.SettlementRecord) error {
	dbTx, ok := tx.(pgx.Tx)
	if !ok {
		return fmt.Errorf("expected pgx.Tx, got %T", tx)
	}
	_, err := dbTx.Exec(ctx, `
		INSERT INTO settlements (id, saga_id, beneficiary_iban, amount_halalas, currency, status, attempts, last_error, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, settlement.ID, settlement.SagaID, settlement.BeneficiaryIBAN, settlement.AmountHalalas,
		settlement.Currency, string(settlement.Status), settlement.Attempts, nullString(settlement.LastError),
		settlement.CreatedAt, settlement.UpdatedAt)
	return err
}

// InsertSagaInTx stages saga row in the same DB tx.
func (r *SagaRepository) InsertSagaInTx(ctx context.Context, tx any, saga domain.SagaRecord) error {
	dbTx, ok := tx.(pgx.Tx)
	if !ok {
		return fmt.Errorf("expected pgx.Tx, got %T", tx)
	}
	_, err := dbTx.Exec(ctx, `
		INSERT INTO transfer_sagas (id, state, idempotency_key, command_json, transaction_id, settlement_id, failure_reason, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, saga.ID, string(saga.State), saga.IdempotencyKey, saga.CommandJSON, nullString(saga.TransactionID),
		nullString(saga.SettlementID), nullString(saga.FailureReason), saga.CreatedAt, saga.UpdatedAt)
	return err
}

// UpdateSagaStateInTx updates saga inside an open transaction.
func (r *SagaRepository) UpdateSagaStateInTx(ctx context.Context, tx any, sagaID string, state domain.SagaState, fields map[string]string) error {
	dbTx, ok := tx.(pgx.Tx)
	if !ok {
		return fmt.Errorf("expected pgx.Tx, got %T", tx)
	}
	_, err := dbTx.Exec(ctx, `
		UPDATE transfer_sagas
		SET state = $2,
		    transaction_id = COALESCE(NULLIF($3, ''), transaction_id),
		    settlement_id = COALESCE(NULLIF($4, ''), settlement_id),
		    failure_reason = COALESCE(NULLIF($5, ''), failure_reason),
		    updated_at = NOW()
		WHERE id = $1
	`, sagaID, string(state), fields["transaction_id"], fields["settlement_id"], fields["failure_reason"])
	return err
}

// CommandFromJSON decodes a persisted command.
func CommandFromJSON(data []byte) (domain.Command, error) {
	var payload struct {
		IdempotencyKey  string `json:"idempotency_key"`
		FromAccountID   string `json:"from_account_id"`
		ToAccountID     string `json:"to_account_id"`
		Amount          string `json:"amount"`
		BeneficiaryIBAN string `json:"beneficiary_iban"`
		Description     string `json:"description"`
		BeneficiaryName string `json:"beneficiary_name"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return domain.Command{}, err
	}
	amount, err := shareddomain.ParseSAR(payload.Amount)
	if err != nil {
		return domain.Command{}, err
	}
	return domain.Command{
		IdempotencyKey:  payload.IdempotencyKey,
		FromAccountID:   ledgerdomain.AccountID(payload.FromAccountID),
		ToAccountID:     ledgerdomain.AccountID(payload.ToAccountID),
		Amount:          amount,
		BeneficiaryIBAN: payload.BeneficiaryIBAN,
		Description:     payload.Description,
		BeneficiaryName: payload.BeneficiaryName,
	}, nil
}
