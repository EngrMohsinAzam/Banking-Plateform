"use client";

import { useState } from "react";
import Link from "next/link";
import {
  ArrowRight,
  ArrowUpRight,
  CheckCircle2,
  ChevronLeft,
  CreditCard,
  Send,
  Shield,
  User,
} from "lucide-react";
import { api, ApiError } from "@/lib/api";
import type { TransferResponse } from "@/lib/types";
import {
  formatSAR,
  humanSagaState,
  humanSettlementStatus,
  newIdempotencyKey,
  sagaTone,
} from "@/lib/utils";
import {
  Alert,
  Badge,
  Button,
  Card,
  Divider,
  Input,
  Label,
  StepIndicator,
} from "@/components/ui";

type Step = "details" | "review" | "result";

export default function TransferPage() {
  const [step, setStep] = useState<Step>("details");
  const [form, setForm] = useState({
    from_account_id: "wallet-alice",
    to_account_id: "wallet-bob",
    amount: "100.00",
    beneficiary_iban: "SA0380000000608010167519",
    beneficiary_name: "Mohsin Azam",
    description: "P2P transfer",
  });
  const [idempotencyKey, setIdempotencyKey] = useState(newIdempotencyKey);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<TransferResponse | null>(null);

  const stepNames = ["Transfer Details", "Review & Confirm", "Complete"];
  const currentStepIndex = step === "details" ? 0 : step === "review" ? 1 : 2;

  async function onSubmit() {
    setLoading(true);
    setError(null);
    try {
      const res = await api.createTransfer(form, idempotencyKey);
      setResult(res);
      setStep("result");
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Transfer failed. Please try again.");
      setStep("details");
    } finally {
      setLoading(false);
    }
  }

  function resetForm() {
    setForm({
      from_account_id: "wallet-alice",
      to_account_id: "wallet-bob",
      amount: "100.00",
      beneficiary_iban: "SA0380000000608010167519",
      beneficiary_name: "Mohsin Azam",
      description: "P2P transfer",
    });
    setIdempotencyKey(newIdempotencyKey());
    setResult(null);
    setError(null);
    setStep("details");
  }

  return (
    <div className="mx-auto max-w-2xl space-y-6 animate-in">
      <header>
        <div className="flex items-center gap-3">
          <div className="flex h-11 w-11 items-center justify-center rounded-xl bg-[var(--accent-glow-strong)] text-[var(--accent)]">
            <Send className="h-5 w-5" />
          </div>
          <div>
            <h1 className="text-xl font-bold tracking-tight md:text-2xl">Transfer Funds</h1>
            <p className="text-sm text-[var(--text-muted)]">
              Send SAR securely between wallets via SARIE
            </p>
          </div>
        </div>
      </header>

      <div className="flex justify-center">
        <StepIndicator steps={stepNames} current={currentStepIndex} />
      </div>

      {error && (
        <Alert tone="danger" title="Transfer Failed">
          {error}
        </Alert>
      )}

      {/* Step 1: Details */}
      {step === "details" && (
        <Card className="animate-in">
          <form
            onSubmit={(e) => {
              e.preventDefault();
              setStep("review");
            }}
            className="space-y-5"
          >
            <div>
              <p className="mb-4 text-sm font-semibold text-[var(--text)]">Source & Destination</p>
              <div className="grid gap-4 sm:grid-cols-2">
                <div>
                  <Label htmlFor="from" required>From Account</Label>
                  <div className="relative">
                    <CreditCard className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-[var(--text-muted)]" />
                    <Input
                      id="from"
                      value={form.from_account_id}
                      onChange={(e) => setForm({ ...form, from_account_id: e.target.value })}
                      className="pl-10"
                      required
                    />
                  </div>
                </div>
                <div>
                  <Label htmlFor="to" required>To Account</Label>
                  <div className="relative">
                    <CreditCard className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-[var(--text-muted)]" />
                    <Input
                      id="to"
                      value={form.to_account_id}
                      onChange={(e) => setForm({ ...form, to_account_id: e.target.value })}
                      className="pl-10"
                      required
                    />
                  </div>
                </div>
              </div>
            </div>

            <Divider />

            <div>
              <Label htmlFor="amount" required>Amount (SAR)</Label>
              <div className="relative">
                <span className="absolute left-4 top-1/2 -translate-y-1/2 text-sm font-semibold text-[var(--text-muted)]">
                  SAR
                </span>
                <Input
                  id="amount"
                  type="text"
                  inputMode="decimal"
                  placeholder="0.00"
                  value={form.amount}
                  onChange={(e) => setForm({ ...form, amount: e.target.value })}
                  className="pl-14 text-lg font-semibold"
                  required
                />
              </div>
            </div>

            <Divider />

            <div>
              <p className="mb-4 text-sm font-semibold text-[var(--text)]">Beneficiary Information</p>
              <div className="space-y-4">
                <div>
                  <Label htmlFor="iban" required>Beneficiary IBAN</Label>
                  <Input
                    id="iban"
                    value={form.beneficiary_iban}
                    onChange={(e) => setForm({ ...form, beneficiary_iban: e.target.value })}
                    placeholder="SA..."
                    className="font-mono text-sm"
                    required
                  />
                  <p className="mt-1 text-[11px] text-[var(--text-muted)]">Saudi Arabia IBAN format: SA + 22 digits</p>
                </div>
                <div>
                  <Label htmlFor="name">Beneficiary Name</Label>
                  <div className="relative">
                    <User className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-[var(--text-muted)]" />
                    <Input
                      id="name"
                      value={form.beneficiary_name}
                      onChange={(e) => setForm({ ...form, beneficiary_name: e.target.value })}
                      className="pl-10"
                    />
                  </div>
                </div>
                <div>
                  <Label htmlFor="desc">Transfer Description</Label>
                  <Input
                    id="desc"
                    value={form.description}
                    onChange={(e) => setForm({ ...form, description: e.target.value })}
                    placeholder="Payment for..."
                  />
                </div>
              </div>
            </div>

            <Button type="submit" className="w-full" size="lg">
              Continue to Review
              <ArrowRight className="h-4 w-4" />
            </Button>
          </form>
        </Card>
      )}

      {/* Step 2: Review */}
      {step === "review" && (
        <div className="space-y-4 animate-in">
          <Card title="Review Transfer Details">
            <div className="space-y-0">
              <ReviewRow label="From Account" value={form.from_account_id} />
              <ReviewRow label="To Account" value={form.to_account_id} />
              <ReviewRow
                label="Amount"
                value={formatSAR(form.amount)}
                highlight
              />
              <ReviewRow label="Beneficiary IBAN" value={form.beneficiary_iban} mono />
              {form.beneficiary_name && (
                <ReviewRow label="Beneficiary Name" value={form.beneficiary_name} />
              )}
              {form.description && (
                <ReviewRow label="Description" value={form.description} />
              )}
              <ReviewRow label="Idempotency Key" value={idempotencyKey} mono />
            </div>
          </Card>

          <Card className="border-amber-500/20 bg-amber-500/5">
            <div className="flex gap-3">
              <Shield className="mt-0.5 h-5 w-5 shrink-0 text-amber-400" />
              <div>
                <p className="text-sm font-medium text-amber-300">Security Notice</p>
                <p className="mt-1 text-xs text-amber-300/80">
                  This transfer will undergo fraud screening, compliance verification, and KSA sanctions checking before funds are moved. Settlement via SARIE is processed by the background worker.
                </p>
              </div>
            </div>
          </Card>

          <div className="flex gap-3">
            <Button
              variant="secondary"
              className="flex-1"
              onClick={() => setStep("details")}
            >
              <ChevronLeft className="h-4 w-4" />
              Back
            </Button>
            <Button
              className="flex-1"
              onClick={onSubmit}
              disabled={loading}
              size="lg"
            >
              {loading ? (
                <>
                  <span className="h-4 w-4 animate-spin rounded-full border-2 border-white/30 border-t-white" />
                  Processing…
                </>
              ) : (
                <>
                  <Shield className="h-4 w-4" />
                  Confirm & Send {formatSAR(form.amount)}
                </>
              )}
            </Button>
          </div>
        </div>
      )}

      {/* Step 3: Result */}
      {step === "result" && result && (
        <div className="space-y-4 animate-in">
          <div className="flex flex-col items-center py-6 text-center">
            <div className="flex h-16 w-16 items-center justify-center rounded-full bg-emerald-500/10 ring-2 ring-emerald-500/20">
              <CheckCircle2 className="h-8 w-8 text-emerald-400" />
            </div>
            <h2 className="mt-4 text-xl font-bold">Transfer Submitted</h2>
            <p className="mt-1 text-sm text-[var(--text-muted)]">
              Your transfer is being processed through the settlement pipeline
            </p>
          </div>

          {result.replayed && (
            <Alert tone="warning" title="Idempotent Replay">
              This was a safe replay — the original transfer was returned without posting twice.
            </Alert>
          )}

          <Card>
            <div className="space-y-0">
              <ReviewRow label="Amount" value={formatSAR(result.amount)} highlight />
              <ReviewRow label="Transaction ID" value={result.transaction_id} mono />
              <ReviewRow
                label="Saga State"
                value={
                  <Badge tone={sagaTone(result.saga_state)} dot>
                    {humanSagaState(result.saga_state)}
                  </Badge>
                }
              />
              <ReviewRow
                label="Settlement"
                value={
                  <Badge tone="info" dot>
                    {humanSettlementStatus(result.settlement_status)}
                  </Badge>
                }
              />
              <ReviewRow label="From → To" value={`${result.from_account_id} → ${result.to_account_id}`} />
            </div>
          </Card>

          <div className="flex gap-3">
            <Link href={`/track?tx=${encodeURIComponent(result.transaction_id)}`} className="flex-1">
              <Button variant="secondary" className="w-full">
                <ArrowUpRight className="h-4 w-4" />
                Track Settlement
              </Button>
            </Link>
            <Button variant="ghost" className="flex-1" onClick={resetForm}>
              New Transfer
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}

function ReviewRow({
  label,
  value,
  mono,
  highlight,
}: {
  label: string;
  value: React.ReactNode;
  mono?: boolean;
  highlight?: boolean;
}) {
  return (
    <div className="flex items-center justify-between border-b border-[var(--border)]/50 py-3 last:border-0">
      <span className="text-sm text-[var(--text-muted)]">{label}</span>
      <span
        className={
          highlight
            ? "text-lg font-bold text-[var(--accent)]"
            : mono
              ? "max-w-[200px] truncate font-mono text-xs text-[var(--text-secondary)]"
              : "text-sm font-medium"
        }
      >
        {value}
      </span>
    </div>
  );
}
