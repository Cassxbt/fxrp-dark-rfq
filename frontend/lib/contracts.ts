// Addresses verified live on Coston2 — see the repo's commit history for how
// each was confirmed (explorer API for tokens, `cast create`/`forge create`
// output for the deployed contract).
export const RFQ_SETTLEMENT_ADDRESS = "0xaBf47C48c00DDa806f1d9243c936A8153C7E6FcE" as const;
export const FXRP_ADDRESS = "0x0b6A3645c240605887a5532109323A3E12273dc7" as const;
export const USDT0_ADDRESS = "0xC1A5B41512496B80903D1f32d6dEa3a73212E71F" as const;

// Browser traffic goes through our own server route, never straight at the
// tunnel — the extension's proxy sends no CORS headers and ngrok's free tier
// serves an interstitial to browser User-Agents, so a direct call fails in a
// real browser even from localhost. See app/api/ext/[...path]/route.ts.
//
// Node scripts (scripts/e2e-demo.mts) have neither problem and set
// NEXT_PUBLIC_EXT_PROXY_URL to the tunnel directly, skipping the hop.
export const EXT_PROXY_URL = process.env.NEXT_PUBLIC_EXT_PROXY_URL ?? "/api/ext";

export const ERC20_ABI = [
  {
    type: "function",
    name: "approve",
    stateMutability: "nonpayable",
    inputs: [
      { name: "spender", type: "address" },
      { name: "amount", type: "uint256" },
    ],
    outputs: [{ type: "bool" }],
  },
  {
    type: "function",
    name: "allowance",
    stateMutability: "view",
    inputs: [
      { name: "owner", type: "address" },
      { name: "spender", type: "address" },
    ],
    outputs: [{ type: "uint256" }],
  },
  {
    type: "function",
    name: "balanceOf",
    stateMutability: "view",
    inputs: [{ name: "account", type: "address" }],
    outputs: [{ type: "uint256" }],
  },
  {
    type: "function",
    name: "decimals",
    stateMutability: "view",
    inputs: [],
    outputs: [{ type: "uint8" }],
  },
] as const;

export const RFQ_SETTLEMENT_ABI = [
  {
    type: "event",
    name: "Filled",
    inputs: [
      { name: "rfqId", type: "bytes32", indexed: true },
      { name: "taker", type: "address", indexed: true },
      { name: "maker", type: "address", indexed: true },
      { name: "side", type: "uint8", indexed: false },
      { name: "size", type: "uint256", indexed: false },
      { name: "price", type: "uint256", indexed: false },
    ],
  },
  {
    type: "function",
    name: "settled",
    stateMutability: "view",
    inputs: [{ name: "", type: "bytes32" }],
    outputs: [{ type: "bool" }],
  },
] as const;
