package http

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mohsinazam/banking/internal/platform/audit"
	"github.com/mohsinazam/banking/internal/platform/auth"
	"github.com/mohsinazam/banking/internal/platform/config"
	"github.com/mohsinazam/banking/internal/platform/http/handlers"
	"github.com/mohsinazam/banking/internal/platform/http/middleware"
	"github.com/mohsinazam/banking/internal/platform/observability"
	ledgerapp "github.com/mohsinazam/banking/internal/ledger/app"
	transferapp "github.com/mohsinazam/banking/internal/transfer/app"
)

// Dependencies bundles HTTP handler dependencies.
type Dependencies struct {
	Config          config.Config
	Logger          *slog.Logger
	Pool            *pgxpool.Pool
	RedisClient     *goredis.Client
	LedgerService   *ledgerapp.Poster
	TransferService *transferapp.Service
	AuditLogger     *audit.Logger
	APIKeyAuth      *auth.APIKeyValidator
}

// Server wraps the HTTP listener and route table.
type Server struct {
	cfg    config.Config
	logger *slog.Logger
	http   *http.Server
}

// NewServer builds the API HTTP server.
func NewServer(deps Dependencies) *Server {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handlers.Health(deps.Config.ServiceName))
	mux.HandleFunc("GET /ready", handlers.Ready(deps.Pool, deps.RedisClient))
	mux.HandleFunc("GET /openapi.yaml", handlers.OpenAPI())
	mux.Handle("GET /metrics", observability.MetricsHandler())

	transferHandler := handlers.NewTransferHandler(deps.TransferService)
	mux.Handle("POST /v1/transfers", middleware.Audit(deps.AuditLogger, deps.Logger, "transfer.create")(
		http.HandlerFunc(transferHandler.Create)))
	mux.HandleFunc("GET /v1/transfers/{transaction_id}", transferHandler.GetByTransactionID)
	mux.HandleFunc("GET /v1/transfers/by-key/{idempotency_key}", transferHandler.GetByIdempotencyKey)

	accountHandler := handlers.NewAccountHandler(deps.LedgerService)
	mux.Handle("POST /v1/accounts", middleware.Audit(deps.AuditLogger, deps.Logger, "account.create")(
		http.HandlerFunc(accountHandler.Create)))
	mux.HandleFunc("GET /v1/accounts/{account_id}/balance", accountHandler.GetBalance)
	mux.HandleFunc("GET /v1/accounts/{account_id}/entries", accountHandler.ListEntries)

	var handler http.Handler = mux
	handler = middleware.AccessLog(deps.Logger, handler)
	handler = observability.MetricsMiddleware(handler)
	handler = middleware.RequestID(handler)
	handler = middleware.APIKeyAuth(deps.APIKeyAuth, deps.Logger)(handler)
	handler = middleware.CORS(deps.Config.CORSAllowedOrigins, handler)

	return &Server{
		cfg:    deps.Config,
		logger: deps.Logger,
		http: &http.Server{
			Addr:              deps.Config.HTTPAddr,
			Handler:           handler,
			ReadHeaderTimeout: 5 * time.Second,
		},
	}
}

// ListenAndServe blocks until the server stops or returns an error.
func (s *Server) ListenAndServe() error {
	s.logger.Info("http server listening", "addr", s.cfg.HTTPAddr, "auth_enabled", s.cfg.AuthEnabled)
	if err := s.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("http server: %w", err)
	}
	return nil
}

// Shutdown gracefully stops in-flight requests.
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("http server shutting down")
	return s.http.Shutdown(ctx)
}
