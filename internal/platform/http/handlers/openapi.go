package handlers

import (
	"net/http"

	"github.com/mohsinazam/banking/api"
)

// OpenAPI serves the embedded OpenAPI specification.
func OpenAPI() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(api.OpenAPISpec)
	}
}
