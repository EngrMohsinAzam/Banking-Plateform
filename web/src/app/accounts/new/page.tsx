"use client";

import { useState } from "react";
import Link from "next/link";
import { api, ApiError } from "@/lib/api";
import { Alert, Button, Card, Input, Label } from "@/components/ui";

const TYPES = ["LIABILITY", "ASSET", "EQUITY", "REVENUE", "EXPENSE"] as const;

export default function NewAccountPage() {
  const [form, setForm] = useState({
    id: "wallet-carol",
    name: "Carol Wallet",
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
      setSuccess(`Account ${res.id} created successfully.`);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Failed to create account");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="mx-auto max-w-lg space-y-6 animate-in">
      <header>
        <h1 className="text-2xl font-bold">Create account</h1>
        <p className="mt-1 text-[var(--text-muted)]">
          Add a new ledger account. Wallet transfers use LIABILITY accounts.
        </p>
      </header>

      <Card>
        <form onSubmit={onSubmit} className="space-y-4">
          <div>
            <Label htmlFor="id">Account ID</Label>
            <Input
              id="id"
              value={form.id}
              onChange={(e) => setForm({ ...form, id: e.target.value })}
              required
            />
          </div>
          <div>
            <Label htmlFor="name">Display name</Label>
            <Input
              id="name"
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              required
            />
          </div>
          <div>
            <Label htmlFor="type">Account type</Label>
            <select
              id="type"
              className="w-full rounded-xl border border-[var(--border)] bg-[var(--bg-elevated)] px-3 py-2.5 text-sm"
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
            </select>
          </div>

          {error && <Alert tone="danger">{error}</Alert>}
          {success && <Alert tone="success">{success}</Alert>}

          <Button type="submit" className="w-full" disabled={loading}>
            {loading ? "Creating…" : "Create account"}
          </Button>
        </form>
      </Card>

      <Link href="/accounts" className="text-sm text-[var(--accent)] hover:underline">
        ← Back to accounts
      </Link>
    </div>
  );
}
