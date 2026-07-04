"use client";

import { useState } from "react";
import Link from "next/link";
import { Send } from "lucide-react";
import { api, ApiError } from "@/lib/api";
import type { TransferResponse } from "@/lib/types";
import { formatSAR, humanSagaState, humanSettlementStatus, newIdempotencyKey, sagaTone } from "@/lib/utils";
import { Alert, Badge, Button, Card, Input, Label } from "@/components/ui";

export default function TransferPage() {
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

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError(null);
    setResult(null);
    try {
      const res = await api.createTransfer(form, idempotencyKey);
      setResult(res);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Transfer failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="mx-auto max-w-2xl space-y-6 animate-in">
      <header>
        <h1 className="flex items-center gap-2 text-2xl font-bold">
          <Send className="h-7 w-7 text-[var(--accent)]" />
          Send money
        </h1>
        <p className="mt-1 text-[var(--text-muted)]">
          Transfer SAR between wallets. Every request is idempotent and screened for fraud & compliance.
        </p>
      </header>

      <Card>
        <form onSubmit={onSubmit} className="space-y-4">
          <div className="grid gap-4 sm:grid-cols-2">
            <div>
              <Label htmlFor="from">From wallet</Label>
              <Input
                id="from"
                value={form.from_account_id}
                onChange={(e) => setForm({ ...form, from_account_id: e.target.value })}
                required
              />
            </div>
            <div>
              <Label htmlFor="to">To wallet</Label>
              <Input
                id="to"
                value={form.to_account_id}
                onChange={(e) => setForm({ ...form, to_account_id: e.target.value })}
                required
              />
            </div>
          </div>

          <div>
            <Label htmlFor="amount">Amount (SAR)</Label>
            <Input
              id="amount"
              type="text"
              inputMode="decimal"
              placeholder="100.00"
              value={form.amount}
              onChange={(e) => setForm({ ...form, amount: e.target.value })}
              required
            />
          </div>

          <div>
            <Label htmlFor="iban">Beneficiary IBAN</Label>
            <Input
              id="iban"
              value={form.beneficiary_iban}
              onChange={(e) => setForm({ ...form, beneficiary_iban: e.target.value })}
              required
            />
          </div>

          <div>
            <Label htmlFor="name">Beneficiary name</Label>
            <Input
              id="name"
              value={form.beneficiary_name}
              onChange={(e) => setForm({ ...form, beneficiary_name: e.target.value })}
            />
          </div>

          <div>
            <Label htmlFor="desc">Description</Label>
            <Input
              id="desc"
              value={form.description}
              onChange={(e) => setForm({ ...form, description: e.target.value })}
            />
          </div>

          <div>
            <Label htmlFor="idem">Idempotency key</Label>
            <div className="flex gap-2">
              <Input id="idem" value={idempotencyKey} readOnly className="font-mono text-xs" />
              <Button type="button" variant="secondary" onClick={() => setIdempotencyKey(newIdempotencyKey())}>
                New
              </Button>
            </div>
            <p className="mt-1 text-xs text-[var(--text-muted)]">
              Reusing this key safely retries the same transfer without double-charging.
            </p>
          </div>

          {error && <Alert tone="danger">{error}</Alert>}

          <Button type="submit" className="w-full" disabled={loading}>
            {loading ? "Processing…" : `Send ${formatSAR(form.amount)}`}
          </Button>
        </form>
      </Card>

      {result && (
        <Card title="Transfer accepted" className="animate-in">
          {result.replayed && (
            <Alert tone="warning" >
              This was a safe replay — the original transfer was returned without posting twice.
            </Alert>
          )}
          <dl className="mt-4 space-y-3 text-sm">
            <Row label="Amount" value={formatSAR(result.amount)} />
            <Row label="Transaction ID" value={result.transaction_id} mono />
            <Row label="Saga" value={
              <Badge tone={sagaTone(result.saga_state)}>{humanSagaState(result.saga_state)}</Badge>
            } />
            <Row label="Settlement" value={
              <Badge tone="info">{humanSettlementStatus(result.settlement_status)}</Badge>
            } />
          </dl>
          <Link href={`/track?tx=${encodeURIComponent(result.transaction_id)}`} className="mt-4 block">
            <Button variant="secondary" className="w-full">
              Track settlement progress
            </Button>
          </Link>
        </Card>
      )}
    </div>
  );
}

function Row({ label, value, mono }: { label: string; value: React.ReactNode; mono?: boolean }) {
  return (
    <div className="flex flex-col gap-0.5 sm:flex-row sm:justify-between">
      <dt className="text-[var(--text-muted)]">{label}</dt>
      <dd className={mono ? "font-mono text-xs break-all" : "font-medium"}>{value}</dd>
    </div>
  );
}
