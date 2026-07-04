"use client";

import { Suspense, useEffect, useState } from "react";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { Building2, Plus } from "lucide-react";
import { api, ApiError } from "@/lib/api";
import type { LedgerEntry } from "@/lib/types";
import { formatSAR } from "@/lib/utils";
import { Alert, Button, Card } from "@/components/ui";

const DEFAULT_IDS = ["wallet-alice", "wallet-bob"];

function AccountsContent() {
  const params = useSearchParams();
  const focus = params.get("focus") ?? DEFAULT_IDS[0];
  const [selected, setSelected] = useState(focus);
  const [balance, setBalance] = useState<string | null>(null);
  const [entries, setEntries] = useState<LedgerEntry[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function load() {
      setError(null);
      try {
        const b = await api.getBalance(selected);
        setBalance(b.balance);
        const e = await api.getEntries(selected);
        setEntries(e.entries ?? []);
      } catch (err) {
        setBalance(null);
        setEntries([]);
        setError(err instanceof ApiError ? err.message : "Failed to load account");
      }
    }
    load();
  }, [selected]);

  return (
    <div className="mx-auto max-w-4xl space-y-6 animate-in">
      <header className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <h1 className="flex items-center gap-2 text-2xl font-bold">
            <Building2 className="h-7 w-7 text-[var(--accent)]" />
            Accounts
          </h1>
          <p className="mt-1 text-[var(--text-muted)]">
            View wallet balances and journal entries from the append-only ledger.
          </p>
        </div>
        <Link href="/accounts/new">
          <Button variant="secondary">
            <Plus className="h-4 w-4" /> New account
          </Button>
        </Link>
      </header>

      <div className="flex flex-wrap gap-2">
        {DEFAULT_IDS.map((id) => (
          <Button
            key={id}
            variant={selected === id ? "primary" : "secondary"}
            onClick={() => setSelected(id)}
          >
            {id}
          </Button>
        ))}
        <InputAccountPicker value={selected} onChange={setSelected} />
      </div>

      {error && <Alert tone="danger">{error}</Alert>}

      <Card title="Balance">
        <p className="text-4xl font-bold tracking-tight">
          {balance != null ? formatSAR(balance) : "—"}
        </p>
        <p className="mt-1 font-mono text-xs text-[var(--text-muted)]">{selected}</p>
      </Card>

      <Card title="Journal entries" subtitle="Append-only ledger lines for this account">
        {entries.length === 0 ? (
          <p className="text-sm text-[var(--text-muted)]">No entries yet.</p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-sm">
              <thead>
                <tr className="border-b border-[var(--border)] text-[var(--text-muted)]">
                  <th className="pb-2 pr-4 font-medium">Entry</th>
                  <th className="pb-2 pr-4 font-medium">Side</th>
                  <th className="pb-2 font-medium text-right">Amount</th>
                </tr>
              </thead>
              <tbody>
                {entries.map((e) => (
                  <tr key={e.id} className="border-b border-[var(--border)]/50">
                    <td className="py-2.5 pr-4 font-mono text-xs">{e.id}</td>
                    <td className="py-2.5 pr-4">
                      <span
                        className={
                          e.side === "DEBIT"
                            ? "text-amber-400"
                            : "text-emerald-400"
                        }
                      >
                        {e.side}
                      </span>
                    </td>
                    <td className="py-2.5 text-right font-medium">{formatSAR(e.amount)}</td>
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

function InputAccountPicker({
  value,
  onChange,
}: {
  value: string;
  onChange: (v: string) => void;
}) {
  const [custom, setCustom] = useState("");
  return (
    <form
      className="flex gap-2"
      onSubmit={(e) => {
        e.preventDefault();
        if (custom.trim()) onChange(custom.trim());
      }}
    >
      <input
        className="rounded-xl border border-[var(--border)] bg-[var(--bg-elevated)] px-3 py-2 text-xs font-mono"
        placeholder="Other account id…"
        value={custom}
        onChange={(e) => setCustom(e.target.value)}
      />
      <Button type="submit" variant="ghost">
        Load
      </Button>
    </form>
  );
}

export default function AccountsPage() {
  return (
    <Suspense fallback={<p className="text-[var(--text-muted)]">Loading…</p>}>
      <AccountsContent />
    </Suspense>
  );
}
