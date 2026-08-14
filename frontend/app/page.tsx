import Link from "next/link";
import { Tag, Medallion } from "@/components/ui";
import { ExtStatus } from "@/components/ExtStatus";
import {
  RFQ_SETTLEMENT_ADDRESS,
  FXRP_ADDRESS,
  USDT0_ADDRESS,
} from "@/lib/contracts";

const FILL_TX = "0xe158ffe70bd1df2790ca3bc09c501cf214f6c7a7406872882361698551a7a8e9";
const FILL_TX_SELL = "0x92d60cc432e423fc6f37cd3de95ab3f7620efdf4f64f2bbe17a13652c1cbed01";

// Both directions, settled on-chain. Two receipts rather than one: a single
// fill reads as a one-off, and the pair shows the winner flips with the side.
const FILLS = [
  { side: "buy", size: "1.000000", price: "2.95", quotes: "2.95 · 2.99", tx: FILL_TX },
  { side: "sell", size: "1.000000", price: "2.60", quotes: "2.55 · 2.60", tx: FILL_TX_SELL },
] as const;
const EXPLORER = "https://coston2-explorer.flare.network";
const REPO = "https://github.com/Cassxbt/fxrp-dark-rfq";

function Section({
  id,
  eyebrow,
  title,
  lede,
  children,
}: {
  id: string;
  eyebrow: string;
  title: string;
  lede?: string;
  children?: React.ReactNode;
}) {
  return (
    <section id={id} className="border-b border-line py-16">
      <p className="eyebrow-ruled font-mono text-[11px] uppercase tracking-[0.22em] text-accent">
        {eyebrow}
      </p>
      <h2 className="display mt-4 max-w-2xl text-[34px] leading-[1.15] text-ink">{title}</h2>
      {lede && (
        <p className="mt-3 max-w-2xl font-display text-[19px] italic leading-relaxed text-accent-lt/85">
          {lede}
        </p>
      )}
      {children && <div className="mt-9">{children}</div>}
    </section>
  );
}

/** Visibility matrix — the clearest single statement of what this system does. */
const VISIBILITY = [
  { field: "Taker's limit price", taker: "yes", makers: "no", chain: "no", note: "3.00 — never leaves the enclave" },
  { field: "Your own quote", taker: "no", makers: "own", chain: "no", note: "makers cannot see each other" },
  { field: "Losing quote", taker: "no", makers: "no", chain: "no", note: "2.99 was never published anywhere" },
  { field: "Side and size", taker: "yes", makers: "yes", chain: "yes", note: "shared out of band, then on-chain" },
  { field: "Winning maker + price", taker: "yes", makers: "yes", chain: "yes", note: "the Filled event, after the fact" },
];

function Cell({ v }: { v: string }) {
  if (v === "yes") return <span className="font-mono text-[12px] text-positive">visible</span>;
  if (v === "own") return <span className="font-mono text-[12px] text-accent">own only</span>;
  return <span className="font-mono text-[12px] text-faint">hidden</span>;
}

export default function Home() {
  return (
    <main className="lit min-h-screen">
      <div className="mx-auto w-full max-w-5xl px-5">
        <header className="sticky top-0 z-10 flex h-14 items-center gap-2.5 border-b border-line bg-base/85 backdrop-blur">
          <span className="h-2.5 w-2.5 rounded-[2px] bg-accent" />
          <span className="display text-[19px] tracking-[0.02em]">FXRP Dark RFQ</span>
          <nav className="ml-6 hidden gap-4 text-[12px] text-faint sm:flex">
            <a href="#how" className="transition-colors duration-150 hover:text-ink">How it works</a>
            <a href="#visibility" className="transition-colors duration-150 hover:text-ink">Who sees what</a>
            <a href="#trust" className="transition-colors duration-150 hover:text-ink">Trust model</a>
          </nav>
          <span className="ml-auto">
            <ExtStatus />
          </span>
        </header>

        {/* Hero */}
        <section className="grid gap-10 border-b border-line py-16 md:grid-cols-[1.35fr_1fr] md:gap-14 md:py-20">
          <div>
            <p className="eyebrow-ruled mb-6 font-mono text-[11px] uppercase tracking-[0.28em] text-accent">
              Sealed-bid OTC
            </p>
            <h1 className="display text-[42px] leading-[1.08] text-ink sm:text-[56px]">
              Trade FXRP without
              <br />
              showing your hand.
            </h1>
            <p className="mt-6 max-w-lg font-display text-[19px] italic leading-relaxed text-accent-lt/85">
              Ask five desks for a price and you have told five desks what you want.
            </p>
            <p className="mt-4 max-w-md text-[14px] leading-relaxed text-muted">
              This one takes your limit, seals it inside a TEE, collects quotes that cannot see each
              other, and settles only the winner on-chain.
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

          <aside className="self-start overflow-hidden rounded-[10px] border border-line bg-panel">
            <div className="flex items-center justify-between border-b border-line px-4 py-3">
              <span className="font-mono text-[11px] uppercase tracking-[0.16em] text-muted">
                Settled fills · FXRP/USDT0
              </span>
              <span className="h-1.5 w-1.5 rounded-full bg-positive" />
            </div>

            {FILLS.map((f) => (
              <a
                key={f.tx}
                href={`${EXPLORER}/tx/${f.tx}`}
                target="_blank"
                rel="noreferrer"
                className="block border-b border-line px-4 py-3 transition-colors duration-150 last:border-b-0 hover:bg-raised"
              >
                <div className="flex items-baseline justify-between gap-3">
                  <span
                    className={`font-mono text-[11px] uppercase tracking-[0.16em] ${
                      f.side === "buy" ? "text-positive" : "text-negative"
                    }`}
                  >
                    {f.side}
                  </span>
                  <span className="font-mono tnum text-[14px] text-ink">
                    {f.size} <span className="text-faint">@</span> {f.price}
                  </span>
                </div>
                <div className="mt-1.5 flex items-baseline justify-between gap-3">
                  <span className="text-[11px] text-faint">quotes {f.quotes}</span>
                  <span className="font-mono text-[11px] text-accent">{f.tx.slice(0, 12)}… ↗</span>
                </div>
              </a>
            ))}

            <p className="border-t border-line px-4 py-2.5 text-[11px] leading-relaxed text-faint">
              Lowest qualifying quote wins a buy, highest wins a sell. Both settled on Coston2.
            </p>
          </aside>
        </section>

        {/* Problem */}
        <Section
          id="problem"
          eyebrow="The problem"
          title="Price discovery costs you the information you were trying to protect."
          lede="Every venue leaks something different. The leak is not incidental — it is how the venue works."
        >
          <div className="grid gap-5 md:grid-cols-3">
            {[
              {
                kicker: "Lit venue",
                h: "Public order book",
                d: "Your resting bid is the signal. Size and price are visible to everyone, including whoever wants to trade ahead of you.",
                leak: "Size + price, before the trade",
              },
              {
                kicker: "Relationship",
                h: "Calling desks",
                d: "Each desk you ask learns your direction and size. Ask enough of them and the market knows your intent before anyone fills you.",
                leak: "Intent, to every counterparty",
              },
              {
                kicker: "Sealed",
                h: "This desk",
                d: "One sealed intent. Makers quote blind, against each other, without seeing your limit or their competition.",
                leak: "The fill, after it settled",
                good: true,
              },
            ].map((c) => (
              <div
                key={c.h}
                className="lift flex flex-col rounded-[10px] border border-t-2 border-line bg-panel px-6 py-7"
                style={{ borderTopColor: c.good ? "var(--color-accent)" : "var(--color-negative)" }}
              >
                <p className="font-mono text-[11px] uppercase tracking-[0.16em] text-faint">{c.kicker}</p>
                <h3 className={`display mt-2 text-[24px] ${c.good ? "text-accent" : "text-ink"}`}>{c.h}</h3>
                <p className="prose-serif mt-3 flex-1 text-muted">{c.d}</p>
                <div className="mt-6 border-t border-line pt-4">
                  <Tag tone={c.good ? "positive" : "negative"}>{c.leak}</Tag>
                </div>
              </div>
            ))}
          </div>
        </Section>

        {/* How it works */}
        <Section
          id="how"
          eyebrow="How it works"
          title="Matching happens off-chain inside the enclave. Only settlement is public."
          lede="Intents and quotes are EIP-712 signed, then ECIES-encrypted to the TEE's own key. The extension recovers each signer itself — nobody can file a quote under someone else's address."
        >
          <div className="overflow-hidden rounded-[10px] border border-line bg-panel">
            <div className="flex items-center gap-2.5 border-b border-line px-6 py-3">
              <span className="h-1.5 w-1.5 rounded-full bg-accent" />
              <span className="font-mono text-[11px] uppercase tracking-[0.16em] text-accent">
                Sealed inside the TEE
              </span>
            </div>
            <ol className="divide-y divide-line">
              {[
                {
                  n: "01",
                  t: "Taker seals an intent",
                  d: "Side, size, limit price and an expiry, signed and encrypted client-side. The extension recovers the taker's address from the signature and derives a deterministic rfqId.",
                },
                {
                  n: "02",
                  t: "Makers quote blind",
                  d: "Each quote is bound to that rfqId inside the signed struct, so a captured quote cannot be replayed onto a different RFQ. One live quote per maker.",
                },
                {
                  n: "03",
                  t: "Close and rank",
                  d: "Expired quotes are dropped, then the best qualifying price wins — lowest at or below the limit on a buy, highest at or above it on a sell. Ties break to whoever quoted first.",
                },
              ].map((s) => (
                <li key={s.n} className="flex gap-5 px-6 py-6">
                  <Medallion>{s.n}</Medallion>
                  <div>
                    <h3 className="display text-[19px] text-ink">{s.t}</h3>
                    <p className="prose-serif mt-2 max-w-2xl text-muted">{s.d}</p>
                  </div>
                </li>
              ))}
            </ol>

            <div className="flex items-center gap-2.5 border-y border-line bg-raised px-6 py-3">
              <span className="h-1.5 w-1.5 rounded-full bg-positive" />
              <span className="font-mono text-[11px] uppercase tracking-[0.16em] text-positive">
                Public on Coston2
              </span>
            </div>
            <div className="flex gap-5 px-6 py-6">
              <Medallion>04</Medallion>
              <div>
                <h3 className="display text-[19px] text-ink">Atomic settlement</h3>
                <p className="prose-serif mt-2 max-w-2xl text-muted">
                  The enclave signs the winning fill with its attested key. The contract verifies that
                  signature, then moves both ERC-20 legs in one transaction and emits{" "}
                  <code className="font-mono text-[12px] text-ink">Filled</code>. Nothing about the
                  losing quotes is ever written.
                </p>
              </div>
            </div>
          </div>
        </Section>

        {/* Visibility matrix */}
        <Section
          id="visibility"
          eyebrow="Who sees what"
          title="The privacy claim, stated precisely."
          lede="Taken from the fill linked above: a taker buying 1 FXRP with a 3.00 limit, against makers quoting 2.95 and 2.99."
        >
          <div className="overflow-x-auto border border-line bg-panel">
            <table className="w-full min-w-[620px] border-collapse">
              <thead>
                <tr className="border-b border-line text-left">
                  <th className="px-4 py-2.5 text-[11px] font-medium uppercase tracking-[0.13em] text-muted">Data</th>
                  <th className="px-4 py-2.5 text-[11px] font-medium uppercase tracking-[0.13em] text-muted">Taker</th>
                  <th className="px-4 py-2.5 text-[11px] font-medium uppercase tracking-[0.13em] text-muted">Makers</th>
                  <th className="px-4 py-2.5 text-[11px] font-medium uppercase tracking-[0.13em] text-muted">Chain</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-line">
                {VISIBILITY.map((r) => (
                  <tr key={r.field}>
                    <td className="px-4 py-3">
                      <span className="text-[13px] text-ink">{r.field}</span>
                      <span className="mt-0.5 block text-[12px] text-faint">{r.note}</span>
                    </td>
                    <td className="px-4 py-3"><Cell v={r.taker} /></td>
                    <td className="px-4 py-3"><Cell v={r.makers} /></td>
                    <td className="px-4 py-3"><Cell v={r.chain} /></td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <p className="mt-4 max-w-2xl text-[12px] leading-relaxed text-faint">
            One caveat worth stating plainly: the transport endpoint is unauthenticated on this
            deployment, and opening an RFQ returns its id, side and size in cleartext to whoever
            calls it. The limit price and the losing quotes are the parts that genuinely never leave
            the enclave.
          </p>
        </Section>

        {/* Worked example */}
        <Section
          id="example"
          eyebrow="A real fill, walked through"
          title="You were willing to pay 3.00. Nobody learned that. You paid 2.95."
          lede="Every number here is read from the transaction linked in the header — not an illustration."
        >
          <div className="grid gap-px bg-line md:grid-cols-4">
            {[
              { k: "Taker limit", v: "3.00", s: "sealed", tone: "text-accent" },
              { k: "Maker A", v: "2.95", s: "won", tone: "text-positive" },
              { k: "Maker B", v: "2.99", s: "never published", tone: "text-faint" },
              { k: "Settled at", v: "2.95", s: "on-chain", tone: "text-ink" },
            ].map((c) => (
              <div key={c.k} className="bg-panel px-5 py-7 text-center">
                <p className={`display tnum text-[42px] leading-none ${c.tone}`}>{c.v}</p>
                <p className="mt-3 font-mono text-[11px] uppercase tracking-[0.16em] text-muted">{c.k}</p>
                <p className="mt-1.5 text-[12px] text-faint">{c.s}</p>
              </div>
            ))}
          </div>
          <p className="mt-5 max-w-2xl text-[13px] leading-relaxed text-muted">
            Maker B never learned it lost to 2.95, so its next quote is not calibrated against Maker A.
            The taker never revealed the extra 0.05 of room it had. That gap is the entire product.
            The sell fill linked above is the same story inverted — there the taker wanted the
            highest bid, and the maker quoting 2.60 beat the one at 2.55.
          </p>
        </Section>

        {/* Why a TEE */}
        <Section
          id="why"
          eyebrow="Why a TEE"
          title="Sealed matching is a solved problem four ways. Three of them don't fit an RFQ."
          lede="A quote is worthless a minute later, so the constraint is latency, not elegance."
        >
          <div className="overflow-hidden rounded-[10px] border border-line bg-panel">
            {[
              {
                h: "ZK proof of correct matching",
                v: "Strongest guarantee. But the prover has to run before anyone can settle, and an RFQ quote is stale by then.",
                verdict: "too slow",
                tone: "text-faint",
              },
              {
                h: "MPC across makers",
                v: "No single trusted party, but every maker has to stay online and participate in the protocol to produce a match.",
                verdict: "too coordinated",
                tone: "text-faint",
              },
              {
                h: "A trusted operator",
                v: "Fast and simple. Also just asks you to believe the operator didn't look at your limit price.",
                verdict: "no guarantee",
                tone: "text-negative",
              },
              {
                h: "TEE on Flare Confidential Compute",
                v: "Matches in milliseconds, and the enclave signs the fill with an attested key the settlement contract verifies on the same chain. No bridge, no prover, no external committee between matching and settlement.",
                verdict: "chosen",
                tone: "text-positive",
              },
            ].map((r) => (
              <div key={r.h} className="border-b border-line px-4 py-5 transition-colors duration-150 last:border-b-0 hover:bg-raised">
                <div className="flex flex-wrap items-center justify-between gap-x-4 gap-y-2">
                  <h3 className="display text-[19px] text-ink">{r.h}</h3>
                  <Tag tone={r.verdict === "chosen" ? "positive" : r.verdict === "no guarantee" ? "negative" : "neutral"}>
                    {r.verdict}
                  </Tag>
                </div>
                <p className="prose-serif mt-2 max-w-2xl text-muted">{r.v}</p>
              </div>
            ))}
          </div>
          <p className="mt-4 max-w-2xl text-[12px] leading-relaxed text-faint">
            The honest trade: a TEE buys that speed with a hardware trust assumption a ZK proof
            wouldn&apos;t need. And on this deployment the enclave is simulated, so even that
            assumption isn&apos;t being enforced yet — see below.
          </p>
        </Section>

        {/* Trust model */}
        <Section
          id="trust"
          eyebrow="Trust model"
          title="What this prototype does not do."
          lede="This is a hackathon build. The interesting parts are real; these parts are not, and pretending otherwise would be the fastest way to lose your trust."
        >
          <ul className="grid gap-5 sm:grid-cols-2">
            {[
              ["Simulated TEE", "Runs FCC's sanctioned simulated mode. The signing key's integrity rests on the process, not on hardware attestation."],
              ["Owner allowlist", "The contract checks the signer against an owner-controlled allowlist, not a live TeeExtensionRegistry read."],
              ["No FTSO bound", "The contract supports an oracle price bound and has tests for it. It is switched off here — ftso() is the zero address."],
              ["Unauthenticated close", "Anyone holding an RFQ id can trigger the match. There is no listing to discover ids from, but it is not taker-only."],
              ["In-memory book", "Open RFQs live in the extension's memory. A restart forgets them."],
              ["Non-binding quotes", "Settlement is transferFrom-based. If the winner pulled their allowance, the fill simply reverts."],
            ].map(([h, d]) => (
              <li key={h} className="lift rounded-[10px] border border-line bg-panel px-6 py-6">
                <h3 className="display text-[19px] text-ink">{h}</h3>
                <p className="prose-serif mt-2 text-muted">{d}</p>
              </li>
            ))}
          </ul>
        </Section>

        {/* Verify */}
        <Section
          id="verify"
          eyebrow="Verify it yourself"
          title="Everything above resolves to something you can open."
        >
          <div className="border border-line bg-panel">
            {[
              ["Settlement contract", RFQ_SETTLEMENT_ADDRESS, `${EXPLORER}/address/${RFQ_SETTLEMENT_ADDRESS}`],
              ["FXRP", FXRP_ADDRESS, `${EXPLORER}/address/${FXRP_ADDRESS}`],
              ["USDT0", USDT0_ADDRESS, `${EXPLORER}/address/${USDT0_ADDRESS}`],
              ["Proof-of-fill — taker buy", FILL_TX, `${EXPLORER}/tx/${FILL_TX}`],
              ["Proof-of-fill — taker sell", FILL_TX_SELL, `${EXPLORER}/tx/${FILL_TX_SELL}`],
            ].map(([k, v, href]) => (
              <a
                key={k}
                href={href}
                target="_blank"
                rel="noreferrer"
                className="flex flex-wrap items-baseline justify-between gap-2 border-b border-line px-4 py-3 transition-colors duration-150 last:border-b-0 hover:bg-raised"
              >
                <span className="text-[12px] text-muted">{k}</span>
                <span className="font-mono text-[12px] text-accent">
                  {v.slice(0, 22)}… ↗
                </span>
              </a>
            ))}
          </div>
          <p className="mt-4 text-[12px] text-faint">
            The matcher is{" "}
            <code className="font-mono text-[12px] text-muted">extension/go/internal/extension/rfq.go</code>;
            settlement is{" "}
            <code className="font-mono text-[12px] text-muted">contracts/src/RfqSettlement.sol</code>. Both
            test suites run in CI.
          </p>
        </Section>

        <footer className="flex flex-wrap items-center gap-x-5 gap-y-2 py-8 text-[11px] text-faint">
          <span>Built for Flare Summer Signal · Bounty 2 (Confidential Compute)</span>
          <a
            href={REPO}
            target="_blank"
            rel="noreferrer"
            className="ml-auto transition-colors duration-150 hover:text-muted"
          >
            Source &amp; trust model ↗
          </a>
        </footer>
      </div>
    </main>
  );
}
