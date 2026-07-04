DROP TRIGGER IF EXISTS ledger_entries_no_mutation ON ledger_entries;
DROP TRIGGER IF EXISTS ledger_transactions_no_mutation ON ledger_transactions;
DROP FUNCTION IF EXISTS prevent_ledger_mutation;

DROP TABLE IF EXISTS ledger_entries;
DROP TABLE IF EXISTS ledger_transactions;
DROP TABLE IF EXISTS ledger_accounts;
