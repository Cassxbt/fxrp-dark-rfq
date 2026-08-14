# FXRP Dark RFQ — frontend

Two screens: taker (open + close an RFQ) and maker (submit a quote). Trust
model and known limitations: see [`../README.md`](../README.md).

## Setup

```bash
npm install
npm run dev
```

Set `NEXT_PUBLIC_EXT_PROXY_URL` if the ngrok tunnel URL changes from the one hardcoded as a default in `lib/contracts.ts`.

## Before touching `lib/ecies.ts`

It's a browser-safe port of go-ethereum's `crypto/ecies` package, not a
generic ECIES library — the TEE decrypts with that exact scheme, so a
standard implementation won't produce compatible ciphertext. Re-run
`npx tsx scripts/verify-ecies.mts` against the live tunnel if this file
changes.

## What's verified vs. not

- **Verified**: production build (`next build`) passes cleanly; all routes render with zero console errors; the ECIES encryption is proven byte-compatible with the real TEE; the full approve → sign → submit → settle flow has been run end to end against the live Coston2 contract and the live TEE extension, producing a real `Filled` event (`scripts/e2e-demo.mts` — see the root README's [Demo](../README.md#demo)). That script calls the same `lib/eip712.ts`, `lib/rfqClient.ts`, and `lib/quoteAmount.ts` this UI calls.
- **Not yet verified**: a literal browser click-through with a wallet extension (MetaMask popups, this UI's screens end to end). The e2e script proves the underlying code path works; it doesn't prove the React components wire it up correctly, since it bypasses the UI entirely.

## Known MVP limitations (disclosed, not bugs)

- No public RFQ listing — a maker learns an open RFQ's ID, side, and size directly from the taker, not from the app.
- `CLOSE` reports a match synchronously but settlement is submitted asynchronously (the FCC framework has a hardcoded 2-second response timeout that a chain round-trip can't fit in). Confirm settlement via the chain, not the close response — the taker page polls `settled()` and then reads the `Filled` event back with `getContractEvents`. `useWatchContractEvent` is the fast path on top of that, not the source of truth: it only fires for logs arriving while it happens to be subscribed.
