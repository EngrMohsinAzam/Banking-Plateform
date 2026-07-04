import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  ArrowLeftRight,
  Building2,
  LayoutDashboard,
  Search,
  Settings,
} from "lucide-react";
import { cn } from "@/lib/utils";

const nav = [
  { href: "/", label: "Dashboard", icon: LayoutDashboard },
  { href: "/transfer", label: "Send money", icon: ArrowLeftRight },
  { href: "/track", label: "Track transfer", icon: Search },
  { href: "/accounts", label: "Accounts", icon: Building2 },
  { href: "/settings", label: "Settings", icon: Settings },
];

export function Sidebar() {
  const pathname = usePathname();

  return (
    <aside className="hidden w-64 shrink-0 border-r border-[var(--border)] bg-[var(--bg-elevated)] lg:flex lg:flex-col">
      <div className="border-b border-[var(--border)] px-6 py-5">
        <div className="flex items-center gap-3">
          <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-[var(--accent-glow)] ring-1 ring-[var(--accent)]/30">
            <span className="text-lg font-bold text-[var(--accent)]">ب</span>
          </div>
          <div>
            <p className="text-sm font-semibold tracking-tight">Bayan Bank</p>
            <p className="text-xs text-[var(--text-muted)]">KSA Core Banking</p>
          </div>
        </div>
      </div>
      <nav className="flex flex-1 flex-col gap-1 p-4">
        {nav.map(({ href, label, icon: Icon }) => {
          const active = pathname === href || (href !== "/" && pathname.startsWith(href));
          return (
            <Link
              key={href}
              href={href}
              className={cn(
                "flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-medium transition-colors",
                active
                  ? "bg-[var(--accent-glow)] text-[var(--accent)] ring-1 ring-[var(--accent)]/25"
                  : "text-[var(--text-muted)] hover:bg-[var(--bg-card)] hover:text-[var(--text)]",
              )}
            >
              <Icon className="h-4 w-4" />
              {label}
            </Link>
          );
        })}
      </nav>
      <div className="border-t border-[var(--border)] p-4 text-xs text-[var(--text-muted)]">
        Portfolio demo · SAR wallets · SARIE mock
      </div>
    </aside>
  );
}

export function MobileNav() {
  const pathname = usePathname();
  return (
    <nav className="flex gap-1 overflow-x-auto border-b border-[var(--border)] bg-[var(--bg-elevated)] p-2 lg:hidden">
      {nav.map(({ href, label, icon: Icon }) => {
        const active = pathname === href;
        return (
          <Link
            key={href}
            href={href}
            className={cn(
              "flex shrink-0 items-center gap-2 rounded-lg px-3 py-2 text-xs font-medium",
              active ? "bg-[var(--accent-glow)] text-[var(--accent)]" : "text-[var(--text-muted)]",
            )}
          >
            <Icon className="h-3.5 w-3.5" />
            {label}
          </Link>
        );
      })}
    </nav>
  );
}
