package handlers

import (
	"encoding/json"
	"net/http"

	ledgerapp "github.com/mohsinazam/banking/internal/ledger/app"
	ledgerdomain "github.com/mohsinazam/banking/internal/ledger/domain"
	shareddomain "github.com/mohsinazam/banking/internal/shared/domain"
)

// AccountHandler serves account read/create endpoints.
type AccountHandler struct {
	ledger *ledgerapp.Poster
}

// NewAccountHandler constructs an account HTTP handler.
func NewAccountHandler(ledger *ledgerapp.Poster) *AccountHandler {
	return &AccountHandler{ledger: ledger}
}

type createAccountRequest struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	AccountType string `json:"account_type"`
}

type accountResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	AccountType string `json:"account_type"`
}

type balanceResponse struct {
	AccountID string `json:"account_id"`
	Balance   string `json:"balance"`
	Currency  string `json:"currency"`
}

type entryResponse struct {
	ID        string `json:"id"`
	AccountID string `json:"account_id"`
	Side      string `json:"side"`
	Amount    string `json:"amount"`
}

// Create handles POST /v1/accounts.
func (h *AccountHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, r, shareddomain.NewDomainError(shareddomain.ErrCodeValidation, "invalid JSON body"))
		return
	}
	if req.ID == "" || req.Name == "" || req.AccountType == "" {
		WriteError(w, r, shareddomain.NewDomainError(shareddomain.ErrCodeValidation, "id, name, and account_type are required"))
		return
	}

	accountType, err := parseAccountType(req.AccountType)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	account, err := ledgerdomain.NewAccount(ledgerdomain.AccountID(req.ID), accountType, req.Name)
	if err != nil {
		WriteError(w, r, err)
		return
	}
	if err := h.ledger.CreateAccount(r.Context(), account); err != nil {
		WriteError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(accountResponse{
		ID:          account.ID().String(),
		Name:        account.Name(),
		AccountType: string(account.Type()),
	})
}

// GetBalance handles GET /v1/accounts/{account_id}/balance.
func (h *AccountHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	accountID := ledgerdomain.AccountID(r.PathValue("account_id"))
	balance, err := h.ledger.GetBalanceForAccount(r.Context(), accountID)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(balanceResponse{
		AccountID: accountID.String(),
		Balance:   balance,
		Currency:  "SAR",
	})
}

// ListEntries handles GET /v1/accounts/{account_id}/entries.
func (h *AccountHandler) ListEntries(w http.ResponseWriter, r *http.Request) {
	accountID := ledgerdomain.AccountID(r.PathValue("account_id"))
	entries, err := h.ledger.ListEntries(r.Context(), accountID)
	if err != nil {
		WriteError(w, r, err)
		return
	}

	out := make([]entryResponse, 0, len(entries))
	for _, e := range entries {
		out = append(out, entryResponse{
			ID:        e.ID().String(),
			AccountID: e.AccountID().String(),
			Side:      string(e.Side()),
			Amount:    e.Amount().String(),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"entries": out})
}

func parseAccountType(raw string) (ledgerdomain.AccountType, error) {
	switch ledgerdomain.AccountType(raw) {
	case ledgerdomain.AccountTypeAsset,
		ledgerdomain.AccountTypeLiability,
		ledgerdomain.AccountTypeEquity,
		ledgerdomain.AccountTypeRevenue,
		ledgerdomain.AccountTypeExpense:
		return ledgerdomain.AccountType(raw), nil
	default:
		return "", shareddomain.NewDomainError(shareddomain.ErrCodeValidation, "invalid account_type")
	}
}
