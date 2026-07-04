export type DomainError = {
  code: string;
  message: string;
};

export type HealthResponse = {
  status: string;
  service: string;
  timestamp: string;
};

export type ReadyResponse = {
  status: string;
};

export type TransferRequest = {
  from_account_id: string;
  to_account_id: string;
  amount: string;
  beneficiary_iban: string;
  beneficiary_name?: string;
  description?: string;
};

export type TransferResponse = {
  transaction_id: string;
  amount: string;
  from_account_id: string;
  to_account_id: string;
  saga_id: string;
  settlement_id: string;
  saga_state: string;
  settlement_status: string;
  replayed: boolean;
};

export type TransferStatus = {
  saga_id: string;
  saga_state: string;
  idempotency_key?: string;
  transaction_id: string;
  settlement_id: string;
  settlement_status: string;
  settlement_error?: string;
  failure_reason?: string;
  from_account_id: string;
  to_account_id: string;
  amount: string;
  beneficiary_iban: string;
  created_at: string;
  updated_at: string;
};

export type BalanceResponse = {
  account_id: string;
  balance: string;
  currency: string;
};

export type LedgerEntry = {
  id: string;
  account_id: string;
  side: string;
  amount: string;
};

export type CreateAccountRequest = {
  id: string;
  name: string;
  account_type: "ASSET" | "LIABILITY" | "EQUITY" | "REVENUE" | "EXPENSE";
};

export type AccountResponse = {
  id: string;
  name: string;
  account_type: string;
};

export type ApiSettings = {
  baseUrl: string;
  apiKey: string;
  actor: string;
};
