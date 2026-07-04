export function cn(...classes: (string | false | null | undefined)[]): string {
  return classes.filter(Boolean).join(" ");
}

export function newIdempotencyKey(): string {
  return crypto.randomUUID();
}

export function formatSAR(amount: string): string {
  const n = parseFloat(amount);
  if (Number.isNaN(n)) return amount;
  return new Intl.NumberFormat("en-SA", {
    style: "currency",
    currency: "SAR",
    minimumFractionDigits: 2,
  }).format(n);
}

const SAGA_LABELS: Record<string, string> = {
  STARTED: "Started",
  FRAUD_OK: "Fraud check passed",
  COMPLIANCE_OK: "Compliance cleared",
  POSTED: "Posted to ledger",
  SETTLING: "Settling via SARIE",
  COMPLETED: "Completed",
  COMPENSATING: "Reversing funds",
  COMPENSATED: "Compensated",
  FAILED: "Failed",
};

const SETTLEMENT_LABELS: Record<string, string> = {
  PENDING: "Awaiting settlement",
  SETTLED: "Settled",
  FAILED: "Settlement failed",
};

export function humanSagaState(state: string): string {
  return SAGA_LABELS[state] ?? state.replaceAll("_", " ").toLowerCase();
}

export function humanSettlementStatus(status: string): string {
  return SETTLEMENT_LABELS[status] ?? status.toLowerCase();
}

export function sagaTone(state: string): "success" | "warning" | "danger" | "info" {
  if (state === "COMPLETED") return "success";
  if (state === "COMPENSATED" || state === "FAILED") return "danger";
  if (state === "SETTLING" || state === "COMPENSATING") return "warning";
  return "info";
}
