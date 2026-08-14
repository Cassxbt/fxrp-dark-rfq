# FXRP Dark RFQ — frontend

Two screens: taker (open + close an RFQ) and maker (submit a quote). Trust
model and known limitations: see [`../README.md`](../README.md).

## Setup

```bash
npm install
npm run dev
```

Set `NEXT_PUBLIC_EXT_PROXY_URL` if the ngrok tunnel URL changes from the one hardcoded as a default in `lib/contracts.ts`.

## The one piece worth reading before touching this code

`lib/ecies.ts` is a from-scratch, browser-safe port of go-ethereum's `crypto/ecies` package — not a generic ECIES library. The TEE extension decrypts with that exact Go implementation (NIST concat-KDF, AES-128-CTR, HMAC-SHA256 tag, specific byte layout), and a "standard" ECIES scheme will not produce compatible ciphertext. This was verified byte-for-byte against the live deployed TEE — twice, once with Node's native `crypto` module and once with the actual browser-safe Web Crypto + `@noble/curves` implementation that ships in this app — before any UI was built on top of it. Re-run `node scripts/verify-ecies.mjs` (with the stack's ngrok URL set inside the script) if this file is ever touched.

## What's verified vs. not

- **Verified**: production build (`next build`) passes cleanly; all routes render with zero console errors; the ECIES encryption is proven byte-compatible with the real TEE; the full approve → sign → submit → settle flow has been run end to end against the live Coston2 contract and the live TEE extension, producing a real `Filled` event (`../scripts/e2e-demo.mts` — see the root README's "Proof of a real fill"). That script calls the same `lib/eip712.ts`, `lib/rfqClient.ts`, and `lib/quoteAmount.ts` this UI calls.
- **Not yet verified**: a literal browser click-through with a wallet extension (MetaMask popups, this UI's screens end to end). The e2e script proves the underlying code path works; it doesn't prove the React components wire it up correctly, since it bypasses the UI entirely.

## Known MVP limitations (disclosed, not bugs)

- No public RFQ listing — a maker learns an open RFQ's ID, side, and size directly from the taker, not from the app.
- `CLOSE` reports a match synchronously but settlement is submitted asynchronously (the FCC framework has a hardcoded 2-second response timeout that a chain round-trip can't fit in). Confirm settlement via the `RfqSettlement` contract's `Filled` event, not the close response — the taker page does this via `useWatchContractEvent`.
