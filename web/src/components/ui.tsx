import { cn } from "@/lib/utils";

export function Card({
  className,
  children,
  title,
  subtitle,
}: {
  className?: string;
  children: React.ReactNode;
  title?: string;
  subtitle?: string;
}) {
  return (
    <div
      className={cn(
        "rounded-2xl border border-[var(--border)] bg-[var(--bg-card)] p-5 shadow-lg shadow-black/20",
        className,
      )}
    >
      {(title || subtitle) && (
        <div className="mb-4">
          {title && <h2 className="text-base font-semibold">{title}</h2>}
          {subtitle && <p className="mt-0.5 text-sm text-[var(--text-muted)]">{subtitle}</p>}
        </div>
      )}
      {children}
    </div>
  );
}

export function Badge({
  tone = "info",
  children,
}: {
  tone?: "success" | "warning" | "danger" | "info";
  children: React.ReactNode;
}) {
  const tones = {
    success: "bg-emerald-500/15 text-emerald-400 ring-emerald-500/30",
    warning: "bg-amber-500/15 text-amber-400 ring-amber-500/30",
    danger: "bg-red-500/15 text-red-400 ring-red-500/30",
    info: "bg-sky-500/15 text-sky-400 ring-sky-500/30",
  };
  return (
    <span
      className={cn(
        "inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ring-1 ring-inset",
        tones[tone],
      )}
    >
      {children}
    </span>
  );
}

export function Button({
  className,
  variant = "primary",
  ...props
}: React.ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: "primary" | "secondary" | "ghost";
}) {
  const variants = {
    primary:
      "bg-[var(--accent)] text-[#042f1f] hover:bg-[var(--accent-dim)] font-semibold shadow-lg shadow-emerald-900/30",
    secondary:
      "bg-[var(--bg-elevated)] text-[var(--text)] ring-1 ring-[var(--border)] hover:bg-[var(--bg-card)]",
    ghost: "text-[var(--text-muted)] hover:bg-[var(--bg-card)] hover:text-[var(--text)]",
  };
  return (
    <button
      className={cn(
        "inline-flex items-center justify-center gap-2 rounded-xl px-4 py-2.5 text-sm transition-all disabled:cursor-not-allowed disabled:opacity-50",
        variants[variant],
        className,
      )}
      {...props}
    />
  );
}

export function Input(props: React.InputHTMLAttributes<HTMLInputElement>) {
  return (
    <input
      className="w-full rounded-xl border border-[var(--border)] bg-[var(--bg-elevated)] px-3 py-2.5 text-sm text-[var(--text)] placeholder:text-[var(--text-muted)] focus:border-[var(--accent)] focus:outline-none focus:ring-2 focus:ring-[var(--accent)]/20"
      {...props}
    />
  );
}

export function Label({ children, htmlFor }: { children: React.ReactNode; htmlFor?: string }) {
  return (
    <label htmlFor={htmlFor} className="mb-1.5 block text-sm font-medium text-[var(--text-muted)]">
      {children}
    </label>
  );
}

export function Alert({
  tone = "info",
  children,
}: {
  tone?: "success" | "warning" | "danger" | "info";
  children: React.ReactNode;
}) {
  const tones = {
    success: "border-emerald-500/30 bg-emerald-500/10 text-emerald-200",
    warning: "border-amber-500/30 bg-amber-500/10 text-amber-200",
    danger: "border-red-500/30 bg-red-500/10 text-red-200",
    info: "border-sky-500/30 bg-sky-500/10 text-sky-200",
  };
  return (
    <div className={cn("rounded-xl border px-4 py-3 text-sm", tones[tone])}>{children}</div>
  );
}
