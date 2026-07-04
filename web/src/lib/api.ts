import type { ApiSettings } from "./types";
import { loadSettings } from "./settings";

export class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
    public code?: string,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

async function request<T>(
  path: string,
  options: RequestInit = {},
  settings?: ApiSettings,
): Promise<T> {
  const cfg = settings ?? loadSettings();
  const headers = new Headers(options.headers);
  headers.set("Content-Type", "application/json");
  if (cfg.apiKey) headers.set("X-API-Key", cfg.apiKey);
  if (cfg.actor) headers.set("X-Actor", cfg.actor);

  const base = (cfg.baseUrl || "/api").replace(/\/$/, "");
  const res = await fetch(`${base}${path}`, {
    ...options,
    headers,
  });

  if (!res.ok) {
    let message = res.statusText;
    let code: string | undefined;
    try {
      const body = await res.json();
      message = body.message ?? message;
      code = body.code;
    } catch {
      /* ignore */
    }
    throw new ApiError(message, res.status, code);
  }

  if (res.status === 204) return undefined as T;
  return res.json() as Promise<T>;
}

export const api = {
  health: () => request<{ status: string; service: string; timestamp: string }>("/health"),
  ready: () => request<{ status: string }>("/ready"),
  getBalance: (accountId: string) =>
    request<{ account_id: string; balance: string; currency: string }>(
      `/v1/accounts/${encodeURIComponent(accountId)}/balance`,
    ),
  getEntries: (accountId: string) =>
    request<{ entries: { id: string; account_id: string; side: string; amount: string }[] }>(
      `/v1/accounts/${encodeURIComponent(accountId)}/entries`,
    ),
  createAccount: (body: { id: string; name: string; account_type: string }) =>
    request<{ id: string; name: string; account_type: string }>("/v1/accounts", {
      method: "POST",
      body: JSON.stringify(body),
    }),
  createTransfer: (body: object, idempotencyKey: string) =>
    request<import("./types").TransferResponse>("/v1/transfers", {
      method: "POST",
      headers: { "Idempotency-Key": idempotencyKey },
      body: JSON.stringify(body),
    }),
  getTransferByTx: (txId: string) =>
    request<import("./types").TransferStatus>(
      `/v1/transfers/${encodeURIComponent(txId)}`,
    ),
  getTransferByKey: (key: string) =>
    request<import("./types").TransferStatus>(
      `/v1/transfers/by-key/${encodeURIComponent(key)}`,
    ),
};
