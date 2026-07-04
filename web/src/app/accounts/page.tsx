"use client";

import { Suspense, useEffect, useState } from "react";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import {
  ArrowDownLeft,
  ArrowUpRight,
  Building2,
  CreditCard,
  Loader2,
  Plus,
  Search,
} from "lucide-react";
import { api, ApiError } from "@/lib/api";
import type { LedgerEntry } from "@/lib/types";
import { formatSAR, cn } from "@/lib/utils";
import { Alert, Badge, Button, Card, EmptyState, Input, MetricCard } from "@/components/ui";

const DEFAULT_IDS = ["wallet-alice", "wallet-bob"];

function AccountsContent() {
  const params = useSearchParams();
  const focus = params.get("focus") ?? DEFAULT_IDS[0];
  const [selected, setSelected] = useState(focus);
  const [balance, setBalance] = useState<string | null>(null);
  const [entries, setEntries] = useState<LedgerEntry[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function load() {
      setError(null);
      setLoading(true);
      try {
        const b = await api.getBalance(selected);
        setBalance(b.balance);
        const e = await api.getEntries(selected);
        setEntries(e.entries ?? []);
      } catch (err) {
        setBalance(null);
        setEntries([]);
        setError(err instanceof ApiError ? err.message : "Failed to load account");
      } finally {
        setLoading(false);
      }
    }
    load();
  }, [selected]);

  const debits = entries.filter((e) => e.side === "DEBIT");
  const credits = entries.filter((e) => e.side === "CREDIT");

  return (
    <div className="mx-auto max-w-5xl space-y-6 animate-in">
      <header className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex items-center gap-3">
          <div className="flex h-11 w-11 items-center justify-center rounded-xl bg-purple-500/10 text-purple-400">
            <Building2 className="h-5 w-5" />
          </div>
          <div>
            <h1 className="text-xl font-bold tracking-tight md:text-2xl">Accounts</h1>
            <p className="text-sm text-[var(--text-muted)]">
              View wallet balances and the append-only double-entry ledger
            </p>
          </div>
        </div>
        <Link href="/accounts/new">
          <Button variant="secondary" size="sm">
            <Plus className="h-4 w-4" />
            New Account
          </Button>
        </Link>
      </header>

      {/* Account Selector */}
      <Card noPadding>
        <div className="flex items-center gap-2 border-b border-[var(--border)] px-5 py-3">
          <p className="text-xs font-semibold uppercase tracking-wider text-[var(--text-muted)]">
            Select Account
          </p>
        </div>
        <div className="flex flex-wrap items-center gap-2 p-4">
          {DEFAULT_IDS.map((id) => (
            <button
              key={id}
              onClick={() => setSelected(id)}
              className={cn(
                "flex items-center gap-2.5 rounded-xl px-4 py-2.5 text-sm font-medium transition-all",
                selected === id
                  ? "bg-[var(--accent-glow-strong)] text-[var(--accent)] ring-1 ring-[var(--accent)]/20 shadow-sm"
                  : "bg-[var(--bg-elevated)] text-[var(--text-muted)] ring-1 ring-[var(--border)] hover:bg-[var(--bg-card)] hover:text-[var(--text)]",
              )}
            >
              <CreditCard className="h-4 w-4" />
              {id === "wallet-alice" ? "Alice" : id === "wallet-bob" ? "Bob" : id}
              <span className="font-mono text-[10px] opacity-60">{id}</span>
            </button>
          ))}
          <AccountSearchInput
            onSelect={(id) => setSelected(id)}
          />
        </div>
      </Card>

      {error && <Alert tone="danger">{error}</Alert>}

      {/* Balance & Stats */}
      <div className="grid gap-4 sm:grid-cols-3">
        <MetricCard
          label="Available Balance"
          value={loading ? "..." : balance != null ? formatSAR(balance) : "—"}
          subValue={selected}
          icon={<CreditCard className="h-5 w-5" />}
          className="sm:col-span-1"
        />
        <MetricCard
          label="Total Debits"
          value={debits.length.toString()}
          subValue={
            debits.length > 0
              ? formatSAR(
                  debits.reduce((s, e) => s + parseFloat(e.amount), 0).toFixed(2),
                )
              : "—"
          }
          icon={<ArrowUpRight className="h-5 w-5" />}
        />
        <MetricCard
          label="Total Credits"
          value={credits.length.toString()}
          subValue={
            credits.length > 0
              ? formatSAR(
                  credits.reduce((s, e) => s + parseFloat(e.amount), 0).toFixed(2),
                )
              : "—"
          }
          icon={<ArrowDownLeft className="h-5 w-5" />}
        />
      </div>

      {/* Ledger Entries Table */}
      <Card
        title="Journal Entries"
        subtitle="Append-only immutable ledger records"
        noPadding
      >
        {loading ? (
          <div className="flex items-center justify-center py-16">
            <Loader2 className="h-6 w-6 animate-spin text-[var(--text-muted)]" />
          </div>
        ) : entries.length === 0 ? (
          <div className="px-6 pb-6">
            <EmptyState
              icon={<Building2 className="h-7 w-7" />}
              title="No entries yet"
              description="This account has no ledger entries. Make a transfer to see entries appear here."
              action={
                <Link href="/transfer">
                  <Button size="sm">Make a Transfer</Button>
                </Link>
              }
            />
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-sm">
              <thead>
                <tr className="border-b border-[var(--border)] text-[var(--text-muted)]">
                  <th className="px-6 py-3 text-xs font-semibold uppercase tracking-wider">#</th>
                  <th className="px-6 py-3 text-xs font-semibold uppercase tracking-wider">Entry ID</th>
                  <th className="px-6 py-3 text-xs font-semibold uppercase tracking-wider">Type</th>
                  <th className="px-6 py-3 text-xs font-semibold uppercase tracking-wider text-right">Amount</th>
                </tr>
              </thead>
              <tbody>
                {entries.map((e, i) => (
                  <tr
                    key={e.id}
                    className="border-b border-[var(--border)]/30 transition-colors hover:bg-[var(--bg-elevated)]"
                  >
                    <td className="px-6 py-4 text-xs text-[var(--text-muted)]">{i + 1}</td>
                    <td className="px-6 py-4 font-mono text-xs text-[var(--text-secondary)]">
                      {e.id}
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-2">
                        <div
                          className={cn(
                            "flex h-7 w-7 items-center justify-center rounded-lg",
                            e.side === "DEBIT"
                              ? "bg-amber-500/10 text-amber-400"
                              : "bg-emerald-500/10 text-emerald-400",
                          )}
                        >
                          {e.side === "DEBIT" ? (
                            <ArrowUpRight className="h-3.5 w-3.5" />
                          ) : (
                            <ArrowDownLeft className="h-3.5 w-3.5" />
                          )}
                        </div>
                        <Badge
                          tone={e.side === "DEBIT" ? "warning" : "success"}
                          size="sm"
                        >
                          {e.side}
                        </Badge>
                      </div>
                    </td>
                    <td className="px-6 py-4 text-right">
                      <span
                        className={cn(
                          "font-semibold tabular-nums",
                          e.side === "DEBIT" ? "text-amber-400" : "text-emerald-400",
                        )}
                      >
                        {e.side === "DEBIT" ? "−" : "+"} {formatSAR(e.amount)}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Card>
    </div>
  );
}

function AccountSearchInput({
  onSelect,
}: {
  onSelect: (id: string) => void;
}) {
  const [value, setValue] = useState("");
  return (
    <form
      className="flex items-center gap-2"
      onSubmit={(e) => {
        e.preventDefault();
        if (value.trim()) {
          onSelect(value.trim());
          setValue("");
        }
      }}
    >
      <div className="relative">
        <Search className="absolute left-2.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-[var(--text-muted)]" />
        <input
          className="h-9 w-44 rounded-xl border border-[var(--border)] bg-[var(--bg-elevated)] pl-8 pr-3 text-xs font-mono text-[var(--text)] placeholder:text-[var(--text-muted)]/60 focus:border-[var(--accent)]/40 focus:outline-none focus:ring-2 focus:ring-[var(--accent)]/10"
          placeholder="Other account ID..."
          value={value}
          onChange={(e) => setValue(e.target.value)}
        />
      </div>
      <Button type="submit" variant="ghost" size="sm">
        Load
      </Button>
    </form>
  );
}

export default function AccountsPage() {
  return (
    <Suspense
      fallback={
        <div className="flex items-center justify-center py-20">
          <Loader2 className="h-6 w-6 animate-spin text-[var(--text-muted)]" />
        </div>
      }
    >
      <AccountsContent />
    </Suspense>
  );
}
