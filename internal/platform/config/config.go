package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds runtime configuration loaded from environment variables.
type Config struct {
	ServiceName string
	HTTPAddr    string
	LogLevel    slog.Level

	PostgresDSN string
	RedisAddr   string

	AutoMigrate     bool
	ShutdownTimeout time.Duration

	AuthEnabled bool
	APIKeys     []string

	CORSAllowedOrigins []string

	EventPublisher string
	KafkaBrokers   []string
	KafkaTopic     string

	WorkerInterval          time.Duration
	OutboxInterval          time.Duration
	SarieFailRate           float64
	FraudMaxSingleHalalas   int64
	FraudMaxHourlyTransfers int64
}

// Load reads configuration from the environment with sensible local defaults.
func Load() (Config, error) {
	cfg := Config{
		ServiceName: getenv("SERVICE_NAME", "banking-api"),
		HTTPAddr:    getenv("HTTP_ADDR", ":8080"),
		LogLevel:    parseLogLevel(getenv("LOG_LEVEL", "info")),

		PostgresDSN: getenv(
			"POSTGRES_DSN",
			"postgres://banking:banking@localhost:5432/banking?sslmode=disable",
		),
		RedisAddr: getenv("REDIS_ADDR", "localhost:6379"),

		AutoMigrate:     getenv("AUTO_MIGRATE", "true") == "true",
		ShutdownTimeout: 10 * time.Second,

		AuthEnabled:    getenv("AUTH_ENABLED", "false") == "true",
		APIKeys:        parseCSV(getenv("API_KEYS", "dev-banking-key-change-me")),
		CORSAllowedOrigins: parseCSV(getenv("CORS_ALLOWED_ORIGINS", "http://localhost:3000")),
		EventPublisher: getenv("EVENT_PUBLISHER", "log"),
		KafkaBrokers:   parseCSV(getenv("KAFKA_BROKERS", "localhost:9092")),
		KafkaTopic:     getenv("KAFKA_TOPIC", "banking.events"),

		WorkerInterval:          3 * time.Second,
		OutboxInterval:          2 * time.Second,
		SarieFailRate:           0.15,
		FraudMaxSingleHalalas:   500_000_00,
		FraudMaxHourlyTransfers: 20,
	}

	if v := os.Getenv("SHUTDOWN_TIMEOUT_SEC"); v != "" {
		sec, err := strconv.Atoi(v)
		if err != nil {
			return Config{}, fmt.Errorf("invalid SHUTDOWN_TIMEOUT_SEC: %w", err)
		}
		cfg.ShutdownTimeout = time.Duration(sec) * time.Second
	}
	if v := os.Getenv("WORKER_INTERVAL_SEC"); v != "" {
		sec, err := strconv.Atoi(v)
		if err != nil {
			return Config{}, fmt.Errorf("invalid WORKER_INTERVAL_SEC: %w", err)
		}
		cfg.WorkerInterval = time.Duration(sec) * time.Second
	}
	if v := os.Getenv("OUTBOX_INTERVAL_SEC"); v != "" {
		sec, err := strconv.Atoi(v)
		if err != nil {
			return Config{}, fmt.Errorf("invalid OUTBOX_INTERVAL_SEC: %w", err)
		}
		cfg.OutboxInterval = time.Duration(sec) * time.Second
	}
	if v := os.Getenv("SARIE_FAIL_RATE"); v != "" {
		rate, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return Config{}, fmt.Errorf("invalid SARIE_FAIL_RATE: %w", err)
		}
		cfg.SarieFailRate = rate
	}
	if v := os.Getenv("FRAUD_MAX_SINGLE_HALALAS"); v != "" {
		val, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return Config{}, fmt.Errorf("invalid FRAUD_MAX_SINGLE_HALALAS: %w", err)
		}
		cfg.FraudMaxSingleHalalas = val
	}
	if v := os.Getenv("FRAUD_MAX_HOURLY_TRANSFERS"); v != "" {
		val, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return Config{}, fmt.Errorf("invalid FRAUD_MAX_HOURLY_TRANSFERS: %w", err)
		}
		cfg.FraudMaxHourlyTransfers = val
	}

	return cfg, nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func parseLogLevel(raw string) slog.Level {
	switch raw {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
