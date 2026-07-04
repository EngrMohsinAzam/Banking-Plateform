package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	shareddomain "github.com/mohsinazam/banking/internal/shared/domain"
	"github.com/mohsinazam/banking/internal/platform/http/handlers"
)

func TestWriteErrorMapsNotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/v1/transfers/missing", nil)
	rec := httptest.NewRecorder()

	handlers.WriteError(rec, req, shareddomain.NewDomainError(shareddomain.ErrCodeNotFound, "saga not found"))

	require.Equal(t, http.StatusNotFound, rec.Code)
	require.Contains(t, rec.Body.String(), "NOT_FOUND")
}

func TestWriteErrorMapsForbidden(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/v1/transfers", nil)
	rec := httptest.NewRecorder()

	handlers.WriteError(rec, req, shareddomain.NewDomainError(shareddomain.ErrCodeForbidden, "blocked"))

	require.Equal(t, http.StatusForbidden, rec.Code)
}
