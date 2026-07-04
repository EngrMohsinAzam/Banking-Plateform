import { cn } from "@/lib/utils";
import { ChevronDown } from "lucide-react";
import { forwardRef } from "react";

export function Card({
  className,
  children,
  title,
  subtitle,
  action,
  noPadding,
}: {
  className?: string;
  children: React.ReactNode;
  title?: string;
  subtitle?: string;
  action?: React.ReactNode;
  noPadding?: boolean;
}) {
  return (
    <div
      className={cn(
        "rounded-2xl border border-[var(--border)] bg-[var(--bg-card)] shadow-lg shadow-black/25 transition-colors hover:border-[var(--border-subtle)]",
        !noPadding && "p-6",
        className,
      )}
    >
      {(title || subtitle || action) && (
        <div className={cn("flex items-start justify-between gap-4", noPadding && "px-6 pt-6", (title || subtitle) && "mb-5")}>
          <div>
            {title && <h2 className="text-[15px] font-semibold tracking-tight">{title}</h2>}
            {subtitle && <p className="mt-1 text-sm text-[var(--text-muted)]">{subtitle}</p>}
          </div>
          {action}
        </div>
      )}
      {children}
    </div>
  );
}

export function MetricCard({
  label,
  value,
  subValue,
  icon,
  trend,
  className,
}: {
  label: string;
  value: string;
  subValue?: string;
  icon?: React.ReactNode;
  trend?: { value: string; positive: boolean };
  className?: string;
}) {
  return (
    <div
      className={cn(
        "group relative overflow-hidden rounded-2xl border border-[var(--border)] bg-[var(--bg-card)] p-6 transition-all hover:border-[var(--accent)]/20 hover:shadow-lg hover:shadow-[var(--accent-glow)]",
        className,
      )}
    >
      <div className="absolute -right-6 -top-6 h-24 w-24 rounded-full bg-[var(--accent-glow)] opacity-0 blur-2xl transition-opacity group-hover:opacity-100" />
      <div className="flex items-start justify-between">
        <div className="space-y-3">
          <p className="text-sm font-medium text-[var(--text-muted)]">{label}</p>
          <p className="text-2xl font-bold tracking-tight">{value}</p>
          {subValue && (
            <p className="text-xs text-[var(--text-muted)] font-mono">{subValue}</p>
          )}
        </div>
        <div className="flex flex-col items-end gap-2">
          {icon && (
            <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-[var(--accent-glow-strong)] text-[var(--accent)]">
              {icon}
            </div>
          )}
          {trend && (
            <span
              className={cn(
                "text-xs font-semibold",
                trend.positive ? "text-emerald-400" : "text-red-400",
              )}
            >
              {trend.positive ? "↑" : "↓"} {trend.value}
            </span>
          )}
        </div>
      </div>
    </div>
  );
}

export function Badge({
  tone = "info",
  size = "sm",
  dot,
  children,
}: {
  tone?: "success" | "warning" | "danger" | "info" | "neutral";
  size?: "sm" | "md";
  dot?: boolean;
  children: React.ReactNode;
}) {
  const tones = {
    success: "bg-emerald-500/10 text-emerald-400 ring-emerald-500/20",
    warning: "bg-amber-500/10 text-amber-400 ring-amber-500/20",
    danger: "bg-red-500/10 text-red-400 ring-red-500/20",
    info: "bg-blue-500/10 text-blue-400 ring-blue-500/20",
    neutral: "bg-[var(--bg-surface)] text-[var(--text-muted)] ring-[var(--border)]",
  };
  const dotColors = {
    success: "bg-emerald-400",
    warning: "bg-amber-400",
    danger: "bg-red-400",
    info: "bg-blue-400",
    neutral: "bg-[var(--text-muted)]",
  };
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1.5 rounded-full font-medium ring-1 ring-inset",
        size === "sm" ? "px-2.5 py-0.5 text-[11px]" : "px-3 py-1 text-xs",
        tones[tone],
      )}
    >
      {dot && (
        <span className={cn("h-1.5 w-1.5 rounded-full", dotColors[tone])} />
      )}
      {children}
    </span>
  );
}

export function Button({
  className,
  variant = "primary",
  size = "md",
  ...props
}: React.ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: "primary" | "secondary" | "ghost" | "danger";
  size?: "sm" | "md" | "lg";
}) {
  const variants = {
    primary:
      "bg-[var(--accent)] text-white hover:bg-[var(--accent-hover)] font-semibold shadow-lg shadow-emerald-900/20 active:scale-[0.98]",
    secondary:
      "bg-[var(--bg-surface)] text-[var(--text-secondary)] ring-1 ring-[var(--border)] hover:bg-[var(--bg-card-hover)] hover:text-[var(--text)] active:scale-[0.98]",
    ghost:
      "text-[var(--text-muted)] hover:bg-[var(--bg-surface)] hover:text-[var(--text)]",
    danger:
      "bg-red-500/10 text-red-400 ring-1 ring-red-500/20 hover:bg-red-500/20 active:scale-[0.98]",
  };
  const sizes = {
    sm: "px-3 py-1.5 text-xs rounded-lg gap-1.5",
    md: "px-4 py-2.5 text-sm rounded-xl gap-2",
    lg: "px-6 py-3 text-sm rounded-xl gap-2.5",
  };
  return (
    <button
      className={cn(
        "inline-flex items-center justify-center font-medium transition-all duration-150 disabled:cursor-not-allowed disabled:opacity-40",
        variants[variant],
        sizes[size],
        className,
      )}
      {...props}
    />
  );
}

export const Input = forwardRef<
  HTMLInputElement,
  React.InputHTMLAttributes<HTMLInputElement> & { error?: boolean }
>(({ className, error, ...props }, ref) => (
  <input
    ref={ref}
    className={cn(
      "w-full rounded-xl border bg-[var(--bg-elevated)] px-4 py-2.5 text-sm text-[var(--text)] placeholder:text-[var(--text-muted)]/60 transition-all focus:outline-none focus:ring-2",
      error
        ? "border-red-500/40 focus:border-red-500/60 focus:ring-red-500/15"
        : "border-[var(--border)] focus:border-[var(--accent)]/40 focus:ring-[var(--accent)]/10",
      className,
    )}
    {...props}
  />
));
Input.displayName = "Input";

export function Select({
  className,
  children,
  ...props
}: React.SelectHTMLAttributes<HTMLSelectElement>) {
  return (
    <div className="relative">
      <select
        className={cn(
          "w-full appearance-none rounded-xl border border-[var(--border)] bg-[var(--bg-elevated)] px-4 py-2.5 pr-10 text-sm text-[var(--text)] transition-all focus:border-[var(--accent)]/40 focus:outline-none focus:ring-2 focus:ring-[var(--accent)]/10",
          className,
        )}
        {...props}
      >
        {children}
      </select>
      <ChevronDown className="pointer-events-none absolute right-3 top-1/2 h-4 w-4 -translate-y-1/2 text-[var(--text-muted)]" />
    </div>
  );
}

export function Label({
  children,
  htmlFor,
  required,
}: {
  children: React.ReactNode;
  htmlFor?: string;
  required?: boolean;
}) {
  return (
    <label
      htmlFor={htmlFor}
      className="mb-2 flex items-center gap-1 text-sm font-medium text-[var(--text-secondary)]"
    >
      {children}
      {required && <span className="text-[var(--accent)]">*</span>}
    </label>
  );
}

export function Alert({
  tone = "info",
  title,
  children,
}: {
  tone?: "success" | "warning" | "danger" | "info";
  title?: string;
  children: React.ReactNode;
}) {
  const tones = {
    success: "border-emerald-500/20 bg-emerald-500/5 text-emerald-300",
    warning: "border-amber-500/20 bg-amber-500/5 text-amber-300",
    danger: "border-red-500/20 bg-red-500/5 text-red-300",
    info: "border-blue-500/20 bg-blue-500/5 text-blue-300",
  };
  return (
    <div className={cn("rounded-xl border px-4 py-3 text-sm", tones[tone])}>
      {title && <p className="mb-1 font-semibold">{title}</p>}
      {children}
    </div>
  );
}

export function Divider({ className }: { className?: string }) {
  return <div className={cn("h-px bg-[var(--border)]", className)} />;
}

export function Skeleton({ className }: { className?: string }) {
  return (
    <div
      className={cn(
        "rounded-lg bg-[var(--bg-surface)] shimmer",
        className,
      )}
    />
  );
}

export function EmptyState({
  icon,
  title,
  description,
  action,
}: {
  icon: React.ReactNode;
  title: string;
  description: string;
  action?: React.ReactNode;
}) {
  return (
    <div className="flex flex-col items-center justify-center py-16 text-center">
      <div className="mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-[var(--bg-surface)] text-[var(--text-muted)]">
        {icon}
      </div>
      <h3 className="mb-2 text-lg font-semibold">{title}</h3>
      <p className="mb-6 max-w-sm text-sm text-[var(--text-muted)]">{description}</p>
      {action}
    </div>
  );
}

export function StepIndicator({
  steps,
  current,
}: {
  steps: string[];
  current: number;
}) {
  return (
    <div className="flex items-center gap-1">
      {steps.map((step, i) => (
        <div key={step} className="flex items-center gap-1">
          <div className="flex items-center gap-2">
            <div
              className={cn(
                "flex h-8 w-8 items-center justify-center rounded-full text-xs font-bold transition-all",
                i < current
                  ? "bg-[var(--accent)] text-white"
                  : i === current
                    ? "bg-[var(--accent-glow-strong)] text-[var(--accent)] ring-2 ring-[var(--accent)]/30"
                    : "bg-[var(--bg-surface)] text-[var(--text-muted)] ring-1 ring-[var(--border)]",
              )}
            >
              {i < current ? "✓" : i + 1}
            </div>
            <span
              className={cn(
                "hidden text-xs font-medium sm:inline",
                i <= current ? "text-[var(--text)]" : "text-[var(--text-muted)]",
              )}
            >
              {step}
            </span>
          </div>
          {i < steps.length - 1 && (
            <div
              className={cn(
                "mx-2 h-px w-8 sm:w-12",
                i < current ? "bg-[var(--accent)]" : "bg-[var(--border)]",
              )}
            />
          )}
        </div>
      ))}
    </div>
  );
}
