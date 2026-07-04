"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Activity, ArrowRight, Wallet } from "lucide-react";
import { api } from "@/lib/api";
import { formatSAR } from "@/lib/utils";
import { Alert, Badge, Button, Card } from "@/components/ui";

const DEMO_ACCOUNTS = ["wallet-alice", "wallet-bob"];

export default function DashboardPage() {
  const [health, setHealth] = useState<"loading" | "up" | "down">("loading");
  const [ready, setReady] = useState<"loading" | "up" | "down">("loading");
  const [balances, setBalances] = useState<Record<string, string>>({});
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function load() {
      setError(null);
      try {
        await api.health();
        setHealth("up");
      } catch {
        setHealth("down");
      }
      try {
        await api.ready();
        setReady("up");
      } catch {
        setReady("down");
      }
      const map: Record<string, string> = {};
      for (const id of DEMO_ACCOUNTS) {
        try {
          const b = await api.getBalance(id);
          map[id] = b.balance;
        } catch {
          map[id] = "—";
        }
      }
      setBalances(map);
    }
    load();
  }, []);

  return (
    <div className="mx-auto max-w-5xl space-y-6 animate-in">
      <header>
        <h1 className="text-2xl font-bold tracking-tight md:text-3xl">Good to see you</h1>
        <p className="mt-1 text-[var(--text-muted)]">
          Your Saudi wallet dashboard — balances, transfers, and settlement status in one place.
        </p>
      </header>

      {error && <Alert tone="danger">{error}</Alert>}

      <div className="grid gap-4 sm:grid-cols-2">
        <Card title="API status" subtitle="Live connection to the banking core">
          <div className="flex flex-wrap gap-3">
            <StatusPill label="API" status={health} />
            <StatusPill label="Postgres + Redis" status={ready} />
          </div>
          {(health === "down" || ready === "down") && (
            <p className="mt-3 text-sm text-[var(--text-muted)]">
              Start the backend: <code className="text-[var(--accent)]">go run ./cmd/api</code> and{" "}
              <code className="text-[var(--accent)]">docker compose up -d</code>
            </p>
          )}
        </Card>

        <Card title="Quick actions">
          <div className="flex flex-col gap-2">
            <Link href="/transfer">
              <Button className="w-full">
                Send money <ArrowRight className="h-4 w-4" />
              </Button>
            </Link>
            <Link href="/track">
              <Button variant="secondary" className="w-full">
                Track a transfer
              </Button>
            </Link>
          </div>
        </Card>
      </div>

      <section>
        <h2 className="mb-3 flex items-center gap-2 text-lg font-semibold">
          <Wallet className="h-5 w-5 text-[var(--accent)]" />
          Wallet balances
        </h2>
        <div className="grid gap-4 sm:grid-cols-2">
          {DEMO_ACCOUNTS.map((id) => (
            <Card key={id} className="relative overflow-hidden">
              <div className="absolute right-0 top-0 h-24 w-24 translate-x-6 -translate-y-6 rounded-full bg-[var(--accent-glow)] blur-2xl" />
              <p className="text-sm text-[var(--text-muted)]">{humanName(id)}</p>
              <p className="mt-1 font-mono text-xs text-[var(--text-muted)]">{id}</p>
              <p className="mt-3 text-3xl font-bold tracking-tight">
                {balances[id] ? formatSAR(balances[id]) : "…"}
              </p>
              <Link
                href={`/accounts?focus=${id}`}
                className="mt-4 inline-flex text-sm text-[var(--accent)] hover:underline"
              >
                View ledger entries →
              </Link>
            </Card>
          ))}
        </div>
      </section>

      <Card title="How it works" subtitle="What happens when you send money">
        <ol className="space-y-3 text-sm text-[var(--text-muted)]">
          <li className="flex gap-3">
            <Badge tone="info">1</Badge>
            Fraud & compliance checks run before any money moves.
          </li>
          <li className="flex gap-3">
            <Badge tone="info">2</Badge>
            Funds post to the double-entry ledger atomically with an outbox event.
          </li>
          <li className="flex gap-3">
            <Badge tone="info">3</Badge>
            The worker settles via mock SARIE — track progress on the Track page.
          </li>
        </ol>
      </Card>
    </div>
  );
}

function StatusPill({ label, status }: { label: string; status: "loading" | "up" | "down" }) {
  return (
    <div className="flex items-center gap-2 rounded-xl bg-[var(--bg-elevated)] px-3 py-2 ring-1 ring-[var(--border)]">
      <Activity
        className={`h-4 w-4 ${
          status === "up"
            ? "text-[var(--accent)]"
            : status === "down"
              ? "text-[var(--danger)]"
              : "text-[var(--text-muted)] pulse-dot"
        }`}
      />
      <span className="text-sm">{label}</span>
      <Badge tone={status === "up" ? "success" : status === "down" ? "danger" : "info"}>
        {status === "loading" ? "…" : status === "up" ? "Online" : "Offline"}
      </Badge>
    </div>
  );
}

function humanName(id: string): string {
  if (id === "wallet-alice") return "Alice's wallet";
  if (id === "wallet-bob") return "Bob's wallet";
  return id;
}
