package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

type HealthResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Timestamp string `json:"timestamp"`
}

// Health returns a liveness probe handler.
// Step 0: process is up. Later steps add /ready with Postgres/Redis checks.
func Health(serviceName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		_ = json.NewEncoder(w).Encode(HealthResponse{
			Status:    "ok",
			Service:   serviceName,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		})
	}
}
