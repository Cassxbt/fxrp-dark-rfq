import Link from "next/link";

const FILL_TX = "0xe158ffe70bd1df2790ca3bc09c501cf214f6c7a7406872882361698551a7a8e9";

const STEPS = [
  { n: "01", t: "Taker seals an intent", d: "Side, size and a limit price, EIP-712 signed and encrypted to the TEE. The limit never touches the chain." },
  { n: "02", t: "Makers quote blind", d: "Each maker prices the RFQ without seeing the limit or any competing quote." },
  { n: "03", t: "The TEE matches", d: "Best qualifying quote wins inside the enclave. Losing quotes are never published." },
  { n: "04", t: "Settlement is atomic", d: "One transaction moves both legs. The Filled event is the only public trace." },
];

export default function Home() {
  return (
    <main className="grid-surface min-h-screen">
      <div className="mx-auto w-full max-w-5xl px-5">
        <header className="flex h-14 items-center gap-2.5 border-b border-line">
          <span className="h-2.5 w-2.5 rounded-[2px] bg-accent" />
          <span className="text-[13px] font-semibold tracking-tight">FXRP Dark RFQ</span>
          <span className="ml-auto flex items-center gap-1.5 text-[11px] text-faint">
            <span className="h-1.5 w-1.5 rounded-full bg-positive" />
            Coston2
          </span>
        </header>

        <section className="grid gap-10 border-b border-line py-16 md:grid-cols-[1.35fr_1fr] md:gap-14 md:py-20">
          <div>
            <p className="mb-4 text-[11px] uppercase tracking-[0.16em] text-accent">
              Sealed-bid OTC · Flare Confidential Compute
            </p>
            <h1 className="text-[34px] font-semibold leading-[1.12] tracking-[-0.02em] text-ink sm:text-[42px]">
              Trade FXRP without
              <br />
              showing your hand.
            </h1>
            <p className="mt-5 max-w-md text-[14px] leading-relaxed text-muted">
              A public order book leaks your size and your price before you ever trade. This desk keeps
              the taker&apos;s limit and every losing quote inside a TEE. The chain only learns who won,
              and only after it settled.
            </p>

            <div className="mt-8 flex flex-wrap items-center gap-3">
              <Link
                href="/taker"
                className="bg-accent px-5 py-2.5 text-[13px] font-medium text-base transition-colors duration-150 hover:bg-[#f2b357]"
              >
                Open an RFQ
              </Link>
              <Link
                href="/maker"
                className="border border-line px-5 py-2.5 text-[13px] font-medium text-muted transition-colors duration-150 hover:border-line-strong hover:text-ink"
              >
                Quote as a maker
              </Link>
            </div>
          </div>

          {/* The proof, not a marketing stat: a real settled fill anyone can open. */}
          <aside className="self-start border border-line bg-panel">
            <div className="flex items-center justify-between border-b border-line px-4 py-2.5">
              <span className="text-[11px] uppercase tracking-[0.13em] text-muted">Last settled fill</span>
              <span className="h-1.5 w-1.5 rounded-full bg-positive" />
            </div>
            <dl className="divide-y divide-line px-4">
              {[
                ["Pair", "FXRP / USDT0"],
                ["Size", "1.000000"],
                ["Filled at", "2.95"],
                ["Quotes seen", "2.95 · 2.99"],
              ].map(([k, v]) => (
                <div key={k} className="flex items-baseline justify-between py-2.5">
                  <dt className="text-[12px] text-faint">{k}</dt>
                  <dd className="font-mono tnum text-[12px] text-ink">{v}</dd>
                </div>
              ))}
            </dl>
            <a
              href={`https://coston2-explorer.flare.network/tx/${FILL_TX}`}
              target="_blank"
              rel="noreferrer"
              className="block border-t border-line px-4 py-2.5 font-mono text-[11px] text-accent transition-colors duration-150 hover:bg-raised"
            >
              {FILL_TX.slice(0, 18)}… ↗
            </a>
          </aside>
        </section>

        <section className="grid gap-px border-b border-line bg-line sm:grid-cols-2 lg:grid-cols-4">
          {STEPS.map((s) => (
            <div key={s.n} className="bg-base px-5 py-7">
              <span className="font-mono text-[11px] text-accent">{s.n}</span>
              <h3 className="mt-3 text-[14px] font-medium text-ink">{s.t}</h3>
              <p className="mt-2 text-[12.5px] leading-relaxed text-faint">{s.d}</p>
            </div>
          ))}
        </section>

        <footer className="flex flex-wrap items-center gap-x-5 gap-y-2 py-6 text-[11px] text-faint">
          <span>Simulated TEE · owner allowlist · no FTSO bound on this deployment</span>
          <a
            href="https://github.com/Cassxbt/fxrp-dark-rfq"
            target="_blank"
            rel="noreferrer"
            className="ml-auto transition-colors duration-150 hover:text-muted"
          >
            Source & trust model ↗
          </a>
        </footer>
      </div>
    </main>
  );
}
