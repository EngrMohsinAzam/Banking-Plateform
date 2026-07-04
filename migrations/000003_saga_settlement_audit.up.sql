CREATE TABLE transfer_sagas (
    id               TEXT PRIMARY KEY,
    state            TEXT NOT NULL,
    idempotency_key  TEXT NOT NULL UNIQUE,
    command_json     JSONB NOT NULL,
    transaction_id   TEXT,
    settlement_id    TEXT,
    failure_reason   TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE settlements (
    id               TEXT PRIMARY KEY,
    saga_id          TEXT NOT NULL REFERENCES transfer_sagas (id),
    beneficiary_iban TEXT NOT NULL,
    amount_halalas   BIGINT NOT NULL CHECK (amount_halalas > 0),
    currency         TEXT NOT NULL DEFAULT 'SAR',
    status           TEXT NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'SETTLED', 'FAILED')),
    attempts         INT NOT NULL DEFAULT 0,
    last_error       TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_settlements_pending ON settlements (created_at ASC) WHERE status = 'PENDING';

CREATE TABLE audit_log (
    id            BIGSERIAL PRIMARY KEY,
    request_id    TEXT NOT NULL,
    action        TEXT NOT NULL,
    actor         TEXT,
    resource_type TEXT,
    resource_id   TEXT,
    metadata      JSONB,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_log_created_at ON audit_log (created_at DESC);
