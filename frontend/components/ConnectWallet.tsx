"use client";

import { useState } from "react";
import { useAccount, useConnect, useDisconnect } from "wagmi";
import { formatError } from "@/lib/formatError";

export function ConnectWallet() {
  const { address, isConnected } = useAccount();
  const { connectAsync, connectors, isPending } = useConnect();
  const { disconnect } = useDisconnect();
  const [error, setError] = useState<string | null>(null);

  async function handleConnect() {
    setError(null);
    try {
      // Using connectAsync (not the fire-and-forget connect) so a declined
      // connection or any other failure is actually catchable — the
      // previous version called connect() with no handling at all, so a
      // rejection in the injected provider produced no feedback whatsoever,
      // unlike every other action in this app.
      await connectAsync({ connector: connectors[0] });
    } catch (err) {
      setError(formatError(err));
    }
  }

  if (isConnected && address) {
    return (
      <div className="flex items-center gap-2.5 text-[12px]">
        <span className="font-mono text-muted">
          {address.slice(0, 6)}…{address.slice(-4)}
        </span>
        <button
          onClick={() => disconnect()}
          className="text-faint transition-colors duration-150 hover:text-negative"
        >
          disconnect
        </button>
      </div>
    );
  }

  return (
    <div className="relative">
      <button
        onClick={handleConnect}
        disabled={isPending}
        className="border border-line-strong px-3 py-1.5 text-[12px] font-medium text-ink transition-colors duration-150 hover:border-accent hover:text-accent disabled:opacity-50"
      >
        {isPending ? "Connecting…" : "Connect wallet"}
      </button>
      {error && (
        <p className="absolute right-0 top-full mt-1.5 w-56 text-right text-[11px] leading-snug text-negative">
          {error}
        </p>
      )}
    </div>
  );
}
