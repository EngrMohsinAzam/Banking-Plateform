package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// AccessLog records structured HTTP access logs for every request.
func AccessLog(logger *slog.Logger, next http.Handler) http.Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		attrs := []any{
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration_ms", time.Since(start).Milliseconds(),
			"request_id", RequestIDFromContext(r.Context()),
			"remote_addr", r.RemoteAddr,
		}
		if rec.status >= 500 {
			logger.Error("http request", attrs...)
		} else if rec.status >= 400 {
			logger.Warn("http request", attrs...)
		} else {
			logger.Info("http request", attrs...)
		}
	})
}
