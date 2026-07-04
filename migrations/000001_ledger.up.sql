-- Append-only ledger schema. Balances are derived from ledger_entries at read time.

CREATE TABLE ledger_accounts (
    id           TEXT PRIMARY KEY,
    account_type TEXT NOT NULL CHECK (account_type IN ('ASSET', 'LIABILITY', 'EQUITY', 'REVENUE', 'EXPENSE')),
    name         TEXT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE ledger_transactions (
    id           TEXT PRIMARY KEY,
    description  TEXT NOT NULL DEFAULT '',
    recorded_at  TIMESTAMPTZ NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE ledger_entries (
    id             TEXT PRIMARY KEY,
    transaction_id TEXT NOT NULL REFERENCES ledger_transactions (id),
    account_id     TEXT NOT NULL REFERENCES ledger_accounts (id),
    side           TEXT NOT NULL CHECK (side IN ('DEBIT', 'CREDIT')),
    amount_halalas BIGINT NOT NULL CHECK (amount_halalas > 0),
    currency       TEXT NOT NULL DEFAULT 'SAR',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ledger_entries_account_id ON ledger_entries (account_id);
CREATE INDEX idx_ledger_entries_transaction_id ON ledger_entries (transaction_id);

-- Enforce append-only ledger at the database layer.
CREATE OR REPLACE FUNCTION prevent_ledger_mutation()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'ledger tables are append-only';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER ledger_transactions_no_mutation
    BEFORE UPDATE OR DELETE ON ledger_transactions
    FOR EACH ROW EXECUTE FUNCTION prevent_ledger_mutation();

CREATE TRIGGER ledger_entries_no_mutation
    BEFORE UPDATE OR DELETE ON ledger_entries
    FOR EACH ROW EXECUTE FUNCTION prevent_ledger_mutation();
