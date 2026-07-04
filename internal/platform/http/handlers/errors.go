package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	shareddomain "github.com/mohsinazam/banking/internal/shared/domain"
)

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// WriteError maps domain errors to HTTP responses and logs unexpected failures.
func WriteError(w http.ResponseWriter, r *http.Request, err error) {
	status := statusForError(err)
	if status >= http.StatusInternalServerError {
		slog.Error("request failed",
			"method", r.Method,
			"path", r.URL.Path,
			"status", status,
			"error", err,
		)
	} else if de, ok := shareddomain.AsDomainError(err); ok {
		slog.Warn("request rejected",
			"method", r.Method,
			"path", r.URL.Path,
			"code", de.Code,
			"message", de.Message,
		)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	var de *shareddomain.DomainError
	if errors.As(err, &de) {
		_ = json.NewEncoder(w).Encode(errorBody{Code: string(de.Code), Message: de.Message})
		return
	}
	_ = json.NewEncoder(w).Encode(errorBody{Code: "INTERNAL", Message: "internal server error"})
}

func statusForError(err error) int {
	var de *shareddomain.DomainError
	if !errors.As(err, &de) {
		return http.StatusInternalServerError
	}
	switch de.Code {
	case shareddomain.ErrCodeValidation, shareddomain.ErrCodeInvalidMoney, shareddomain.ErrCodeInvalidIBAN:
		return http.StatusBadRequest
	case shareddomain.ErrCodeInsufficientFunds:
		return http.StatusUnprocessableEntity
	case shareddomain.ErrCodeConflict, shareddomain.ErrCodeRequestInProgress:
		return http.StatusConflict
	case shareddomain.ErrCodeForbidden:
		return http.StatusForbidden
	case shareddomain.ErrCodeNotFound:
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}
