package middleware

import (
	"encoding/json"
	"errors"
	"net/http"

	shareddomain "github.com/mohsinazam/banking/internal/shared/domain"
)

type errorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// DomainError maps domain errors to HTTP status codes (optional middleware).
func DomainError(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		if rec.err == nil {
			return
		}
		status := statusForError(rec.err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		var de *shareddomain.DomainError
		if errors.As(rec.err, &de) {
			_ = json.NewEncoder(w).Encode(errorResponse{Code: string(de.Code), Message: de.Message})
			return
		}
		_ = json.NewEncoder(w).Encode(errorResponse{Code: "INTERNAL", Message: "internal server error"})
	})
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

// SetHandlerError allows handlers to signal domain errors to the middleware.
func SetHandlerError(w http.ResponseWriter, err error) {
	if rec, ok := w.(interface{ setError(error) }); ok {
		rec.setError(err)
	}
}
