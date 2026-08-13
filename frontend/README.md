# FXRP Dark RFQ — frontend

Two screens per [`../BUILD-SPEC.md`](../BUILD-SPEC.md) §2.3: taker (open + close an RFQ) and maker (submit a quote).

## Setup

```bash
npm install
npm run dev
```

Set `NEXT_PUBLIC_EXT_PROXY_URL` if the ngrok tunnel URL changes from the one hardcoded as a default in `lib/contracts.ts`.

## The one piece worth reading before touching this code

`lib/ecies.ts` is a from-scratch, browser-safe port of go-ethereum's `crypto/ecies` package — not a generic ECIES library. The TEE extension decrypts with that exact Go implementation (NIST concat-KDF, AES-128-CTR, HMAC-SHA256 tag, specific byte layout), and a "standard" ECIES scheme will not produce compatible ciphertext. This was verified byte-for-byte against the live deployed TEE — twice, once with Node's native `crypto` module and once with the actual browser-safe Web Crypto + `@noble/curves` implementation that ships in this app — before any UI was built on top of it. Re-run `node scripts/verify-ecies.mjs` (with the stack's ngrok URL set inside the script) if this file is ever touched.

## What's verified vs. not

- **Verified**: production build (`next build`) passes cleanly; all routes render with zero console errors; the ECIES encryption is proven byte-compatible with the real TEE.
- **Not yet verified**: the full click-through flow (approve → sign → submit → settle) — that requires a real browser wallet extension with a funded Coston2 account, which isn't available in this dev environment's automated testing. That's the funded demo run, a separate remaining step.

## Known MVP limitations (disclosed, not bugs)

- No public RFQ listing — a maker learns an open RFQ's ID, side, and size directly from the taker, not from the app. See BUILD-SPEC.md §5.
- `CLOSE` reports a match synchronously but settlement is submitted asynchronously (the FCC framework has a hardcoded 2-second response timeout that a chain round-trip can't fit in). Confirm settlement via the `RfqSettlement` contract's `Filled` event, not the close response — the taker page does this via `useWatchContractEvent`.
