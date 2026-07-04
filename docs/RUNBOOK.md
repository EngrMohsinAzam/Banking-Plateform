# Runbook

## Prerequisites

- Go 1.25+
- Docker Desktop (for Postgres/Redis and integration tests)

## Local development

```bash
make up
make migrate-up
make run-api      # terminal 1
make run-worker   # terminal 2
```

Environment variables (defaults in `internal/platform/config`):

| Variable | Default |
|----------|---------|
| `HTTP_ADDR` | `:8080` |
| `POSTGRES_DSN` | `postgres://banking:banking@localhost:5432/banking?sslmode=disable` |
| `REDIS_ADDR` | `localhost:6379` |
| `LOG_LEVEL` | `info` |
| `AUTO_MIGRATE` | `true` |
| `AUTH_ENABLED` | `false` |
| `API_KEYS` | `dev-banking-key-change-me` |
| `EVENT_PUBLISHER` | `log` (`log` or `kafka`) |
| `KAFKA_BROKERS` | `localhost:9092` |
| `KAFKA_TOPIC` | `banking.events` |

Migration `000004` seeds `wallet-alice` (100,000 SAR), `wallet-bob`, and `suspense-funding`.

## Auth

Set `AUTH_ENABLED=true` and send `X-API-Key` on `/v1/*` routes. `/health`, `/ready`, `/metrics`, and `/openapi.yaml` stay public.

## Kafka (optional)

`make up` starts Redpanda on `:9092`. Set `EVENT_PUBLISHER=kafka` and restart the worker.

## Health checks

- **Liveness:** `GET /health` — process is up
- **Readiness:** `GET /ready` — Postgres + Redis reachable

## Operations

### Stuck idempotency key

If a client sees `409` with `request_in_progress`, the original request may still be running or Redis TTL expired mid-flight. Wait for `ProcessingTTL` (default 15 minutes) or inspect Redis key `idem:transfer:<key>`.

### Failed settlement / compensation

Check `settlements` and `transfer_sagas` tables:

```sql
SELECT id, status, last_error FROM settlements WHERE status = 'FAILED';
SELECT id, state, failure_reason FROM transfer_sagas WHERE state IN ('FAILED', 'COMPENSATED');
```

Compensated transfers reverse the internal journal; external SARIE failure is expected with the mock rail (~15% fail rate).

### Reconciliation

Worker runs `VerifyGlobalLedgerBalanced` every 3s. If it logs reconciliation errors, stop traffic and inspect unbalanced journal rows.

## Testing

```bash
make test                    # unit tests
make test-integration        # needs Docker
go test -bench=. ./tests/load/...   # throughput benchmark
```

## Migrations

```bash
make migrate-up
make migrate-down   # rolls back one step
```
