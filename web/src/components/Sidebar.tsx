"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  ArrowLeftRight,
  Building2,
  CreditCard,
  LayoutDashboard,
  Menu,
  Search,
  Settings,
  Shield,
  X,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { useState } from "react";

const mainNav = [
  { href: "/", label: "Overview", icon: LayoutDashboard },
  { href: "/transfer", label: "Transfer Funds", icon: ArrowLeftRight },
  { href: "/track", label: "Track Transfer", icon: Search },
  { href: "/accounts", label: "Accounts", icon: Building2 },
];

const secondaryNav = [
  { href: "/settings", label: "Settings", icon: Settings },
];

export function Sidebar() {
  const pathname = usePathname();

  return (
    <aside className="hidden w-[var(--sidebar-width)] shrink-0 border-r border-[var(--border)] bg-[var(--bg-elevated)] lg:flex lg:flex-col">
      <div className="border-b border-[var(--border)] px-6 py-5">
        <Link href="/" className="flex items-center gap-3 group">
          <div className="relative flex h-10 w-10 items-center justify-center rounded-xl bg-gradient-to-br from-emerald-500 to-emerald-700 shadow-lg shadow-emerald-900/30">
            <span className="text-lg font-bold text-white">ب</span>
            <div className="absolute inset-0 rounded-xl ring-1 ring-inset ring-white/10" />
          </div>
          <div>
            <p className="text-sm font-bold tracking-tight text-[var(--text)]">Bayan Bank</p>
            <p className="text-[11px] text-[var(--text-muted)]">Core Banking Platform</p>
          </div>
        </Link>
      </div>

      <nav className="flex flex-1 flex-col px-3 pt-6">
        <p className="mb-2 px-3 text-[10px] font-semibold uppercase tracking-widest text-[var(--text-muted)]">
          Main Menu
        </p>
        <div className="space-y-0.5">
          {mainNav.map(({ href, label, icon: Icon }) => {
            const active = pathname === href || (href !== "/" && pathname.startsWith(href));
            return (
              <Link
                key={href}
                href={href}
                className={cn(
                  "group flex items-center gap-3 rounded-xl px-3 py-2.5 text-[13px] font-medium transition-all duration-150",
                  active
                    ? "bg-[var(--accent-glow-strong)] text-[var(--accent)] shadow-sm shadow-emerald-900/10"
                    : "text-[var(--text-muted)] hover:bg-[var(--bg-card)] hover:text-[var(--text-secondary)]",
                )}
              >
                <div
                  className={cn(
                    "flex h-8 w-8 items-center justify-center rounded-lg transition-colors",
                    active
                      ? "bg-[var(--accent)]/10"
                      : "bg-transparent group-hover:bg-[var(--bg-surface)]",
                  )}
                >
                  <Icon className="h-[18px] w-[18px]" />
                </div>
                {label}
                {active && (
                  <div className="ml-auto h-1.5 w-1.5 rounded-full bg-[var(--accent)]" />
                )}
              </Link>
            );
          })}
        </div>

        <div className="my-5 h-px bg-[var(--border)]" />

        <p className="mb-2 px-3 text-[10px] font-semibold uppercase tracking-widest text-[var(--text-muted)]">
          System
        </p>
        <div className="space-y-0.5">
          {secondaryNav.map(({ href, label, icon: Icon }) => {
            const active = pathname === href;
            return (
              <Link
                key={href}
                href={href}
                className={cn(
                  "group flex items-center gap-3 rounded-xl px-3 py-2.5 text-[13px] font-medium transition-all duration-150",
                  active
                    ? "bg-[var(--accent-glow-strong)] text-[var(--accent)]"
                    : "text-[var(--text-muted)] hover:bg-[var(--bg-card)] hover:text-[var(--text-secondary)]",
                )}
              >
                <div
                  className={cn(
                    "flex h-8 w-8 items-center justify-center rounded-lg transition-colors",
                    active ? "bg-[var(--accent)]/10" : "bg-transparent group-hover:bg-[var(--bg-surface)]",
                  )}
                >
                  <Icon className="h-[18px] w-[18px]" />
                </div>
                {label}
              </Link>
            );
          })}
        </div>

        <div className="mt-auto border-t border-[var(--border)] pb-4 pt-4">
          <div className="rounded-xl bg-gradient-to-br from-[var(--bg-card)] to-[var(--bg-surface)] p-4 ring-1 ring-[var(--border)]">
            <div className="flex items-center gap-3">
              <div className="flex h-9 w-9 items-center justify-center rounded-full bg-gradient-to-br from-emerald-600 to-teal-700 text-xs font-bold text-white">
                SA
              </div>
              <div className="min-w-0 flex-1">
                <p className="truncate text-sm font-medium text-[var(--text)]">Saudi Admin</p>
                <p className="truncate text-[11px] text-[var(--text-muted)]">Operations</p>
              </div>
              <Shield className="h-4 w-4 text-[var(--accent)]" />
            </div>
          </div>
        </div>
      </nav>
    </aside>
  );
}

export function MobileNav() {
  const pathname = usePathname();
  const [open, setOpen] = useState(false);

  return (
    <>
      <nav className="flex items-center justify-between border-b border-[var(--border)] bg-[var(--bg-elevated)] px-4 py-3 lg:hidden">
        <Link href="/" className="flex items-center gap-2.5">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-gradient-to-br from-emerald-500 to-emerald-700">
            <span className="text-sm font-bold text-white">ب</span>
          </div>
          <span className="text-sm font-bold">Bayan Bank</span>
        </Link>
        <button
          onClick={() => setOpen(!open)}
          className="flex h-9 w-9 items-center justify-center rounded-lg bg-[var(--bg-card)] text-[var(--text-muted)]"
        >
          {open ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
        </button>
      </nav>
      {open && (
        <div className="border-b border-[var(--border)] bg-[var(--bg-elevated)] px-3 pb-3 lg:hidden animate-in">
          {[...mainNav, ...secondaryNav].map(({ href, label, icon: Icon }) => {
            const active = pathname === href || (href !== "/" && pathname.startsWith(href));
            return (
              <Link
                key={href}
                href={href}
                onClick={() => setOpen(false)}
                className={cn(
                  "flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-medium",
                  active
                    ? "bg-[var(--accent-glow-strong)] text-[var(--accent)]"
                    : "text-[var(--text-muted)]",
                )}
              >
                <Icon className="h-4 w-4" />
                {label}
              </Link>
            );
          })}
        </div>
      )}
    </>
  );
}
