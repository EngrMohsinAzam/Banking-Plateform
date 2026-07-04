"use client";

import { MobileNav, Sidebar } from "./Sidebar";
import { Bell, HelpCircle, Search } from "lucide-react";

export function AppShell({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex min-h-screen">
      <Sidebar />
      <div className="flex min-w-0 flex-1 flex-col">
        <MobileNav />
        <header className="hidden h-[var(--header-height)] shrink-0 items-center justify-between border-b border-[var(--border)] bg-[var(--bg-elevated)]/50 px-8 lg:flex">
          <div className="flex items-center gap-3">
            <div className="flex h-9 items-center gap-2 rounded-xl bg-[var(--bg-card)] px-3 ring-1 ring-[var(--border)] transition-all focus-within:ring-[var(--accent)]/30">
              <Search className="h-4 w-4 text-[var(--text-muted)]" />
              <input
                type="text"
                placeholder="Search accounts, transfers..."
                className="h-full w-60 bg-transparent text-sm text-[var(--text)] placeholder:text-[var(--text-muted)]/60 outline-none"
              />
              <kbd className="hidden rounded bg-[var(--bg-surface)] px-1.5 py-0.5 text-[10px] font-medium text-[var(--text-muted)] sm:inline">
                ⌘K
              </kbd>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <HeaderButton icon={<HelpCircle className="h-[18px] w-[18px]" />} />
            <HeaderButton
              icon={<Bell className="h-[18px] w-[18px]" />}
              badge
            />
            <div className="ml-2 h-6 w-px bg-[var(--border)]" />
            <div className="ml-2 flex items-center gap-3">
              <div className="flex h-8 w-8 items-center justify-center rounded-full bg-gradient-to-br from-emerald-600 to-teal-700 text-xs font-bold text-white ring-2 ring-[var(--bg-elevated)]">
                SA
              </div>
            </div>
          </div>
        </header>
        <main className="flex-1 overflow-y-auto p-4 md:p-8">{children}</main>
      </div>
    </div>
  );
}

function HeaderButton({
  icon,
  badge,
}: {
  icon: React.ReactNode;
  badge?: boolean;
}) {
  return (
    <button className="relative flex h-9 w-9 items-center justify-center rounded-xl text-[var(--text-muted)] transition-colors hover:bg-[var(--bg-card)] hover:text-[var(--text)]">
      {icon}
      {badge && (
        <span className="absolute right-1.5 top-1.5 h-2 w-2 rounded-full bg-[var(--accent)] ring-2 ring-[var(--bg-elevated)]" />
      )}
    </button>
  );
}
