# Architecture

## Style

Modular monolith with hexagonal boundaries per bounded context:

```
domain (pure) → app (use cases) → ports (interfaces) → adapters (Postgres, Redis, HTTP)
```

Two deployable binaries share the same Postgres schema:

- **api** — synchronous transfer requests
- **worker** — outbox relay, SARIE settlement, ledger reconciliation

## Transfer saga

```
Client → Idempotency (Redis)
      → Fraud check (amount + velocity)
      → Compliance check (mock sanctions)
      → Atomic commit (Postgres):
            saga row + journal + outbox event + settlement row
      → Worker settles via mock SARIE
            success → saga COMPLETED
            failure → compensation journal + saga COMPENSATED
```

## Ledger invariants

- Balances are **derived** from append-only entries (never stored as source of truth).
- Every transaction debits = credits.
- Wallet transfers use **liability** accounts with pessimistic `FOR UPDATE` locking and sorted account order to prevent deadlocks.

## Idempotency

`Idempotency-Key` + request fingerprint stored in Redis. Replays return the original result without double-posting. Ledger transaction IDs are deterministic from the idempotency key.

## Outbox

`transfer.posted` events are written in the same DB transaction as the journal. The worker polls with `FOR UPDATE SKIP LOCKED` and publishes to a log publisher (swap for Kafka in production).

## API surface

| Method | Path | Description |
|--------|------|-------------|
| POST | `/v1/transfers` | Create transfer (idempotent) |
| GET | `/v1/transfers/{transaction_id}` | Poll saga + settlement status |
| GET | `/v1/transfers/by-key/{idempotency_key}` | Status by client key |
| POST | `/v1/accounts` | Create ledger account |
| GET | `/v1/accounts/{id}/balance` | Derived balance |
| GET | `/v1/accounts/{id}/entries` | Journal lines |
| GET | `/health`, `/ready`, `/metrics` | Ops endpoints |

## Observability

- Structured JSON logs (`slog`) on API, worker, transfer saga, fraud/compliance blocks, HTTP access log
- Prometheus metrics at `/metrics`
- Audit log table for mutating HTTP actions (with `resource_id`)

## KSA context

- SAR amounts as int64 halalas (no floats)
- SA IBAN mod-97 validation
- Mock SARIE rail for external settlement (real integration would use SAMA-certified APIs)
