"use client";

import { useState } from "react";
import { useAccount, useSignTypedData, useWriteContract } from "wagmi";
import { parseUnits } from "viem";
import { TopBar, Panel, Field, Btn, SideToggle, Breadcrumb, Tag } from "@/components/ui";
import { EIP712_DOMAIN, QUOTE_TYPES } from "@/lib/eip712";
import { RFQ_SETTLEMENT_ADDRESS, FXRP_ADDRESS, USDT0_ADDRESS, ERC20_ABI } from "@/lib/contracts";
import { sendRfqDirect } from "@/lib/rfqClient";
import { formatError } from "@/lib/formatError";
import { quoteAmount } from "@/lib/quoteAmount";

const FXRP_DECIMALS = 6;
const USDT0_DECIMALS = 6;

export default function MakerPage() {
  const { address, isConnected } = useAccount();
  const { signTypedDataAsync } = useSignTypedData();
  const { writeContractAsync } = useWriteContract();

  // Side and size are told to the maker out of band by the taker — there is
  // no public RFQ listing (see docs/TRUST.md). The maker needs them anyway to
  // price intelligently and to know which token to approve.
  const [rfqId, setRfqId] = useState("");
  const [rfqSide, setRfqSide] = useState<"buy" | "sell">("buy");
  const [rfqSize, setRfqSize] = useState("1");
  const [price, setPrice] = useState("2.00");
  const [status, setStatus] = useState("");
  const [busy, setBusy] = useState(false);

  async function submitQuote() {
    if (!address) return;
    if (!rfqId.startsWith("0x") || rfqId.length !== 66) {
      setStatus("RFQ ID must be a 0x-prefixed 32-byte hash");
      return;
    }
    setBusy(true);
    setStatus("Preparing...");

    try {
      const sizeUnits = parseUnits(rfqSize, FXRP_DECIMALS);
      const priceWad = parseUnits(price, 18);
      // Matches the taker intent's 15-minute window. A shorter quote validity
      // is what a real maker would want (less time on the hook for a stale
      // price), but anything under the intent's lifetime means selectWinner
      // silently drops the quote as expired if the taker closes late — which
      // reads as "no qualifying quote" rather than "your quote timed out".
      const expiry = BigInt(Math.floor(Date.now() / 1000) + 900);

      // The maker's own quoted price makes this exact and knowable, unlike
      // the taker-buy case.
      if (rfqSide === "buy") {
        // Taker is buying FXRP, so the maker sells FXRP.
        setStatus("Approving FXRP...");
        await writeContractAsync({
          address: FXRP_ADDRESS,
          abi: ERC20_ABI,
          functionName: "approve",
          args: [RFQ_SETTLEMENT_ADDRESS, sizeUnits],
        });
      } else {
        // Taker is selling FXRP, so the maker pays USDT0.
        const approveAmount = quoteAmount(sizeUnits, priceWad, FXRP_DECIMALS, USDT0_DECIMALS);
        setStatus("Approving USDT0...");
        await writeContractAsync({
          address: USDT0_ADDRESS,
          abi: ERC20_ABI,
          functionName: "approve",
          args: [RFQ_SETTLEMENT_ADDRESS, approveAmount],
        });
      }

      const quote = {
        rfqId: rfqId as `0x${string}`,
        price: priceWad,
        maker: address,
        expiry,
      };

      setStatus("Sign the quote in your wallet...");
      const signature = await signTypedDataAsync({
        domain: EIP712_DOMAIN,
        types: QUOTE_TYPES,
        primaryType: "Quote",
        message: quote,
      });

      setStatus("Submitting quote...");
      const result = await sendRfqDirect("QUOTE", { data: quote, signature });
      if (result.status !== 1) {
        setStatus(`Failed: ${result.log}`);
        return;
      }
      setStatus("Quote submitted. If you win, the taker's CLOSE triggers settlement — watch your wallet balance or the contract's Filled event to confirm.");
    } catch (err) {
      setStatus(`Error: ${formatError(err)}`);
    } finally {
      setBusy(false);
    }
  }

  const submitted = status.startsWith("Quote submitted");
  const failed = status.startsWith("Failed") || status.startsWith("Error") || status.startsWith("RFQ ID must");

  return (
    <div className="lit grid-surface min-h-screen">
      <TopBar active="maker" />

      <main className="mx-auto w-full max-w-5xl px-5 py-8">
        <Breadcrumb trail={[{ label: "Desk", href: "/" }, { label: "Maker" }]} />
        <div className="mb-7">
          <h1 className="display text-[34px] text-ink">Quote an RFQ</h1>
          <p className="mt-1 text-[12px] text-faint">
            You are pricing blind: the taker&apos;s limit and every competing quote stay inside the enclave.
          </p>
        </div>

        {!isConnected ? (
          <Panel className="px-5 py-14 text-center">
            <p className="text-[13px] text-muted">Connect a wallet on Coston2 to submit a quote.</p>
          </Panel>
        ) : (
          <div className="grid gap-5 lg:grid-cols-[minmax(0,380px)_minmax(0,1fr)] lg:items-start">
            <Panel title="Quote ticket">
              <div className="space-y-4 p-4">
                <Field
                  label="RFQ ID"
                  value={rfqId}
                  onChange={setRfqId}
                  placeholder="0x…"
                  hint={<span className="text-[11px] text-faint">from the taker</span>}
                />

                <div>
                  <span className="mb-1.5 block text-[11px] uppercase tracking-[0.13em] text-muted">
                    Taker direction
                  </span>
                  <SideToggle
                    value={rfqSide}
                    onChange={setRfqSide}
                    labels={{ buy: "Taker buys", sell: "Taker sells" }}
                  />
                </div>

                <Field label="Size" value={rfqSize} onChange={setRfqSize} suffix="FXRP" />
                <Field label="Your price" value={price} onChange={setPrice} suffix="USDT0" />

                <p className="border-l border-accent-dim pl-3 text-[12px] leading-relaxed text-faint">
                  {rfqSide === "buy"
                    ? `You sell ${rfqSize || "0"} FXRP if you win. Lowest qualifying price wins.`
                    : `You pay USDT0 for ${rfqSize || "0"} FXRP if you win. Highest qualifying price wins.`}
                </p>

                <Btn onClick={submitQuote} disabled={busy} className="w-full">
                  {busy ? "Working…" : "Approve & submit quote"}
                </Btn>
              </div>
            </Panel>

            <div className="space-y-5">
              <Panel title="Status">
                {!status ? (
                  <p className="px-4 py-9 text-center text-[12px] text-faint">
                    No quote submitted yet.
                  </p>
                ) : (
                  <div className="flex items-start gap-3 p-4">
                    <span
                      className={`mt-1.5 h-1.5 w-1.5 shrink-0 rounded-full ${
                        failed ? "bg-negative" : submitted ? "bg-positive" : "bg-accent"
                      }`}
                    />
                    <p
                      className={`text-[12px] leading-relaxed ${
                        failed ? "text-negative" : submitted ? "text-ink" : "text-muted"
                      }`}
                    >
                      {status}
                    </p>
                  </div>
                )}
              </Panel>

              <Panel title="What you are not shown — and why that is the point">
                <ul className="divide-y divide-line px-4">
                  {[
                    ["The taker's limit", "You cannot shade your price up to just under it, because you do not know where it is. Quote what the flow is worth to you."],
                    ["Competing quotes", "You are not bidding against a number on a screen. Nobody can see your price and undercut it by a tick."],
                    ["Whether you lost", "If another maker wins, you are never told at what price. Your next quote stays your own opinion, not a reaction to theirs."],
                  ].map(([k, d]) => (
                    <li key={k} className="py-3">
                      <div className="flex items-center justify-between gap-3">
                        <span className="text-[13px] text-ink">{k}</span>
                        <Tag>hidden</Tag>
                      </div>
                      <p className="mt-1.5 text-[12px] leading-relaxed text-faint">{d}</p>
                    </li>
                  ))}
                </ul>
              </Panel>

              <Panel title="What the enclave does with this">
                <ol className="divide-y divide-line px-4 text-[12px] text-faint">
                  {[
                    "Recovers your address from the signature — a quote can't be filed under someone else's name.",
                    "Holds it sealed. The taker never sees your price unless you win.",
                    "On close, ranks every quote against the taker's hidden limit and picks the best.",
                    "Signs the winning fill and settles it on-chain in one transaction.",
                  ].map((t, i) => (
                    <li key={i} className="flex gap-3 py-2.5">
                      <span className="font-mono text-[11px] text-accent">{`0${i + 1}`}</span>
                      <span className="leading-relaxed">{t}</span>
                    </li>
                  ))}
                </ol>
              </Panel>
            </div>
          </div>
        )}
      </main>
    </div>
  );
}
