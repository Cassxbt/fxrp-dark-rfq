"use client";

import { useState } from "react";
import { useAccount, useSignTypedData, useWriteContract, useWatchContractEvent, useConfig } from "wagmi";
import { readContract, getPublicClient } from "wagmi/actions";
import { parseUnits, formatUnits } from "viem";
import { TopBar, Panel, Field, Btn, SideToggle, Row, Breadcrumb } from "@/components/ui";
import { EIP712_DOMAIN, RFQ_INTENT_TYPES, SIDE_TAKER_BUY, SIDE_TAKER_SELL } from "@/lib/eip712";
import { RFQ_SETTLEMENT_ADDRESS, FXRP_ADDRESS, USDT0_ADDRESS, ERC20_ABI, RFQ_SETTLEMENT_ABI } from "@/lib/contracts";
import { sendRfqDirect } from "@/lib/rfqClient";
import { hexToBytes } from "@/lib/ecies";
import { quoteAmount } from "@/lib/quoteAmount";

// Explorer-verified, not read live. The contract itself reads decimals()
// live on-chain; this only sizes the approve() call.
const FXRP_DECIMALS = 6;
const USDT0_DECIMALS = 6;

type FilledArgs = { rfqId: `0x${string}`; taker: `0x${string}`; maker: `0x${string}`; side: number; size: bigint; price: bigint };

export default function TakerPage() {
  const { address, isConnected } = useAccount();
  const { signTypedDataAsync } = useSignTypedData();
  const { writeContractAsync } = useWriteContract();
  const wagmiConfig = useConfig();

  const [side, setSide] = useState<"buy" | "sell">("buy");
  const [size, setSize] = useState("1");
  const [limitPrice, setLimitPrice] = useState("3.00");
  const [status, setStatus] = useState("");
  const [busy, setBusy] = useState(false);
  const [copied, setCopied] = useState(false);
  const [rfqId, setRfqId] = useState<`0x${string}` | null>(null);
  const [closeResult, setCloseResult] = useState<string | null>(null);
  const [filled, setFilled] = useState<FilledArgs | null>(null);
  // Captured at open time — the form's live side/size state can be edited
  // after opening (before closing), which would make the share blob below
  // silently drift from what was actually submitted on-chain.
  const [openedRfq, setOpenedRfq] = useState<{ side: "buy" | "sell"; size: string } | null>(null);

  useWatchContractEvent({
    address: RFQ_SETTLEMENT_ADDRESS,
    abi: RFQ_SETTLEMENT_ABI,
    eventName: "Filled",
    onLogs(logs) {
      const match = logs.find((l) => rfqId && l.args.rfqId === rfqId);
      if (match) setFilled(match.args as FilledArgs);
    },
  });

  async function openRfq() {
    if (!address) return;
    setBusy(true);
    setStatus("Preparing...");
    setRfqId(null);
    setFilled(null);
    setCloseResult(null);
    setOpenedRfq(null);

    const sizeUnits = parseUnits(size, FXRP_DECIMALS);
    const limitPriceWad = parseUnits(limitPrice, 18);
    const sideNum = side === "buy" ? SIDE_TAKER_BUY : SIDE_TAKER_SELL;
    // 15 min, not 5 — a three-wallet taker/maker/maker dance in a live demo
    // can easily take longer than 5 minutes, and an intent expiring mid-demo
    // is a self-inflicted failure, not a real limitation worth cutting close on.
    const expiry = BigInt(Math.floor(Date.now() / 1000) + 900);
    const rfqNonce = BigInt(Date.now()) * 1000n + BigInt(Math.floor(Math.random() * 1000));

    // Never max uint. Buy: approve the worst-case quote amount at the
    // taker's own limit, computable now since it's the taker's own number.
    // Sell: approve exact size of FXRP.
    try {
      if (sideNum === SIDE_TAKER_BUY) {
        const approveAmount = quoteAmount(sizeUnits, limitPriceWad, FXRP_DECIMALS, USDT0_DECIMALS);
        setStatus("Approving USDT0...");
        await writeContractAsync({
          address: USDT0_ADDRESS,
          abi: ERC20_ABI,
          functionName: "approve",
          args: [RFQ_SETTLEMENT_ADDRESS, approveAmount],
        });
      } else {
        setStatus("Approving FXRP...");
        await writeContractAsync({
          address: FXRP_ADDRESS,
          abi: ERC20_ABI,
          functionName: "approve",
          args: [RFQ_SETTLEMENT_ADDRESS, sizeUnits],
        });
      }

      const intent = {
        side: sideNum,
        size: sizeUnits,
        limitPrice: limitPriceWad,
        taker: address,
        expiry,
        rfqNonce,
      };

      setStatus("Sign the RFQ intent in your wallet...");
      const signature = await signTypedDataAsync({
        domain: EIP712_DOMAIN,
        types: RFQ_INTENT_TYPES,
        primaryType: "RfqIntent",
        message: intent,
      });

      setStatus("Opening RFQ...");
      const result = await sendRfqDirect("OPEN", { data: intent, signature });
      if (result.status !== 1) {
        setStatus(`Failed: ${result.log}`);
        return;
      }
      const decoded = JSON.parse(new TextDecoder().decode(hexToBytes(result.data)));
      setRfqId(decoded.rfqId);
      setOpenedRfq({ side, size });
      setStatus("RFQ opened.");
    } catch (err) {
      setStatus(`Error: ${err instanceof Error ? err.message : String(err)}`);
    } finally {
      setBusy(false);
    }
  }

  // The event watcher only fires for logs that arrive while it's subscribed.
  // Once settled() is true the fill is a fact on-chain, so read the Filled
  // event back directly instead of leaving the confirmation card dependent on
  // the watcher having been listening at the right moment.
  async function loadFilled(id: `0x${string}`) {
    const client = getPublicClient(wagmiConfig);
    if (!client) return;
    // Coston2's public RPC caps eth_getLogs at 30 blocks per request.
    const latest = await client.getBlockNumber();
    const logs = await client.getContractEvents({
      address: RFQ_SETTLEMENT_ADDRESS,
      abi: RFQ_SETTLEMENT_ABI,
      eventName: "Filled",
      args: { rfqId: id },
      fromBlock: latest - 25n,
      toBlock: "latest",
    });
    const last = logs[logs.length - 1];
    if (last) setFilled(last.args as FilledArgs);
  }

  async function closeRfq() {
    if (!rfqId) return;
    setBusy(true);
    setStatus("Closing RFQ...");
    try {
      const result = await sendRfqDirect("CLOSE", hexToBytes(rfqId));
      if (result.status !== 1) {
        setCloseResult(`Failed: ${result.log}`);
        return;
      }
      const decoded = JSON.parse(new TextDecoder().decode(hexToBytes(result.data)));
      setCloseResult(JSON.stringify(decoded));

      if (!decoded.matched) {
        setStatus("Closed — no qualifying quote, nothing to settle.");
        return;
      }

      // Settlement is submitted asynchronously (the FCC framework's 2s
      // response timeout doesn't fit a chain round trip — see docs/TRUST.md),
      // so poll settled() rather than trusting the CLOSE response as proof.
      setStatus("Closed — waiting for on-chain settlement (up to ~30s)...");
      for (let i = 0; i < 15; i++) {
        await new Promise((r) => setTimeout(r, 2000));
        const isSettled = await readContract(wagmiConfig, {
          address: RFQ_SETTLEMENT_ADDRESS,
          abi: RFQ_SETTLEMENT_ABI,
          functionName: "settled",
          args: [rfqId],
        });
        if (isSettled) {
          setStatus("Settled on-chain.");
          await loadFilled(rfqId);
          return;
        }
      }
      setStatus(
        "No Filled after ~30s — settlement likely reverted (check allowance, balance, or the attested-signer whitelist). Not a UI bug: the chain itself never recorded a fill.",
      );
    } catch (err) {
      setCloseResult(`Error: ${err instanceof Error ? err.message : String(err)}`);
    } finally {
      setBusy(false);
    }
  }

  const shareLine = openedRfq
    ? `RFQ ${rfqId} — taker is ${openedRfq.side === "buy" ? "BUYING" : "SELLING"} ${openedRfq.size} FXRP`
    : "";

  const stage = filled ? 3 : rfqId ? 2 : 1;

  return (
    <div className="lit grid-surface min-h-screen">
      <TopBar active="taker" />

      <main className="mx-auto w-full max-w-5xl px-5 py-8">
        <Breadcrumb trail={[{ label: "Desk", href: "/" }, { label: "Taker" }]} />
        <div className="mb-7 flex items-end justify-between gap-4">
          <div>
            <h1 className="display text-[34px] text-ink">Open an RFQ</h1>
            <p className="mt-1 text-[12px] text-faint">
              Your limit price is sealed to the enclave. Makers price the trade without seeing it.
            </p>
          </div>
          <ol className="hidden items-center gap-2 text-[11px] text-faint sm:flex">
            {["Intent", "Quotes", "Settled"].map((s, i) => (
              <li key={s} className="flex items-center gap-2">
                <span className={i + 1 <= stage ? "text-accent" : ""}>{s}</span>
                {i < 2 && <span className="text-line-strong">/</span>}
              </li>
            ))}
          </ol>
        </div>

        {!isConnected ? (
          <Panel className="px-5 py-14 text-center">
            <p className="text-[13px] text-muted">Connect a wallet on Coston2 to open an RFQ.</p>
          </Panel>
        ) : (
          <div className="grid gap-5 lg:grid-cols-[minmax(0,380px)_minmax(0,1fr)] lg:items-start">
            <Panel title="Order ticket">
              <div className="space-y-4 p-4">
                <SideToggle value={side} onChange={setSide} labels={{ buy: "Buy FXRP", sell: "Sell FXRP" }} />

                <Field label="Size" value={size} onChange={setSize} suffix="FXRP" />

                <Field
                  label="Limit price"
                  value={limitPrice}
                  onChange={setLimitPrice}
                  suffix="USDT0"
                  hint={
                    <span className="flex items-center gap-1.5 text-[11px] uppercase tracking-[0.1em] text-accent">
                      <svg viewBox="0 0 10 12" className="h-2.5 w-2.5 fill-none stroke-current" strokeWidth="1.3">
                        <rect x="0.65" y="5" width="8.7" height="6.35" rx="1" />
                        <path d="M2.6 5V3.1a2.4 2.4 0 0 1 4.8 0V5" />
                      </svg>
                      Sealed
                    </span>
                  }
                />

                <p className="border-l border-accent-dim pl-3 text-[12px] leading-relaxed text-faint">
                  {side === "buy"
                    ? `Approves at most ${limitPrice || "0"} USDT0 per FXRP — never max uint. You fill at the best maker price at or below your limit.`
                    : `Approves exactly ${size || "0"} FXRP. You fill at the best maker price at or above your limit.`}
                </p>

                <Btn onClick={openRfq} disabled={busy} className="w-full">
                  {busy ? "Working…" : "Approve & open RFQ"}
                </Btn>
              </div>
            </Panel>

            <div className="space-y-5">
              <Panel
                title="Lifecycle"
                aside={
                  status ? (
                    <span className="max-w-[22rem] truncate text-[11px] text-muted" title={status}>
                      {status}
                    </span>
                  ) : (
                    <span className="text-[11px] text-faint">idle</span>
                  )
                }
              >
                {!rfqId ? (
                  <p className="px-4 py-9 text-center text-[12px] text-faint">
                    No open RFQ. Submit the ticket to seal an intent.
                  </p>
                ) : (
                  <div className="p-4">
                    <p className="mb-2 text-[11px] text-faint">
                      Send this to your makers — there is no public listing to discover it.
                    </p>
                    <div className="flex items-stretch border border-line bg-base">
                      <code className="min-w-0 flex-1 truncate px-3 py-2.5 font-mono text-[12px] text-ink">
                        {shareLine}
                      </code>
                      <button
                        onClick={() => {
                          navigator.clipboard?.writeText(shareLine);
                          setCopied(true);
                          setTimeout(() => setCopied(false), 1400);
                        }}
                        className="shrink-0 border-l border-line px-3 text-[11px] text-muted transition-colors duration-150 hover:text-accent"
                      >
                        {copied ? "copied" : "copy"}
                      </button>
                    </div>

                    <div className="mt-4 flex items-center gap-3">
                      <Btn onClick={closeRfq} variant="ghost" disabled={busy}>
                        Close & match
                      </Btn>
                      <span className="text-[11px] text-faint">
                        Matches against every quote received so far.
                      </span>
                    </div>

                    {closeResult && (
                      <pre className="mt-4 overflow-x-auto border border-line bg-base px-3 py-2.5 font-mono text-[11px] leading-relaxed text-muted">
                        {closeResult}
                      </pre>
                    )}
                  </div>
                )}
              </Panel>

              {filled && (
                <Panel
                  title="Filled"
                  className="border-positive/35"
                  aside={<span className="text-[11px] text-positive">confirmed on-chain</span>}
                >
                  <div className="divide-y divide-line px-4">
                    <Row k="RFQ" v={`${filled.rfqId.slice(0, 14)}…`} />
                    <Row k="Maker" v={`${filled.maker.slice(0, 10)}…${filled.maker.slice(-6)}`} />
                    <Row
                      k="Side"
                      v={filled.side === 0 ? "taker buy" : "taker sell"}
                      tone={filled.side === 0 ? "text-positive" : "text-negative"}
                    />
                    <Row k="Size" v={`${formatUnits(filled.size, FXRP_DECIMALS)} FXRP`} />
                    <Row k="Price" v={`${formatUnits(filled.price, 18)} USDT0`} tone="text-accent" />
                  </div>
                </Panel>
              )}
            </div>
          </div>
        )}
      </main>
    </div>
  );
}
