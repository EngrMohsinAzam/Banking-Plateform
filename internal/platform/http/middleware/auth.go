package middleware

import (
	"log/slog"
	"net/http"

	"github.com/mohsinazam/banking/internal/platform/auth"
	"github.com/mohsinazam/banking/internal/platform/http/handlers"
)

// APIKeyAuth protects business routes with X-API-Key when enabled.
func APIKeyAuth(validator *auth.APIKeyValidator, logger *slog.Logger) func(http.Handler) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Browsers send OPTIONS preflight without API keys — always pass through to CORS.
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}
			if !validator.Enabled() {
				next.ServeHTTP(w, r)
				return
			}
			key := r.Header.Get("X-API-Key")
			if err := validator.Validate(key); err != nil {
				logger.Warn("auth rejected", "path", r.URL.Path, "reason", err)
				handlers.WriteError(w, r, err)
				return
			}
			if r.Header.Get("X-Actor") == "" {
				r.Header.Set("X-Actor", "api-key:"+maskKey(key))
			}
			next.ServeHTTP(w, r)
		})
	}
}

func maskKey(key string) string {
	if len(key) <= 8 {
		return "***"
	}
	return key[:4] + "..." + key[len(key)-4:]
}
