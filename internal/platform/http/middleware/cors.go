package middleware

import (
	"net/http"
	"strings"
)

// CORS allows browser clients (Next.js) to call the banking API.
// Must be the outermost middleware so preflight OPTIONS always receives headers.
func CORS(allowedOrigins []string, next http.Handler) http.Handler {
	origins := make(map[string]struct{}, len(allowedOrigins)+4)
	for _, o := range allowedOrigins {
		o = strings.TrimSpace(o)
		if o != "" {
			origins[o] = struct{}{}
		}
	}
	// Common local dev origins (localhost vs 127.0.0.1).
	for _, o := range []string{
		"http://localhost:3000",
		"http://127.0.0.1:3000",
		"http://localhost:3001",
		"http://127.0.0.1:3001",
	} {
		origins[o] = struct{}{}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			if _, ok := origins[origin]; ok {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Vary", "Origin")
			}
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Idempotency-Key, X-API-Key, X-Request-ID, X-Actor")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
