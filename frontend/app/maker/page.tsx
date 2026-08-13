"use client";

import { useState } from "react";
import { useAccount, useSignTypedData, useWriteContract } from "wagmi";
import { parseUnits } from "viem";
import { ConnectWallet } from "@/components/ConnectWallet";
import { EIP712_DOMAIN, QUOTE_TYPES } from "@/lib/eip712";
import { RFQ_SETTLEMENT_ADDRESS, FXRP_ADDRESS, USDT0_ADDRESS, ERC20_ABI } from "@/lib/contracts";
import { sendRfqDirect } from "@/lib/rfqClient";

const FXRP_DECIMALS = 6;
const USDT0_DECIMALS = 6;

export default function MakerPage() {
  const { address, isConnected } = useAccount();
  const { signTypedDataAsync } = useSignTypedData();
  const { writeContractAsync } = useWriteContract();

  // Side and size are told to the maker out of band by the taker — there is
  // no public RFQ listing (BUILD-SPEC.md §5, "Public maker discovery/routing"
  // is explicitly out of scope for this MVP). The maker needs them anyway to
  // price intelligently and to know which token to approve.
  const [rfqId, setRfqId] = useState("");
  const [rfqSide, setRfqSide] = useState<"buy" | "sell">("buy");
  const [rfqSize, setRfqSize] = useState("1");
  const [price, setPrice] = useState("2.00");
  const [status, setStatus] = useState("");

  async function submitQuote() {
    if (!address) return;
    if (!rfqId.startsWith("0x") || rfqId.length !== 66) {
      setStatus("RFQ ID must be a 0x-prefixed 32-byte hash");
      return;
    }
    setStatus("Preparing...");

    try {
      const sizeUnits = parseUnits(rfqSize, FXRP_DECIMALS);
      const priceWad = parseUnits(price, 18);
      const expiry = BigInt(Math.floor(Date.now() / 1000) + 300); // 5 min

      // Approve per BUILD-SPEC.md §2.1's role-specific rule — the maker's own
      // quoted price makes this exact and knowable, unlike the taker-buy case.
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
        const quoteAmount =
          (sizeUnits * priceWad * 10n ** BigInt(USDT0_DECIMALS)) / (10n ** BigInt(FXRP_DECIMALS) * 10n ** 18n);
        setStatus("Approving USDT0...");
        await writeContractAsync({
          address: USDT0_ADDRESS,
          abi: ERC20_ABI,
          functionName: "approve",
          args: [RFQ_SETTLEMENT_ADDRESS, quoteAmount],
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
      setStatus(`Error: ${err instanceof Error ? err.message : String(err)}`);
    }
  }

  return (
    <main className="mx-auto max-w-xl p-8 space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold">Maker — quote on an RFQ</h1>
        <ConnectWallet />
      </div>

      {!isConnected ? (
        <p className="text-neutral-500">Connect a wallet on Coston2 to continue.</p>
      ) : (
        <div className="space-y-3">
          <label className="block text-sm">
            RFQ ID (from the taker)
            <input
              className="mt-1 w-full rounded border px-3 py-2 font-mono text-xs"
              value={rfqId}
              onChange={(e) => setRfqId(e.target.value)}
              placeholder="0x..."
            />
          </label>
          <div className="flex gap-4">
            <label className="flex items-center gap-2 text-sm">
              <input type="radio" checked={rfqSide === "buy"} onChange={() => setRfqSide("buy")} /> Taker is buying
            </label>
            <label className="flex items-center gap-2 text-sm">
              <input type="radio" checked={rfqSide === "sell"} onChange={() => setRfqSide("sell")} /> Taker is selling
            </label>
          </div>
          <label className="block text-sm">
            Size (FXRP, told to you by the taker)
            <input
              className="mt-1 w-full rounded border px-3 py-2"
              value={rfqSize}
              onChange={(e) => setRfqSize(e.target.value)}
            />
          </label>
          <label className="block text-sm">
            Your quoted price (USDT0 per FXRP)
            <input className="mt-1 w-full rounded border px-3 py-2" value={price} onChange={(e) => setPrice(e.target.value)} />
          </label>
          <button onClick={submitQuote} className="rounded bg-black px-4 py-2 text-white hover:bg-neutral-800">
            Submit Quote
          </button>
          {status && <p className="break-all text-sm text-neutral-700">{status}</p>}
        </div>
      )}
    </main>
  );
}
