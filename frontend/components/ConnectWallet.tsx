"use client";

import { useAccount, useConnect, useDisconnect } from "wagmi";

export function ConnectWallet() {
  const { address, isConnected } = useAccount();
  const { connect, connectors, isPending } = useConnect();
  const { disconnect } = useDisconnect();

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
    <button
      onClick={() => connect({ connector: connectors[0] })}
      disabled={isPending}
      className="rounded bg-black px-4 py-2 text-sm text-white hover:bg-neutral-800 disabled:opacity-50"
    >
      {isPending ? "Connecting..." : "Connect Wallet"}
    </button>
  );
}
