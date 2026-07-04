# KSA Core Banking — Modular Monolith

Portfolio-grade Go backend modelling Saudi retail banking: double-entry ledger, wallet transfers, idempotency, transactional outbox, fraud/compliance/KYC, mock SARIE settlement, and compensation.

## Features

- **Ledger** — append-only double-entry journal, derived balances, `FOR UPDATE` concurrency
- **Transfers** — saga (fraud → compliance → post → settlement) with idempotency
- **Outbox** — reliable events (log or Kafka publisher)
- **Worker** — outbox relay, SARIE mock settlement, reconciliation
- **HTTP API** — OpenAPI spec, optional API-key auth, Prometheus metrics, audit log
- **KSA** — SAR halalas, SA IBAN validation, mock sanctions + KYC

## Quick start

```bash
cp .env.example .env   # optional
make up
make migrate-up
make run-api           # terminal 1
make run-worker        # terminal 2
```

Demo wallets (`wallet-alice` with 100,000 SAR) are seeded by migration `000004`.

## API

| Method | Path | Description |
|--------|------|-------------|
| POST | `/v1/transfers` | Create transfer (`Idempotency-Key` required) |
| GET | `/v1/transfers/{transaction_id}` | Saga status |
| GET | `/v1/transfers/by-key/{idempotency_key}` | Status by idempotency key |
| POST | `/v1/accounts` | Create account |
| GET | `/v1/accounts/{id}/balance` | Balance |
| GET | `/v1/accounts/{id}/entries` | Journal lines |
| GET | `/health`, `/ready`, `/metrics`, `/openapi.yaml` | Ops |

When `AUTH_ENABLED=true`, send `X-API-Key` on `/v1/*` routes.

```bash
curl -X POST http://localhost:8080/v1/transfers \
  -H "Idempotency-Key: demo-001" \
  -H "Content-Type: application/json" \
  -d '{"from_account_id":"wallet-alice","to_account_id":"wallet-bob","amount":"50.00","beneficiary_iban":"SA0380000000608010167519","beneficiary_name":"Bob"}'
```

## Documentation

- [Architecture](docs/ARCHITECTURE.md)
- [Runbook](docs/RUNBOOK.md)
- [OpenAPI](api/openapi.yaml) — also at `GET /openapi.yaml`

## Development

```bash
make test              # unit tests
make test-integration  # Postgres integration (Docker)
make test-load         # throughput benchmark
make build             # bin/api, bin/worker
make docker-build      # container images
```

## Stack

Go 1.25 · PostgreSQL 16 · Redis 7 · Redpanda/Kafka (optional) · **Next.js** web console

### Web frontend (`web/`)

```bash
cd web
cp .env.local.example .env.local
npm install
npm run dev    # http://localhost:3000
```

Requires the API running with CORS (`CORS_ALLOWED_ORIGINS=http://localhost:3000` — default).

## License

MIT — see [LICENSE](LICENSE).
