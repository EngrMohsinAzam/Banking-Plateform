package middleware

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/mohsinazam/banking/internal/platform/audit"
)

// Audit logs successful money-moving HTTP actions.
func Audit(logger *audit.Logger, slogger *slog.Logger, action string) func(http.Handler) http.Handler {
	if slogger == nil {
		slogger = slog.Default()
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)
			if rec.status >= 400 {
				return
			}

			resourceID, resourceType := extractAuditResource(rec.body.Bytes())
			if err := logger.Log(r.Context(), audit.Entry{
				RequestID:    RequestIDFromContext(r.Context()),
				Action:       action,
				Actor:        r.Header.Get("X-Actor"),
				ResourceType: resourceType,
				ResourceID:   resourceID,
				Metadata: map[string]any{
					"method": r.Method,
					"path":   r.URL.Path,
					"status": rec.status,
				},
			}); err != nil {
				slogger.Error("audit log failed", "action", action, "error", err)
			}
		})
	}
}

func extractAuditResource(body []byte) (id, resourceType string) {
	if len(body) == 0 {
		return "", ""
	}
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", ""
	}
	if v, ok := payload["transaction_id"].(string); ok && v != "" {
		return v, "transfer"
	}
	if v, ok := payload["id"].(string); ok && v != "" {
		return v, "account"
	}
	return "", ""
}
