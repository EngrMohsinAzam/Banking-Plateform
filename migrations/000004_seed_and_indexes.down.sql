DROP INDEX IF EXISTS idx_transfer_sagas_idempotency_key;
DROP INDEX IF EXISTS idx_transfer_sagas_transaction_id;

DELETE FROM ledger_entries WHERE id IN ('seed-alice-credit', 'seed-suspense-debit');
DELETE FROM ledger_transactions WHERE id = 'tx-seed-funding';
DELETE FROM ledger_accounts WHERE id IN ('wallet-alice', 'wallet-bob', 'suspense-funding');
