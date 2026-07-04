"use client";

import { useState } from "react";
import Link from "next/link";
import { ArrowLeft, Building2, CheckCircle2, Plus } from "lucide-react";
import { api, ApiError } from "@/lib/api";
import { Alert, Button, Card, Divider, Input, Label, Select } from "@/components/ui";

const TYPES = ["LIABILITY", "ASSET", "EQUITY", "REVENUE", "EXPENSE"] as const;

const TYPE_DESCRIPTIONS: Record<string, string> = {
  LIABILITY: "Customer wallets and deposit accounts",
  ASSET: "Bank-owned assets and receivables",
  EQUITY: "Owner's equity and retained earnings",
  REVENUE: "Income from fees, interest, etc.",
  EXPENSE: "Operational costs and charges",
};

export default function NewAccountPage() {
  const [form, setForm] = useState({
    id: "",
    name: "",
    account_type: "LIABILITY" as (typeof TYPES)[number],
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError(null);
    setSuccess(null);
    try {
      const res = await api.createAccount(form);
      setSuccess(res.id);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Failed to create account");
    } finally {
      setLoading(false);
    }
  }

  if (success) {
    return (
      <div className="mx-auto max-w-lg space-y-6 animate-in">
        <div className="flex flex-col items-center py-10 text-center">
          <div className="flex h-16 w-16 items-center justify-center rounded-full bg-emerald-500/10 ring-2 ring-emerald-500/20">
            <CheckCircle2 className="h-8 w-8 text-emerald-400" />
          </div>
          <h2 className="mt-4 text-xl font-bold">Account Created</h2>
          <p className="mt-2 text-sm text-[var(--text-muted)]">
            Account <span className="font-mono text-[var(--accent)]">{success}</span> has been successfully created
          </p>
          <div className="mt-6 flex gap-3">
            <Link href={`/accounts?focus=${encodeURIComponent(success)}`}>
              <Button>View Account</Button>
            </Link>
            <Button
              variant="secondary"
              onClick={() => {
                setSuccess(null);
                setForm({ id: "", name: "", account_type: "LIABILITY" });
              }}
            >
              Create Another
            </Button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-lg space-y-6 animate-in">
      <header>
        <Link
          href="/accounts"
          className="mb-3 inline-flex items-center gap-1 text-sm text-[var(--text-muted)] hover:text-[var(--text)] transition-colors"
        >
          <ArrowLeft className="h-4 w-4" />
          Back to Accounts
        </Link>
        <div className="flex items-center gap-3">
          <div className="flex h-11 w-11 items-center justify-center rounded-xl bg-purple-500/10 text-purple-400">
            <Plus className="h-5 w-5" />
          </div>
          <div>
            <h1 className="text-xl font-bold tracking-tight">Create Account</h1>
            <p className="text-sm text-[var(--text-muted)]">
              Add a new ledger account to the banking system
            </p>
          </div>
        </div>
      </header>

      <Card>
        <form onSubmit={onSubmit} className="space-y-5">
          <div>
            <Label htmlFor="id" required>Account ID</Label>
            <Input
              id="id"
              value={form.id}
              onChange={(e) => setForm({ ...form, id: e.target.value })}
              placeholder="wallet-carol"
              required
            />
            <p className="mt-1.5 text-[11px] text-[var(--text-muted)]">
              Unique identifier. Use lowercase with hyphens (e.g., wallet-carol)
            </p>
          </div>

          <div>
            <Label htmlFor="name" required>Display Name</Label>
            <Input
              id="name"
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              placeholder="Carol's Wallet"
              required
            />
          </div>

          <Divider />

          <div>
            <Label htmlFor="type" required>Account Type</Label>
            <Select
              id="type"
              value={form.account_type}
              onChange={(e) =>
                setForm({ ...form, account_type: e.target.value as (typeof TYPES)[number] })
              }
            >
              {TYPES.map((t) => (
                <option key={t} value={t}>
                  {t}
                </option>
              ))}
            </Select>
            <p className="mt-1.5 text-[11px] text-[var(--text-muted)]">
              {TYPE_DESCRIPTIONS[form.account_type]}
            </p>
          </div>

          {error && <Alert tone="danger">{error}</Alert>}

          <Button type="submit" className="w-full" size="lg" disabled={loading || !form.id || !form.name}>
            {loading ? (
              <>
                <span className="h-4 w-4 animate-spin rounded-full border-2 border-white/30 border-t-white" />
                Creating…
              </>
            ) : (
              <>
                <Building2 className="h-4 w-4" />
                Create Account
              </>
            )}
          </Button>
        </form>
      </Card>
    </div>
  );
}
