"use client";

import { useState } from "react";
import { useAccount, useSignTypedData, useWriteContract, useWatchContractEvent, useConfig } from "wagmi";
import { readContract, getPublicClient } from "wagmi/actions";
import { parseUnits } from "viem";
import { ConnectWallet } from "@/components/ConnectWallet";
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
    }
  }

  return (
    <main className="mx-auto w-full max-w-xl p-8 space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold">Taker — open an RFQ</h1>
        <ConnectWallet />
      </div>

      {!isConnected ? (
        <p className="text-neutral-500">Connect a wallet on Coston2 to continue.</p>
      ) : (
        <>
          <div className="space-y-3">
            <div className="flex gap-4">
              <label className="flex items-center gap-2 text-sm">
                <input type="radio" checked={side === "buy"} onChange={() => setSide("buy")} /> Buy FXRP
              </label>
              <label className="flex items-center gap-2 text-sm">
                <input type="radio" checked={side === "sell"} onChange={() => setSide("sell")} /> Sell FXRP
              </label>
            </div>
            <label className="block text-sm">
              Size (FXRP)
              <input
                className="mt-1 w-full rounded border px-3 py-2"
                value={size}
                onChange={(e) => setSize(e.target.value)}
              />
            </label>
            <label className="block text-sm">
              Limit price (USDT0 per FXRP) — stays inside the TEE, never revealed to makers or on-chain
              <input
                className="mt-1 w-full rounded border px-3 py-2"
                value={limitPrice}
                onChange={(e) => setLimitPrice(e.target.value)}
              />
            </label>
            <button onClick={openRfq} className="rounded bg-black px-4 py-2 text-white hover:bg-neutral-800">
              Open RFQ
            </button>
          </div>

          {status && <p className="text-sm text-neutral-700 break-all">{status}</p>}

          {rfqId && (
            <div className="space-y-3 rounded border p-4">
              <div>
                <p className="text-xs font-medium text-neutral-500">
                  Share this with makers — there&apos;s no public listing:
                </p>
                <p className="mt-1 break-all rounded bg-neutral-50 p-2 font-mono text-xs">
                  RFQ {rfqId} — {openedRfq?.side === "buy" ? "taker is BUYING" : "taker is SELLING"} {openedRfq?.size} FXRP
                </p>
              </div>
              <button onClick={closeRfq} className="rounded border px-4 py-2 text-sm hover:bg-neutral-50">
                Close RFQ (match against submitted quotes)
              </button>
              {closeResult && <p className="break-all text-sm">{closeResult}</p>}
            </div>
          )}

          {filled && (
            <div className="rounded border border-green-600 bg-green-50 p-4 text-sm">
              <p className="font-semibold text-green-800">Filled event confirmed on-chain</p>
              <pre className="mt-2 overflow-x-auto text-xs">
                {JSON.stringify(filled, (_k, v) => (typeof v === "bigint" ? v.toString() : v), 2)}
              </pre>
            </div>
          )}
        </>
      )}
    </main>
  );
}
