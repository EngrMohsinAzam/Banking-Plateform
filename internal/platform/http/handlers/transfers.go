package handlers

import (
	"encoding/json"
	"net/http"

	goredis "github.com/redis/go-redis/v9"
	"github.com/jackc/pgx/v5/pgxpool"
	platformredis "github.com/mohsinazam/banking/internal/platform/redis"
	shareddomain "github.com/mohsinazam/banking/internal/shared/domain"
	transferapp "github.com/mohsinazam/banking/internal/transfer/app"
	"github.com/mohsinazam/banking/internal/transfer/domain"
	ledgerdomain "github.com/mohsinazam/banking/internal/ledger/domain"
)

// TransferHandler serves money movement endpoints.
type TransferHandler struct {
	service *transferapp.Service
}

// NewTransferHandler constructs a transfer HTTP handler.
func NewTransferHandler(service *transferapp.Service) *TransferHandler {
	return &TransferHandler{service: service}
}

type transferRequest struct {
	FromAccountID   string `json:"from_account_id"`
	ToAccountID     string `json:"to_account_id"`
	Amount          string `json:"amount"`
	BeneficiaryIBAN string `json:"beneficiary_iban"`
	BeneficiaryName string `json:"beneficiary_name"`
	Description     string `json:"description"`
}

type transferResponse struct {
	TransactionID    string `json:"transaction_id"`
	Amount           string `json:"amount"`
	FromAccountID    string `json:"from_account_id"`
	ToAccountID      string `json:"to_account_id"`
	SagaID           string `json:"saga_id"`
	SettlementID     string `json:"settlement_id"`
	SagaState        string `json:"saga_state"`
	SettlementStatus string `json:"settlement_status"`
	Replayed         bool   `json:"replayed"`
}

type transferStatusResponse struct {
	SagaID           string `json:"saga_id"`
	SagaState        string `json:"saga_state"`
	IdempotencyKey   string `json:"idempotency_key,omitempty"`
	TransactionID    string `json:"transaction_id"`
	SettlementID     string `json:"settlement_id"`
	SettlementStatus string `json:"settlement_status"`
	SettlementError  string `json:"settlement_error,omitempty"`
	FailureReason    string `json:"failure_reason,omitempty"`
	FromAccountID    string `json:"from_account_id"`
	ToAccountID      string `json:"to_account_id"`
	Amount           string `json:"amount"`
	BeneficiaryIBAN  string `json:"beneficiary_iban"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

// Create handles POST /v1/transfers.
func (h *TransferHandler) Create(w http.ResponseWriter, r *http.Request) {
	idempotencyKey := r.Header.Get("Idempotency-Key")
	var req transferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, r, shareddomain.NewDomainError(shareddomain.ErrCodeValidation, "invalid JSON body"))
		return
	}

	amount, err := shareddomain.ParseSAR(req.Amount)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	cmd := domain.Command{
		IdempotencyKey:  idempotencyKey,
		FromAccountID:   ledgerdomain.AccountID(req.FromAccountID),
		ToAccountID:     ledgerdomain.AccountID(req.ToAccountID),
		Amount:          amount,
		BeneficiaryIBAN: req.BeneficiaryIBAN,
		BeneficiaryName: req.BeneficiaryName,
		Description:     req.Description,
	}

	result, err := h.service.Execute(r.Context(), cmd)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(transferResponse{
		TransactionID:    result.TransactionID.String(),
		Amount:           result.Amount.String(),
		FromAccountID:    result.FromAccountID.String(),
		ToAccountID:      result.ToAccountID.String(),
		SagaID:           result.SagaID,
		SettlementID:     result.SettlementID,
		SagaState:        string(result.SagaState),
		SettlementStatus: string(result.SettlementStatus),
		Replayed:         result.Replayed,
	})
}

// GetByTransactionID handles GET /v1/transfers/{transaction_id}.
func (h *TransferHandler) GetByTransactionID(w http.ResponseWriter, r *http.Request) {
	txID := r.PathValue("transaction_id")
	if txID == "" {
		WriteError(w, r, shareddomain.NewDomainError(shareddomain.ErrCodeValidation, "transaction_id is required"))
		return
	}
	status, err := h.service.GetStatusByTransactionID(r.Context(), txID)
	if err != nil {
		WriteError(w, r, err)
		return
	}
	writeStatus(w, status)
}

// GetByIdempotencyKey handles GET /v1/transfers/by-key/{idempotency_key}.
func (h *TransferHandler) GetByIdempotencyKey(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("idempotency_key")
	if key == "" {
		WriteError(w, r, shareddomain.NewDomainError(shareddomain.ErrCodeValidation, "idempotency_key is required"))
		return
	}
	status, err := h.service.GetStatusByIdempotencyKey(r.Context(), key)
	if err != nil {
		WriteError(w, r, err)
		return
	}
	writeStatus(w, status)
}

func writeStatus(w http.ResponseWriter, status domain.Status) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(transferStatusResponse{
		SagaID:           status.SagaID,
		SagaState:        string(status.SagaState),
		IdempotencyKey:   status.IdempotencyKey,
		TransactionID:    status.TransactionID,
		SettlementID:     status.SettlementID,
		SettlementStatus: string(status.SettlementStatus),
		SettlementError:  status.SettlementError,
		FailureReason:    status.FailureReason,
		FromAccountID:    status.FromAccountID.String(),
		ToAccountID:      status.ToAccountID.String(),
		Amount:           status.Amount.String(),
		BeneficiaryIBAN:  status.BeneficiaryIBAN,
		CreatedAt:        status.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:        status.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	})
}

// Ready returns readiness based on Postgres and Redis connectivity.
func Ready(pool *pgxpool.Pool, redisClient *goredis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if err := pool.Ping(ctx); err != nil {
			http.Error(w, "postgres unavailable", http.StatusServiceUnavailable)
			return
		}
		if err := platformredis.Ping(ctx, redisClient); err != nil {
			http.Error(w, "redis unavailable", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
	}
}
