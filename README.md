# FXRP Dark RFQ

Sealed-bid RFQ matcher for FXRP on Coston2. A taker opens an intent (side, size,
limit price); makers submit quotes; matching runs privately inside a **Flare
Confidential Compute (FCC)** TEE extension — only the winning fill is ever
written on-chain. Limit prices, losing quotes, and maker identity of non-winners
never leave the TEE.

Built for [Flare Summer Signal](https://dorahacks.io/hackathon/flaresummersignal/detail)
— Bounty 2 (Confidential Compute) primary, Bounty 1 (Interoperable Assets / FXRP)
as integration proof.

## Proof of a real fill

[`0xe158ffe7...a7a8e9`](https://coston2-explorer.flare.network/tx/0xe158ffe70bd1df2790ca3bc09c501cf214f6c7a7406872882361698551a7a8e9) —
a taker buying 1 FXRP, two makers quoting 2.95 and 2.99 USDT0, the TEE
selecting the cheaper of the two, and both ERC-20 legs settling atomically.
Produced by `frontend/scripts/e2e-demo.mts`, which drives the real
`lib/eip712.ts` / `lib/rfqClient.ts` / `lib/quoteAmount.ts` the UI itself
calls — not a reimplementation.

## What's deployed

- **RfqSettlement contract (Coston2):** [`0xaBf47C48c00DDa806f1d9243c936A8153C7E6FcE`](https://coston2-explorer.flare.network/address/0xaBf47C48c00DDa806f1d9243c936A8153C7E6FcE)
- **FXRP:** `0x0b6A3645c240605887a5532109323A3E12273dc7`
- **USDT0:** `0xC1A5B41512496B80903D1f32d6dEa3a73212E71F`
- **FCC extension:** Go service running against Coston2 in **simulated TEE mode**
  (`SIMULATED_TEE=true` — sanctioned by the Flare team for judging; no real
  hardware attestation backs the signer key in this deployment)

## How it works

1. **Taker** opens an RFQ intent (buy/sell, size, limit price), EIP-712-signed,
   sent ECIES-encrypted straight to the TEE extension's `/direct` endpoint. The
   limit price never touches the chain or a public order book — there is no
   listing to browse; the taker shares the RFQ ID out of band with makers.
2. **Makers** submit non-binding quotes against that RFQ ID, also sealed to the
   TEE. Quotes are matching signals, not on-chain commitments — nothing is
   escrowed at quote time.
3. Taker (or anyone — see Known limitations) **closes** the RFQ. The extension
   picks the winning quote inside the TEE, checked against the taker's limit,
   and submits a signed `Fill` to the contract. The contract also supports an
   optional FTSO price-bound check (owner-configurable, covered by 3 passing
   tests) — **it is not enabled on the current deployment** (`ftso()` reads as
   unset), so no fill has been oracle-checked yet. Don't read this deployment
   as having a live price guard.
4. The contract verifies the TEE's attestation signature and settles the
   ERC-20 transfer atomically. A `Filled` event is the only public trace of
   the trade — no losing quotes, no rejected intents, no maker identities
   beyond the winner.

## Run the demo

Needs three funded Coston2 wallets (taker + 2 makers — matching requires at
least one competing quote to prove selection is happening, not just pass-through).

```bash
cd frontend
npm install
npm run dev
```

- `/taker` — connect wallet, open an RFQ, share the printed RFQ ID with makers,
  close once quotes are in, watch for the `Filled` event.
- `/maker` — connect wallet, paste the RFQ ID, submit a quote.

The extension and proxy are already running against Coston2; `EXT_PROXY_URL` in
`frontend/lib/contracts.ts` points at the current tunnel.

## Trust model and known limitations

- **No FTSO price bound is active on this deployment.** The contract supports
  one (owner-configurable, tested), but it was never turned on — `ftso()`
  reads as the zero address, and `_checkFtsoBound` short-circuits when unset.
  Every fill, including the one linked below, settled on the taker's limit
  and the maker's quote only, with no oracle check.
- **Simulated TEE, not hardware-attested.** This deployment runs FCC's
  officially-sanctioned simulated mode. The signer key's trustworthiness rests
  on the extension process, not a hardware attestation chain.
- **`isAttestedSigner` is an owner-controlled allowlist**, not a live read of
  FCC's `TeeExtensionRegistry`. That registry's exact ABI wasn't confirmed
  against the deployed scaffold in time for this submission — see
  `contracts/README.md`.
- **`CLOSE` is unauthenticated.** Any address can trigger matching for any open
  RFQ ID, not just the taker who opened it. Since IDs aren't listed publicly
  and closing early just picks among whatever quotes exist so far, this is a
  disclosed scope cut, not an oversight.
- **On-chain settlement is asynchronous from the `CLOSE` call** (FCC's
  response window doesn't fit a chain round trip). The taker UI polls the
  contract's `settled()` for up to ~30s and reports success/failure
  explicitly rather than leaving the call looking like it hung.
- **The contract verifies only the TEE's `Fill` signature**, not the
  underlying `RfqIntent`/`Quote` signatures a second time — those are checked
  once inside the extension. See `contracts/README.md`.

## Repo layout

- `contracts/` — `RfqSettlement.sol` + Foundry tests (9 passing)
- `extension/` — Go FCC extension: RFQ intake, matching, TEE-signed settlement submission
- `frontend/` — Next.js taker/maker UI
