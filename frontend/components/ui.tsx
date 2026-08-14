import Link from "next/link";
import { ConnectWallet } from "./ConnectWallet";

export function TopBar({ active }: { active: "taker" | "maker" }) {
  return (
    <header className="sticky top-0 z-10 border-b border-line bg-base/85 backdrop-blur">
      <div className="mx-auto flex h-14 w-full max-w-5xl items-center gap-6 px-5">
        <Link href="/" className="flex items-center gap-2.5 text-ink hover:text-accent">
          <span className="h-2.5 w-2.5 rounded-[2px] bg-accent" />
          <span className="display text-[19px] tracking-[0.02em]">FXRP Dark RFQ</span>
        </Link>

        <nav className="flex items-center gap-1 text-[13px]">
          {(["taker", "maker"] as const).map((r) => (
            <Link
              key={r}
              href={`/${r}`}
              className={
                active === r
                  ? "rounded-sm bg-raised px-2.5 py-1 text-ink"
                  : "rounded-sm px-2.5 py-1 text-faint transition-colors duration-150 hover:text-muted"
              }
            >
              {r === "taker" ? "Taker" : "Maker"}
            </Link>
          ))}
        </nav>

        <div className="ml-auto flex items-center gap-3">
          <span className="hidden items-center gap-1.5 text-[11px] text-faint sm:flex">
            <span className="h-1.5 w-1.5 rounded-full bg-positive" />
            Coston2
          </span>
          <ConnectWallet />
        </div>
      </div>
    </header>
  );
}

export function Panel({
  title,
  aside,
  children,
  className = "",
}: {
  title?: string;
  aside?: React.ReactNode;
  children: React.ReactNode;
  className?: string;
}) {
  return (
    <section className={`overflow-hidden rounded-[10px] border border-line bg-panel ${className}`}>
      {title && (
        <div className="flex items-center justify-between border-b border-line px-4 py-2.5">
          <h2 className="text-[11px] font-medium uppercase tracking-[0.13em] text-muted">{title}</h2>
          {aside}
        </div>
      )}
      {children}
    </section>
  );
}

export function Field({
  label,
  hint,
  value,
  onChange,
  suffix,
  mono = true,
  placeholder,
}: {
  label: string;
  hint?: React.ReactNode;
  value: string;
  onChange: (v: string) => void;
  suffix?: string;
  mono?: boolean;
  placeholder?: string;
}) {
  return (
    <label className="block">
      <div className="mb-1.5 flex items-baseline justify-between gap-3">
        <span className="text-[11px] uppercase tracking-[0.13em] text-muted">{label}</span>
        {hint}
      </div>
      <div className="flex items-center rounded-[4px] border border-line bg-base focus-within:border-line-strong">
        <input
          value={value}
          onChange={(e) => onChange(e.target.value)}
          placeholder={placeholder}
          className={`w-full bg-transparent px-3 py-2.5 text-[14px] text-ink outline-none ${mono ? "font-mono tnum" : ""}`}
        />
        {/* Mono so USDT0's trailing zero doesn't read as the letter O. */}
        {suffix && <span className="shrink-0 pr-3 font-mono text-[12px] text-faint">{suffix}</span>}
      </div>
    </label>
  );
}

export function Btn({
  children,
  onClick,
  variant = "primary",
  disabled,
  className = "",
}: {
  children: React.ReactNode;
  onClick?: () => void;
  variant?: "primary" | "ghost";
  disabled?: boolean;
  className?: string;
}) {
  const base =
    "inline-flex items-center justify-center px-4 py-2.5 text-[13px] font-medium transition-colors duration-150 disabled:cursor-not-allowed disabled:opacity-40";
  const styles =
    variant === "primary"
      ? "bg-accent text-base hover:bg-[#f2b357]"
      : "border border-line text-muted hover:border-line-strong hover:text-ink";
  return (
    <button onClick={onClick} disabled={disabled} className={`${base} ${styles} ${className}`}>
      {children}
    </button>
  );
}

/** Segmented buy/sell control. Direction is the one place color carries meaning. */
export function SideToggle({
  value,
  onChange,
  labels,
}: {
  value: "buy" | "sell";
  onChange: (v: "buy" | "sell") => void;
  labels: { buy: string; sell: string };
}) {
  return (
    <div className="grid grid-cols-2 gap-px border border-line bg-line">
      {(["buy", "sell"] as const).map((s) => {
        const on = value === s;
        const tone = s === "buy" ? "text-positive" : "text-negative";
        return (
          <button
            key={s}
            onClick={() => onChange(s)}
            className={`px-3 py-2.5 text-[13px] font-medium transition-colors duration-150 ${
              on ? `bg-raised ${tone}` : "bg-panel text-faint hover:text-muted"
            }`}
          >
            {labels[s]}
          </button>
        );
      })}
    </div>
  );
}

/** Tiny mono category chip. Colour is semantic, never decorative. */
export function Tag({
  children,
  tone = "neutral",
}: {
  children: React.ReactNode;
  tone?: "neutral" | "accent" | "positive" | "negative";
}) {
  const tones = {
    neutral: "border-line-strong bg-raised text-muted",
    accent: "border-accent/35 bg-accent/10 text-accent",
    positive: "border-positive/30 bg-positive/10 text-positive",
    negative: "border-negative/30 bg-negative/10 text-negative",
  }[tone];
  return (
    <span
      className={`inline-block rounded-[2px] border px-[0.6em] py-[0.22em] font-mono text-[11px] uppercase tracking-[0.12em] ${tones}`}
    >
      {children}
    </span>
  );
}

/** Mono trail with an underlined brass link and a faded separator. */
export function Breadcrumb({ trail }: { trail: { label: string; href?: string }[] }) {
  return (
    <nav className="mb-7 font-mono text-[11px] uppercase tracking-[0.15em] text-faint">
      {trail.map((t, i) => (
        <span key={t.label}>
          {i > 0 && <span className="mx-2 opacity-40">›</span>}
          {t.href ? (
            <Link
              href={t.href}
              className="text-accent-dim underline decoration-accent/40 underline-offset-[0.25em] transition-colors duration-150 hover:text-accent hover:decoration-accent"
            >
              {t.label}
            </Link>
          ) : (
            <span className="text-muted">{t.label}</span>
          )}
        </span>
      ))}
    </nav>
  );
}

/** Brass-ringed disc for step markers. */
export function Medallion({ children }: { children: React.ReactNode }) {
  return (
    <span className="medallion inline-flex h-11 w-11 shrink-0 items-center justify-center rounded-full font-mono text-[12px] text-accent">
      {children}
    </span>
  );
}

/** Label/value row. Values are mono so hashes and amounts align down the column. */
export function Row({ k, v, tone = "" }: { k: string; v: React.ReactNode; tone?: string }) {
  return (
    <div className="flex items-baseline justify-between gap-4 py-1.5">
      <span className="shrink-0 text-[12px] text-faint">{k}</span>
      <span className={`truncate font-mono tnum text-[12px] ${tone || "text-ink"}`}>{v}</span>
    </div>
  );
}
