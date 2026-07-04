"use client";

import { Suspense, useCallback, useEffect, useState } from "react";
import { useSearchParams } from "next/navigation";
import { RefreshCw, Search } from "lucide-react";
import { api, ApiError } from "@/lib/api";
import type { TransferStatus } from "@/lib/types";
import {
  formatSAR,
  humanSagaState,
  humanSettlementStatus,
  sagaTone,
} from "@/lib/utils";
import { Alert, Badge, Button, Card, Input, Label } from "@/components/ui";

const TERMINAL = new Set(["COMPLETED", "COMPENSATED", "FAILED"]);

function TrackContent() {
  const params = useSearchParams();
  const [mode, setMode] = useState<"tx" | "key">("tx");
  const [query, setQuery] = useState(params.get("tx") ?? "");
  const [status, setStatus] = useState<TransferStatus | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [polling, setPolling] = useState(false);

  const fetchStatus = useCallback(async () => {
    if (!query.trim()) return;
    setError(null);
    try {
      const res =
        mode === "tx"
          ? await api.getTransferByTx(query.trim())
          : await api.getTransferByKey(query.trim());
      setStatus(res);
      return res;
    } catch (err) {
      setStatus(null);
      setError(err instanceof ApiError ? err.message : "Lookup failed");
      return null;
    }
  }, [mode, query]);

  useEffect(() => {
    const tx = params.get("tx");
    if (tx) {
      setMode("tx");
      setQuery(tx);
    }
  }, [params]);

  useEffect(() => {
    if (!query.trim()) return;
    let cancelled = false;
    let timer: ReturnType<typeof setTimeout>;

    async function poll() {
      setPolling(true);
      const res = await fetchStatus();
      if (cancelled) return;
      if (res && !TERMINAL.has(res.saga_state)) {
        timer = setTimeout(poll, 3000);
      } else {
        setPolling(false);
      }
    }
    poll();
    return () => {
      cancelled = true;
      clearTimeout(timer);
      setPolling(false);
    };
  }, [query, mode, fetchStatus]);

  return (
    <div className="mx-auto max-w-2xl space-y-6 animate-in">
      <header>
        <h1 className="flex items-center gap-2 text-2xl font-bold">
          <Search className="h-7 w-7 text-[var(--accent)]" />
          Track transfer
        </h1>
        <p className="mt-1 text-[var(--text-muted)]">
          Follow your transfer through fraud checks, ledger posting, and SARIE settlement.
        </p>
      </header>

      <Card>
        <div className="mb-4 flex gap-2">
          <Button
            type="button"
            variant={mode === "tx" ? "primary" : "secondary"}
            onClick={() => setMode("tx")}
          >
            Transaction ID
          </Button>
          <Button
            type="button"
            variant={mode === "key" ? "primary" : "secondary"}
            onClick={() => setMode("key")}
          >
            Idempotency key
          </Button>
        </div>
        <Label htmlFor="q">{mode === "tx" ? "Transaction ID" : "Idempotency key"}</Label>
        <div className="flex gap-2">
          <Input
            id="q"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder={mode === "tx" ? "tx_…" : "uuid"}
            className="font-mono text-xs"
          />
          <Button type="button" variant="secondary" onClick={() => fetchStatus()}>
            <RefreshCw className={`h-4 w-4 ${polling ? "animate-spin" : ""}`} />
          </Button>
        </div>
        {polling && (
          <p className="mt-2 text-xs text-[var(--accent)]">Auto-refreshing every 3 seconds…</p>
        )}
      </Card>

      {error && <Alert tone="danger">{error}</Alert>}

      {status && (
        <Card title="Transfer status">
          <div className="mb-4 flex flex-wrap gap-2">
            <Badge tone={sagaTone(status.saga_state)}>{humanSagaState(status.saga_state)}</Badge>
            <Badge tone="info">{humanSettlementStatus(status.settlement_status)}</Badge>
          </div>

          <Timeline current={status.saga_state} />

          <dl className="mt-6 space-y-2 border-t border-[var(--border)] pt-4 text-sm">
            <Dt label="Amount" value={formatSAR(status.amount)} />
            <Dt label="From → To" value={`${status.from_account_id} → ${status.to_account_id}`} />
            <Dt label="Beneficiary IBAN" value={status.beneficiary_iban} mono />
            <Dt label="Transaction ID" value={status.transaction_id} mono />
            <Dt label="Saga ID" value={status.saga_id} mono />
            {status.settlement_error && (
              <Dt label="Settlement error" value={status.settlement_error} />
            )}
            {status.failure_reason && (
              <Dt label="Failure" value={status.failure_reason} />
            )}
          </dl>
        </Card>
      )}
    </div>
  );
}

const STEPS = [
  { key: "COMPLIANCE_OK", label: "Compliance" },
  { key: "POSTED", label: "Ledger posted" },
  { key: "SETTLING", label: "SARIE settling" },
  { key: "COMPLETED", label: "Complete" },
];

const ORDER = ["COMPLIANCE_OK", "POSTED", "SETTLING", "COMPLETED", "COMPENSATED", "FAILED"];

function Timeline({ current }: { current: string }) {
  const idx = ORDER.indexOf(current);
  const failed = current === "FAILED" || current === "COMPENSATED";

  return (
    <ol className="relative space-y-0">
      {STEPS.map((step, i) => {
        const stepIdx = ORDER.indexOf(step.key);
        const done = idx >= stepIdx && !failed;
        const active = current === step.key || (current === "SETTLING" && step.key === "SETTLING");
        return (
          <li key={step.key} className="flex gap-4 pb-6 last:pb-0">
            <div className="flex flex-col items-center">
              <div
                className={`flex h-8 w-8 items-center justify-center rounded-full text-xs font-bold ring-2 ${
                  done
                    ? "bg-[var(--accent)] text-[#042f1f] ring-[var(--accent)]"
                    : active
                      ? "bg-[var(--accent-glow)] text-[var(--accent)] ring-[var(--accent)] pulse-dot"
                      : "bg-[var(--bg-elevated)] text-[var(--text-muted)] ring-[var(--border)]"
                }`}
              >
                {i + 1}
              </div>
              {i < STEPS.length - 1 && (
                <div className={`mt-1 w-0.5 flex-1 ${done ? "bg-[var(--accent)]" : "bg-[var(--border)]"}`} />
              )}
            </div>
            <div className="pt-1">
              <p className="font-medium">{step.label}</p>
            </div>
          </li>
        );
      })}
      {failed && (
        <Alert tone={current === "COMPENSATED" ? "warning" : "danger"}>
          {current === "COMPENSATED"
            ? "Settlement failed — funds were reversed automatically."
            : "Transfer failed. Check settlement error above."}
        </Alert>
      )}
    </ol>
  );
}

function Dt({ label, value, mono }: { label: string; value: string; mono?: boolean }) {
  return (
    <div className="flex flex-col gap-0.5 sm:flex-row sm:justify-between">
      <dt className="text-[var(--text-muted)]">{label}</dt>
      <dd className={mono ? "font-mono text-xs break-all text-right" : ""}>{value}</dd>
    </div>
  );
}

export default function TrackPage() {
  return (
    <Suspense fallback={<p className="text-[var(--text-muted)]">Loading…</p>}>
      <TrackContent />
    </Suspense>
  );
}
