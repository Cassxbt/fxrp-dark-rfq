"use client";

import { useState } from "react";
import { useAccount, useConnect, useDisconnect } from "wagmi";

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
      setError(err instanceof Error ? err.message : String(err));
    }
  }

  if (isConnected && address) {
    return (
      <div className="flex items-center gap-3 text-sm">
        <span className="font-mono text-neutral-600">
          {address.slice(0, 6)}...{address.slice(-4)}
        </span>
        <button onClick={() => disconnect()} className="text-neutral-500 underline hover:text-neutral-800">
          disconnect
        </button>
      </div>
    );
  }

  return (
    <div className="text-right">
      <button
        onClick={handleConnect}
        disabled={isPending}
        className="rounded bg-black px-4 py-2 text-sm text-white hover:bg-neutral-800 disabled:opacity-50"
      >
        {isPending ? "Connecting..." : "Connect Wallet"}
      </button>
      {error && <p className="mt-1 text-xs text-red-600">{error}</p>}
    </div>
  );
}
