"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import {
  Activity,
  ArrowDownLeft,
  ArrowRight,
  ArrowUpRight,
  CreditCard,
  Eye,
  EyeOff,
  RefreshCw,
  Send,
  Shield,
  TrendingUp,
  Wallet,
  Zap,
} from "lucide-react";
import { api } from "@/lib/api";
import { formatSAR } from "@/lib/utils";
import { Alert, Badge, Button, Card, MetricCard } from "@/components/ui";

const DEMO_ACCOUNTS = ["wallet-alice", "wallet-bob"];

export default function DashboardPage() {
  const [health, setHealth] = useState<"loading" | "up" | "down">("loading");
  const [ready, setReady] = useState<"loading" | "up" | "down">("loading");
  const [balances, setBalances] = useState<Record<string, string>>({});
  const [error, setError] = useState<string | null>(null);
  const [balanceVisible, setBalanceVisible] = useState(true);
  const [loading, setLoading] = useState(true);

  async function load() {
    setError(null);
    setLoading(true);
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
    setLoading(false);
  }

  useEffect(() => {
    load();
  }, []);

  const totalBalance = Object.values(balances)
    .filter((v) => v !== "—")
    .reduce((sum, v) => sum + parseFloat(v), 0);

  return (
    <div className="mx-auto max-w-6xl space-y-8 animate-in">
      <header className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <p className="text-sm font-medium text-[var(--text-muted)]">
            {new Date().toLocaleDateString("en-US", {
              weekday: "long",
              year: "numeric",
              month: "long",
              day: "numeric",
            })}
          </p>
          <h1 className="mt-1 text-2xl font-bold tracking-tight md:text-3xl">
            Welcome back
          </h1>
          <p className="mt-1 text-sm text-[var(--text-muted)]">
            Here&apos;s an overview of your banking operations
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="ghost" size="sm" onClick={load}>
            <RefreshCw className={`h-4 w-4 ${loading ? "animate-spin" : ""}`} />
            Refresh
          </Button>
          <Link href="/transfer">
            <Button size="sm">
              <Send className="h-4 w-4" />
              New Transfer
            </Button>
          </Link>
        </div>
      </header>

      {error && <Alert tone="danger">{error}</Alert>}

      {/* Total Balance Hero */}
      <div className="relative overflow-hidden rounded-2xl border border-[var(--border)] bg-gradient-to-br from-[var(--bg-card)] via-[var(--bg-card)] to-emerald-950/20 p-8">
        <div className="absolute -right-20 -top-20 h-64 w-64 rounded-full bg-emerald-500/5 blur-3xl" />
        <div className="absolute -bottom-10 -left-10 h-40 w-40 rounded-full bg-blue-500/5 blur-3xl" />
        <div className="relative">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <p className="text-sm font-medium text-[var(--text-muted)]">Total Portfolio Balance</p>
              <button
                onClick={() => setBalanceVisible(!balanceVisible)}
                className="text-[var(--text-muted)] hover:text-[var(--text)] transition-colors"
              >
                {balanceVisible ? (
                  <Eye className="h-4 w-4" />
                ) : (
                  <EyeOff className="h-4 w-4" />
                )}
              </button>
            </div>
            <Badge tone="success" dot>
              SAR
            </Badge>
          </div>
          <p className="mt-3 text-4xl font-bold tracking-tight md:text-5xl">
            {loading ? (
              <span className="inline-block h-12 w-60 rounded-lg bg-[var(--bg-surface)] shimmer" />
            ) : balanceVisible ? (
              formatSAR(totalBalance.toFixed(2))
            ) : (
              "SAR •••••••"
            )}
          </p>
          <div className="mt-6 flex flex-wrap gap-3">
            <Link href="/transfer">
              <Button size="md">
                <ArrowUpRight className="h-4 w-4" />
                Send
              </Button>
            </Link>
            <Link href="/accounts">
              <Button variant="secondary" size="md">
                <ArrowDownLeft className="h-4 w-4" />
                Accounts
              </Button>
            </Link>
            <Link href="/track">
              <Button variant="ghost" size="md">
                <Activity className="h-4 w-4" />
                Track Transfer
              </Button>
            </Link>
          </div>
        </div>
      </div>

      {/* Wallet Cards */}
      <section>
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-lg font-semibold tracking-tight">Wallets</h2>
          <Link href="/accounts" className="text-sm text-[var(--accent)] hover:underline">
            View all →
          </Link>
        </div>
        <div className="grid gap-4 sm:grid-cols-2">
          {DEMO_ACCOUNTS.map((id, i) => (
            <WalletCard
              key={id}
              id={id}
              balance={balances[id]}
              loading={loading}
              visible={balanceVisible}
              index={i}
            />
          ))}
        </div>
      </section>

      {/* Status + Quick Actions grid */}
      <div className="grid gap-4 lg:grid-cols-3">
        {/* System Status */}
        <Card title="System Status" subtitle="Live infrastructure health" className="lg:col-span-1">
          <div className="space-y-3">
            <StatusRow
              label="Banking API"
              description="Core transaction engine"
              status={health}
            />
            <StatusRow
              label="Database"
              description="PostgreSQL + Redis"
              status={ready}
            />
            <StatusRow
              label="Settlement"
              description="SARIE gateway mock"
              status={health === "up" && ready === "up" ? "up" : health === "loading" ? "loading" : "down"}
            />
          </div>
          {(health === "down" || ready === "down") && (
            <div className="mt-4 rounded-lg bg-red-500/5 p-3 text-xs text-red-300">
              Backend is offline. Run{" "}
              <code className="rounded bg-red-500/10 px-1 py-0.5 font-mono text-red-400">
                go run ./cmd/api
              </code>{" "}
              and{" "}
              <code className="rounded bg-red-500/10 px-1 py-0.5 font-mono text-red-400">
                docker compose up -d
              </code>
            </div>
          )}
        </Card>

        {/* Transfer Flow */}
        <Card
          title="Transfer Pipeline"
          subtitle="How your money moves securely"
          className="lg:col-span-2"
        >
          <div className="grid gap-3 sm:grid-cols-4">
            <PipelineStep
              step={1}
              icon={<Shield className="h-5 w-5" />}
              title="Fraud Check"
              description="Amount limits, velocity checks, sanctions screening"
              color="blue"
            />
            <PipelineStep
              step={2}
              icon={<Zap className="h-5 w-5" />}
              title="Compliance"
              description="KYC verification, IBAN validation, KSA regulations"
              color="purple"
            />
            <PipelineStep
              step={3}
              icon={<CreditCard className="h-5 w-5" />}
              title="Ledger Post"
              description="Atomic double-entry debit/credit with outbox event"
              color="amber"
            />
            <PipelineStep
              step={4}
              icon={<TrendingUp className="h-5 w-5" />}
              title="Settlement"
              description="SARIE network settlement with auto-compensation"
              color="emerald"
            />
          </div>
        </Card>
      </div>

      {/* Quick Stats */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <MetricCard
          label="Active Wallets"
          value={DEMO_ACCOUNTS.length.toString()}
          icon={<Wallet className="h-5 w-5" />}
        />
        <MetricCard
          label="Settlement Mode"
          value="SARIE Mock"
          icon={<Activity className="h-5 w-5" />}
        />
        <MetricCard
          label="Currency"
          value="SAR"
          subValue="Saudi Riyal"
          icon={<CreditCard className="h-5 w-5" />}
        />
        <MetricCard
          label="Auth Mode"
          value="API Key"
          subValue="X-API-Key header"
          icon={<Shield className="h-5 w-5" />}
        />
      </div>
    </div>
  );
}

function WalletCard({
  id,
  balance,
  loading,
  visible,
  index,
}: {
  id: string;
  balance?: string;
  loading: boolean;
  visible: boolean;
  index: number;
}) {
  const name = id === "wallet-alice" ? "Alice" : id === "wallet-bob" ? "Bob" : id;
  const initials = name.slice(0, 2).toUpperCase();
  const gradients = [
    "from-emerald-600 to-teal-700",
    "from-blue-600 to-indigo-700",
  ];

  return (
    <div className="group relative overflow-hidden rounded-2xl border border-[var(--border)] bg-[var(--bg-card)] p-6 transition-all hover:border-[var(--accent)]/15 hover:shadow-lg hover:shadow-emerald-900/5">
      <div className="absolute -right-12 -top-12 h-32 w-32 rounded-full bg-[var(--accent-glow)] opacity-50 blur-3xl transition-opacity group-hover:opacity-100" />
      <div className="relative flex items-start justify-between">
        <div className="flex items-center gap-3">
          <div className={`flex h-11 w-11 items-center justify-center rounded-xl bg-gradient-to-br ${gradients[index % 2]} text-sm font-bold text-white shadow-lg`}>
            {initials}
          </div>
          <div>
            <p className="font-semibold">{name}&apos;s Wallet</p>
            <p className="font-mono text-xs text-[var(--text-muted)]">{id}</p>
          </div>
        </div>
        <Badge tone="success" size="sm" dot>
          Active
        </Badge>
      </div>
      <div className="mt-5">
        <p className="text-xs font-medium text-[var(--text-muted)]">Available Balance</p>
        <p className="mt-1 text-2xl font-bold tracking-tight">
          {loading ? (
            <span className="inline-block h-8 w-40 rounded-lg bg-[var(--bg-surface)] shimmer" />
          ) : visible && balance ? (
            formatSAR(balance)
          ) : (
            "•••••"
          )}
        </p>
      </div>
      <div className="mt-5 flex gap-2">
        <Link href="/transfer" className="flex-1">
          <Button variant="secondary" size="sm" className="w-full">
            <Send className="h-3.5 w-3.5" />
            Send
          </Button>
        </Link>
        <Link href={`/accounts?focus=${id}`} className="flex-1">
          <Button variant="ghost" size="sm" className="w-full">
            Ledger →
          </Button>
        </Link>
      </div>
    </div>
  );
}

function StatusRow({
  label,
  description,
  status,
}: {
  label: string;
  description: string;
  status: "loading" | "up" | "down";
}) {
  return (
    <div className="flex items-center justify-between rounded-xl bg-[var(--bg-elevated)] px-4 py-3 ring-1 ring-[var(--border)]">
      <div className="flex items-center gap-3">
        <div className="relative">
          <div
            className={`h-2.5 w-2.5 rounded-full ${
              status === "up"
                ? "bg-emerald-400"
                : status === "down"
                  ? "bg-red-400"
                  : "bg-[var(--text-muted)] pulse-dot"
            }`}
          />
          {status === "up" && (
            <div className="absolute inset-0 h-2.5 w-2.5 rounded-full bg-emerald-400 pulse-ring" />
          )}
        </div>
        <div>
          <p className="text-sm font-medium">{label}</p>
          <p className="text-[11px] text-[var(--text-muted)]">{description}</p>
        </div>
      </div>
      <Badge
        tone={status === "up" ? "success" : status === "down" ? "danger" : "neutral"}
        size="sm"
      >
        {status === "loading" ? "Checking…" : status === "up" ? "Online" : "Offline"}
      </Badge>
    </div>
  );
}

function PipelineStep({
  step,
  icon,
  title,
  description,
  color,
}: {
  step: number;
  icon: React.ReactNode;
  title: string;
  description: string;
  color: "blue" | "purple" | "amber" | "emerald";
}) {
  const colors = {
    blue: "bg-blue-500/10 text-blue-400 ring-blue-500/20",
    purple: "bg-purple-500/10 text-purple-400 ring-purple-500/20",
    amber: "bg-amber-500/10 text-amber-400 ring-amber-500/20",
    emerald: "bg-emerald-500/10 text-emerald-400 ring-emerald-500/20",
  };
  return (
    <div className="rounded-xl bg-[var(--bg-elevated)] p-4 ring-1 ring-[var(--border)]">
      <div className={`mb-3 flex h-10 w-10 items-center justify-center rounded-xl ring-1 ${colors[color]}`}>
        {icon}
      </div>
      <p className="text-xs font-bold text-[var(--text-muted)]">Step {step}</p>
      <p className="mt-0.5 text-sm font-semibold">{title}</p>
      <p className="mt-1 text-[11px] leading-relaxed text-[var(--text-muted)]">{description}</p>
    </div>
  );
}
