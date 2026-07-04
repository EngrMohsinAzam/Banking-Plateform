-- Demo wallet accounts and opening balances for local development.
INSERT INTO ledger_accounts (id, account_type, name)
VALUES
    ('wallet-alice', 'LIABILITY', 'Alice Wallet'),
    ('wallet-bob', 'LIABILITY', 'Bob Wallet'),
    ('suspense-funding', 'ASSET', 'Suspense Funding')
ON CONFLICT (id) DO NOTHING;

INSERT INTO ledger_transactions (id, description, recorded_at)
VALUES ('tx-seed-funding', 'opening balances for demo wallets', NOW())
ON CONFLICT (id) DO NOTHING;

INSERT INTO ledger_entries (id, transaction_id, account_id, side, amount_halalas, currency)
VALUES
    ('seed-alice-credit', 'tx-seed-funding', 'wallet-alice', 'CREDIT', 10000000, 'SAR'),
    ('seed-suspense-debit', 'tx-seed-funding', 'suspense-funding', 'DEBIT', 10000000, 'SAR')
ON CONFLICT (id) DO NOTHING;

CREATE INDEX IF NOT EXISTS idx_transfer_sagas_transaction_id ON transfer_sagas (transaction_id);
CREATE INDEX IF NOT EXISTS idx_transfer_sagas_idempotency_key ON transfer_sagas (idempotency_key);
