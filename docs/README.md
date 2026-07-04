# KSA Core Banking Portfolio (Modular Monolith)

A Go learning project modelling a Saudi-focused retail banking core: double-entry ledger, wallet transfers, idempotency, transactional outbox, fraud/compliance saga steps, mock SARIE settlement, and compensation.

## Quick start

```bash
make up          # Postgres + Redis
make migrate-up  # includes demo wallet seed (Alice 100,000 SAR)
make run-api     # HTTP API on :8080
make run-worker  # outbox + settlement + reconciliation
make test        # unit tests (no Docker)
```

## API

### Transfers

```http
POST /v1/transfers
Idempotency-Key: <uuid>
Content-Type: application/json

{
  "from_account_id": "wallet-alice",
  "to_account_id": "wallet-bob",
  "amount": "100.00",
  "beneficiary_iban": "SA0380000000608010167519",
  "beneficiary_name": "Mohsin Azam",
  "description": "rent"
}
```

Response includes `saga_id`, `settlement_id`, `saga_state`, `settlement_status` for polling.

```http
GET /v1/transfers/{transaction_id}
GET /v1/transfers/by-key/{idempotency_key}
```

### Accounts

```http
POST /v1/accounts
{"id":"wallet-carol","name":"Carol","account_type":"LIABILITY"}

GET /v1/accounts/wallet-alice/balance
GET /v1/accounts/wallet-alice/entries
```

### Ops

| Endpoint | Purpose |
|----------|---------|
| `GET /health` | Liveness |
| `GET /ready` | Postgres + Redis |
| `GET /metrics` | Prometheus |

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
| `AUTO_MIGRATE` | `true` | Run migrations on boot |
| `FRAUD_MAX_SINGLE_HALALAS` | `50000000` | 500k SAR limit |
| `FRAUD_MAX_HOURLY_TRANSFERS` | `20` | Velocity cap |
| `SARIE_FAIL_RATE` | `0.15` | Mock rail failure rate |
| `WORKER_INTERVAL_SEC` | `3` | Settlement/reconcile loop |
| `OUTBOX_INTERVAL_SEC` | `2` | Outbox relay poll |

See [ARCHITECTURE.md](./ARCHITECTURE.md) and [RUNBOOK.md](./RUNBOOK.md).
