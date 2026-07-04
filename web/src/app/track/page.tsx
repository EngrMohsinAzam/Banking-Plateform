"use client";

import { Suspense, useCallback, useEffect, useState } from "react";
import { useSearchParams } from "next/navigation";
import {
  AlertTriangle,
  ArrowRight,
  CheckCircle2,
  Clock,
  Copy,
  Loader2,
  RefreshCw,
  Search,
  Shield,
  XCircle,
  Zap,
} from "lucide-react";
import { api, ApiError } from "@/lib/api";
import type { TransferStatus } from "@/lib/types";
import {
  formatSAR,
  humanSagaState,
  humanSettlementStatus,
  sagaTone,
} from "@/lib/utils";
import { Alert, Badge, Button, Card, Input, Label } from "@/components/ui";
import { cn } from "@/lib/utils";

const TERMINAL = new Set(["COMPLETED", "COMPENSATED", "FAILED"]);

function TrackContent() {
  const params = useSearchParams();
  const [mode, setMode] = useState<"tx" | "key">("tx");
  const [query, setQuery] = useState(params.get("tx") ?? "");
  const [status, setStatus] = useState<TransferStatus | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [polling, setPolling] = useState(false);
  const [copied, setCopied] = useState<string | null>(null);

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
      setError(err instanceof ApiError ? err.message : "Transfer not found");
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

  function copyToClipboard(text: string, field: string) {
    navigator.clipboard.writeText(text);
    setCopied(field);
    setTimeout(() => setCopied(null), 2000);
  }

  return (
    <div className="mx-auto max-w-3xl space-y-6 animate-in">
      <header>
        <div className="flex items-center gap-3">
          <div className="flex h-11 w-11 items-center justify-center rounded-xl bg-blue-500/10 text-blue-400">
            <Search className="h-5 w-5" />
          </div>
          <div>
            <h1 className="text-xl font-bold tracking-tight md:text-2xl">Track Transfer</h1>
            <p className="text-sm text-[var(--text-muted)]">
              Monitor your transfer through each stage of the settlement pipeline
            </p>
          </div>
        </div>
      </header>

      <Card>
        <div className="mb-4 flex gap-1 rounded-lg bg-[var(--bg-elevated)] p-1 ring-1 ring-[var(--border)]">
          <button
            onClick={() => setMode("tx")}
            className={cn(
              "flex-1 rounded-md px-3 py-2 text-xs font-semibold transition-all",
              mode === "tx"
                ? "bg-[var(--bg-card)] text-[var(--text)] shadow-sm"
                : "text-[var(--text-muted)] hover:text-[var(--text)]",
            )}
          >
            Transaction ID
          </button>
          <button
            onClick={() => setMode("key")}
            className={cn(
              "flex-1 rounded-md px-3 py-2 text-xs font-semibold transition-all",
              mode === "key"
                ? "bg-[var(--bg-card)] text-[var(--text)] shadow-sm"
                : "text-[var(--text-muted)] hover:text-[var(--text)]",
            )}
          >
            Idempotency Key
          </button>
        </div>
        <Label htmlFor="q">{mode === "tx" ? "Transaction ID" : "Idempotency Key"}</Label>
        <div className="flex gap-2">
          <Input
            id="q"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder={mode === "tx" ? "Enter transaction ID..." : "Enter idempotency key..."}
            className="font-mono text-xs"
          />
          <Button
            type="button"
            variant="secondary"
            onClick={() => fetchStatus()}
            disabled={!query.trim()}
          >
            <RefreshCw className={`h-4 w-4 ${polling ? "animate-spin" : ""}`} />
          </Button>
        </div>
        {polling && (
          <div className="mt-3 flex items-center gap-2 text-xs text-[var(--accent)]">
            <Loader2 className="h-3 w-3 animate-spin" />
            Live tracking — auto-refreshing every 3 seconds
          </div>
        )}
      </Card>

      {error && (
        <Alert tone="danger" title="Lookup Failed">
          {error}
        </Alert>
      )}

      {status && (
        <div className="space-y-4 animate-in">
          {/* Status Banner */}
          <StatusBanner state={status.saga_state} amount={status.amount} />

          {/* Pipeline Timeline */}
          <Card title="Settlement Pipeline">
            <PipelineTimeline current={status.saga_state} />
          </Card>

          {/* Transfer Details */}
          <Card title="Transfer Details">
            <div className="space-y-0">
              <DetailRow label="Amount" value={formatSAR(status.amount)} highlight />
              <DetailRow
                label="Saga State"
                value={
                  <Badge tone={sagaTone(status.saga_state)} dot size="md">
                    {humanSagaState(status.saga_state)}
                  </Badge>
                }
              />
              <DetailRow
                label="Settlement"
                value={
                  <Badge tone={status.settlement_status === "SETTLED" ? "success" : status.settlement_status === "FAILED" ? "danger" : "info"} dot size="md">
                    {humanSettlementStatus(status.settlement_status)}
                  </Badge>
                }
              />
              <DetailRow
                label="Route"
                value={
                  <span className="flex items-center gap-2 text-sm">
                    <span className="font-mono text-xs">{status.from_account_id}</span>
                    <ArrowRight className="h-3 w-3 text-[var(--text-muted)]" />
                    <span className="font-mono text-xs">{status.to_account_id}</span>
                  </span>
                }
              />
              <DetailRow
                label="Beneficiary IBAN"
                value={status.beneficiary_iban}
                mono
                copyable
                onCopy={() => copyToClipboard(status.beneficiary_iban, "iban")}
                copied={copied === "iban"}
              />
              <DetailRow
                label="Transaction ID"
                value={status.transaction_id}
                mono
                copyable
                onCopy={() => copyToClipboard(status.transaction_id, "tx")}
                copied={copied === "tx"}
              />
              <DetailRow
                label="Saga ID"
                value={status.saga_id}
                mono
                copyable
                onCopy={() => copyToClipboard(status.saga_id, "saga")}
                copied={copied === "saga"}
              />
              {status.created_at && (
                <DetailRow
                  label="Created"
                  value={new Date(status.created_at).toLocaleString()}
                />
              )}
              {status.updated_at && (
                <DetailRow
                  label="Last Updated"
                  value={new Date(status.updated_at).toLocaleString()}
                />
              )}
            </div>
          </Card>

          {/* Error Details */}
          {(status.settlement_error || status.failure_reason) && (
            <Alert
              tone="danger"
              title="Error Details"
            >
              {status.settlement_error && (
                <p className="text-sm">Settlement: {status.settlement_error}</p>
              )}
              {status.failure_reason && (
                <p className="text-sm">Failure: {status.failure_reason}</p>
              )}
            </Alert>
          )}
        </div>
      )}
    </div>
  );
}

function StatusBanner({ state, amount }: { state: string; amount: string }) {
  const config = {
    COMPLETED: {
      icon: <CheckCircle2 className="h-6 w-6" />,
      title: "Transfer Complete",
      desc: `${formatSAR(amount)} successfully settled via SARIE`,
      bg: "from-emerald-950/40 to-[var(--bg-card)] border-emerald-500/20",
      iconColor: "text-emerald-400",
    },
    FAILED: {
      icon: <XCircle className="h-6 w-6" />,
      title: "Transfer Failed",
      desc: "The transfer could not be completed",
      bg: "from-red-950/40 to-[var(--bg-card)] border-red-500/20",
      iconColor: "text-red-400",
    },
    COMPENSATED: {
      icon: <AlertTriangle className="h-6 w-6" />,
      title: "Transfer Reversed",
      desc: "Settlement failed — funds were automatically reversed",
      bg: "from-amber-950/40 to-[var(--bg-card)] border-amber-500/20",
      iconColor: "text-amber-400",
    },
  }[state] ?? {
    icon: <Clock className="h-6 w-6" />,
    title: "Processing Transfer",
    desc: `${formatSAR(amount)} is being processed through the pipeline`,
    bg: "from-blue-950/40 to-[var(--bg-card)] border-blue-500/20",
    iconColor: "text-blue-400",
  };

  return (
    <div className={`flex items-center gap-4 rounded-2xl border bg-gradient-to-r p-5 ${config.bg}`}>
      <div className={config.iconColor}>{config.icon}</div>
      <div>
        <p className="font-semibold">{config.title}</p>
        <p className="text-sm text-[var(--text-muted)]">{config.desc}</p>
      </div>
    </div>
  );
}

const PIPELINE_STEPS = [
  { key: "FRAUD_OK", label: "Fraud Check", desc: "Limits & sanctions", icon: Shield },
  { key: "COMPLIANCE_OK", label: "Compliance", desc: "KYC & regulations", icon: Zap },
  { key: "POSTED", label: "Ledger Posted", desc: "Double-entry recorded", icon: CheckCircle2 },
  { key: "SETTLING", label: "SARIE Settlement", desc: "Bank network transfer", icon: ArrowRight },
  { key: "COMPLETED", label: "Complete", desc: "Funds delivered", icon: CheckCircle2 },
];

const STATE_ORDER = ["STARTED", "FRAUD_OK", "COMPLIANCE_OK", "POSTED", "SETTLING", "COMPLETED"];

function PipelineTimeline({ current }: { current: string }) {
  const currentIdx = STATE_ORDER.indexOf(current);
  const failed = current === "FAILED" || current === "COMPENSATED";

  return (
    <div className="space-y-0">
      {PIPELINE_STEPS.map((step, i) => {
        const stepIdx = STATE_ORDER.indexOf(step.key);
        const done = !failed && currentIdx >= stepIdx;
        const active = current === step.key;
        const Icon = step.icon;

        return (
          <div key={step.key} className="flex gap-4">
            <div className="flex flex-col items-center">
              <div
                className={cn(
                  "flex h-10 w-10 items-center justify-center rounded-xl text-sm font-bold transition-all",
                  done
                    ? "bg-[var(--accent)] text-white shadow-md shadow-emerald-900/30"
                    : active
                      ? "bg-[var(--accent-glow-strong)] text-[var(--accent)] ring-2 ring-[var(--accent)]/30 pulse-dot"
                      : failed && stepIdx > currentIdx
                        ? "bg-red-500/10 text-red-400/50 ring-1 ring-red-500/20"
                        : "bg-[var(--bg-surface)] text-[var(--text-muted)] ring-1 ring-[var(--border)]",
                )}
              >
                {done ? (
                  <CheckCircle2 className="h-5 w-5" />
                ) : active ? (
                  <Loader2 className="h-5 w-5 animate-spin" />
                ) : (
                  <Icon className="h-5 w-5" />
                )}
              </div>
              {i < PIPELINE_STEPS.length - 1 && (
                <div
                  className={cn(
                    "my-1 w-0.5 flex-1 min-h-[24px]",
                    done ? "bg-[var(--accent)]" : "bg-[var(--border)]",
                  )}
                />
              )}
            </div>
            <div className={cn("pb-6 pt-2", i === PIPELINE_STEPS.length - 1 && "pb-0")}>
              <p className={cn("text-sm font-semibold", done || active ? "text-[var(--text)]" : "text-[var(--text-muted)]")}>
                {step.label}
              </p>
              <p className="text-xs text-[var(--text-muted)]">{step.desc}</p>
            </div>
          </div>
        );
      })}

      {failed && (
        <div className="mt-4">
          <Alert tone={current === "COMPENSATED" ? "warning" : "danger"}>
            {current === "COMPENSATED"
              ? "Settlement failed — funds were reversed automatically via compensation saga."
              : "Transfer failed during processing. Check error details below."}
          </Alert>
        </div>
      )}
    </div>
  );
}

function DetailRow({
  label,
  value,
  mono,
  highlight,
  copyable,
  onCopy,
  copied,
}: {
  label: string;
  value: React.ReactNode;
  mono?: boolean;
  highlight?: boolean;
  copyable?: boolean;
  onCopy?: () => void;
  copied?: boolean;
}) {
  return (
    <div className="flex items-center justify-between border-b border-[var(--border)]/50 py-3 last:border-0">
      <span className="text-sm text-[var(--text-muted)]">{label}</span>
      <div className="flex items-center gap-2">
        <span
          className={
            highlight
              ? "text-lg font-bold text-[var(--accent)]"
              : mono
                ? "max-w-[180px] truncate font-mono text-xs text-[var(--text-secondary)] sm:max-w-[280px]"
                : "text-sm font-medium"
          }
        >
          {value}
        </span>
        {copyable && onCopy && (
          <button
            onClick={onCopy}
            className="flex h-6 w-6 items-center justify-center rounded-md text-[var(--text-muted)] hover:bg-[var(--bg-surface)] hover:text-[var(--text)]"
          >
            {copied ? (
              <CheckCircle2 className="h-3 w-3 text-[var(--accent)]" />
            ) : (
              <Copy className="h-3 w-3" />
            )}
          </button>
        )}
      </div>
    </div>
  );
}

export default function TrackPage() {
  return (
    <Suspense
      fallback={
        <div className="flex items-center justify-center py-20">
          <Loader2 className="h-6 w-6 animate-spin text-[var(--text-muted)]" />
        </div>
      }
    >
      <TrackContent />
    </Suspense>
  );
}
