# FXRP Dark RFQ

Sealed-bid RFQ matcher for FXRP that matches privately inside a Flare Confidential Compute TEE, settling atomically on Coston2.

[![Solidity](https://img.shields.io/badge/Solidity-0.8.24-363636?style=flat-square&logo=solidity&logoColor=white)](contracts/src/RfqSettlement.sol)
[![Next.js](https://img.shields.io/badge/Next.js-16-000000?style=flat-square&logo=next.js&logoColor=white)](frontend/package.json)
[![Network](https://img.shields.io/badge/Network-Coston2-e84142?style=flat-square)](https://coston2-explorer.flare.network)
[![contracts](https://img.shields.io/github/actions/workflow/status/Cassxbt/fxrp-dark-rfq/test.yml?style=flat-square&label=forge%20test)](https://github.com/Cassxbt/fxrp-dark-rfq/actions/workflows/test.yml)

Built for [Flare Summer Signal](https://dorahacks.io/hackathon/flaresummersignal/detail) — Bounty 2 (Confidential Compute) primary, Bounty 1 (Interoperable Assets / FXRP) as integration proof.

## Demo

**[`0xe158ffe7...a7a8e9`](https://coston2-explorer.flare.network/tx/0xe158ffe70bd1df2790ca3bc09c501cf214f6c7a7406872882361698551a7a8e9)** —
a taker buying 1 FXRP, two makers quoting 2.95 and 2.99 USDT0, the TEE
selecting the cheaper of the two, both ERC-20 legs settling atomically in one
transaction. Real, independently checkable on the explorer — not a mock.

Produced by [`frontend/scripts/e2e-demo.mts`](frontend/scripts/e2e-demo.mts), which drives the same `lib/eip712.ts`, `lib/rfqClient.ts`, and `lib/quoteAmount.ts` the UI itself calls — not a reimplementation.

## Overview

An RFQ desk leaks information by design: a public limit order book shows your size and price to everyone before you trade. FXRP Dark RFQ keeps the taker's limit price and every losing quote inside a Flare Confidential Compute (FCC) TEE extension — the chain only ever learns the winning fill, after the fact. Matching logic (best-price selection against a taker's sealed limit) runs off-chain inside the TEE; settlement is a single atomic on-chain transaction that a Solidity contract executes once it verifies the TEE's own attestation signature over the fill.

## Features

- **Sealed-bid matching** — taker limit price and losing maker quotes never touch the chain or a public order book; only the winning `Filled` event is public
- **Atomic on-chain settlement** — both ERC-20 legs (FXRP and USDT0) transfer in one transaction once the TEE-signed `Fill` is verified
- **Real assets** — live FXRP and USDT0 on Coston2, not toy ERC-20s
- **Real funded proof** — `scripts/e2e-demo.mts` runs the full OPEN → QUOTE ×2 → CLOSE → `Filled` flow against live Coston2 state through the actual production code paths, not a simulation

## Architecture

```mermaid
graph TD
    A[Taker wallet] -- "EIP-712 RfqIntent, ECIES-encrypted" --> B["POST /direct (ngrok tunnel)"]
    C[Maker wallet] -- "EIP-712 Quote, ECIES-encrypted" --> B
    B --> D[tee-proxy]
    D --> E[tee-node]
    E --> F["extension /action (Go, simulated TEE)"]
    F -- "OPEN: decrypt, recover taker" --> G[in-memory RFQ book]
    F -- "QUOTE: decrypt, recover maker" --> G
    F -- "CLOSE: selectWinner, sign Fill" --> H["hot key: settle(Fill, sig)"]
    H --> I["RfqSettlement.sol (Coston2)"]
    I -- "recover allowlisted signer" --> I
    I -- "safeTransferFrom both legs" --> J[(FXRP / USDT0)]
    I -- "emit Filled" --> K[On-chain: winner, size, price]
```

The taker's limit price and every losing quote never leave the TEE's encrypted
boundary — only the winning `Filled` event reaches the chain. That's a
narrower claim than "nothing else is ever visible": the `/direct` HTTP layer
itself is unauthenticated, and `OPEN`'s result exposes `{rfqId, side, size}`
in cleartext to anyone polling it — see Known limitations and
[`docs/TRUST.md`](docs/TRUST.md).

## Tech stack

| Component | Technology |
|---|---|
| Settlement contract | Solidity 0.8.24, Foundry, OpenZeppelin |
| Confidential compute | Flare Confidential Compute (FCC), Go extension, simulated TEE mode |
| Frontend | Next.js 16 · React 19 · wagmi · viem |
| Crypto | EIP-712 typed signing · ECIES (browser-safe port of go-ethereum's scheme) |
| Network | Flare Coston2 testnet |
| Assets | FXRP · USDT0 |

## Quickstart

```bash
git clone https://github.com/Cassxbt/fxrp-dark-rfq
cd fxrp-dark-rfq/frontend
npm install
npm run dev
```

This starts the UI only — `/taker` and `/maker` render, but neither can
actually open or quote an RFQ without the FCC extension running behind them,
which for this submission means the live ngrok tunnel described in **Judge
testing notes** below staying up, plus three funded Coston2 wallets. If you
just want to see it work without any of that, the proof-of-fill transaction
above and `scripts/e2e-demo.mts`'s output are the evidence — see
[Usage](#usage) for both paths.

## Configuration

| Variable | Description | Required |
|---|---|---|
| `NEXT_PUBLIC_EXT_PROXY_URL` | URL of the FCC extension's proxy (`/info`, `/direct`) | No — falls back to the current default in `frontend/lib/contracts.ts` |

The extension and its ngrok tunnel are already running against Coston2 for judging; see **Judge testing notes** below before relying on the live tunnel.

## Usage

1. **Taker** (`/taker`): connect a Coston2 wallet, pick buy/sell + size + limit price, click **Open RFQ** (approve, then sign). A shareable blob appears — `RFQ 0x... — taker is BUYING/SELLING X FXRP` — since there's no public listing.
2. **Makers** (`/maker`, 2 different wallets): paste the RFQ ID, match the stated side/size, enter a price, click **Submit Quote** (approve, then sign). Use two different prices — this is what proves selection is happening, not pass-through.
3. **Taker**: click **Close RFQ**. The UI polls the contract's `settled()` for up to ~30s and reports success or an explicit failure reason. A confirmed fill shows the winning maker, size, and price, pulled from the on-chain `Filled` event.

### Reproducing the funded proof (maintainers)

```bash
cd frontend
cp .env.e2e.local.example .env.e2e.local   # fill in three funded burner keys
npx tsx --env-file=.env.e2e.local scripts/e2e-demo.mts
```

Bypasses the browser entirely, calling the same `lib/` code directly. This is
what produced the transaction linked under Demo above.

## Deployed contracts

| Contract | Network | Address |
|---|---|---|
| `RfqSettlement.sol` | Coston2 | [`0xaBf47C48c00DDa806f1d9243c936A8153C7E6FcE`](https://coston2-explorer.flare.network/address/0xaBf47C48c00DDa806f1d9243c936A8153C7E6FcE) |
| FXRP | Coston2 | [`0x0b6A3645c240605887a5532109323A3E12273dc7`](https://coston2-explorer.flare.network/address/0x0b6A3645c240605887a5532109323A3E12273dc7) |
| USDT0 | Coston2 | [`0xC1A5B41512496B80903D1f32d6dEa3a73212E71F`](https://coston2-explorer.flare.network/address/0xC1A5B41512496B80903D1f32d6dEa3a73212E71F) |

Proof-of-fill transaction: [`0xe158ffe7...a7a8e9`](https://coston2-explorer.flare.network/tx/0xe158ffe70bd1df2790ca3bc09c501cf214f6c7a7406872882361698551a7a8e9).

## Known limitations

- **No FTSO price bound is active on this deployment.** The contract supports one (owner-configurable, covered by 3 passing tests), but it was never turned on — `ftso()` reads as the zero address, and the bound check short-circuits when unset. Every fill, including the one linked above, settled on the taker's limit and the maker's quote only, with no oracle check.
- **Simulated TEE, not hardware-attested.** This deployment runs FCC's officially-sanctioned simulated mode. The signer key's trustworthiness rests on the extension process, not a hardware attestation chain.
- **`isAttestedSigner` is an owner-controlled allowlist**, not a live read of FCC's `TeeExtensionRegistry`. That registry's exact ABI wasn't confirmed against the deployed scaffold in time for this submission.
- **`CLOSE` is unauthenticated**, and the `/direct` HTTP layer that carries OPEN/QUOTE/CLOSE has no authentication at all. Any address can trigger matching for any open RFQ ID. `OPEN`'s result also returns `{rfqId, side, size}` in cleartext to anyone polling it — side and size are not as tightly held as the on-chain summary alone would suggest, though the limit price and losing quotes never leave the TEE. Disclosed scope cuts, not oversights.
- **On-chain settlement is asynchronous from the `CLOSE` call** (FCC's response window doesn't fit a chain round trip). The taker UI polls `settled()` and reports success/failure explicitly rather than leaving the call looking like it hung.
- **The contract verifies only the TEE's `Fill` signature**, not the underlying `RfqIntent`/`Quote` signatures a second time — those are checked once inside the extension.

### Judge testing notes

The FCC extension is reachable through an ngrok tunnel whose URL is hardcoded as a fallback default in `frontend/lib/contracts.ts` — it goes dead if the host machine sleeps or the tunnel restarts. If `/taker` or `/maker` can't reach `/info`, the live tunnel is down; treat the proof-of-fill transaction above and `scripts/e2e-demo.mts`'s output as the evidence of record regardless of live tunnel uptime.

## Repo layout

- `contracts/` — `RfqSettlement.sol` + Foundry tests (10 passing)
- `extension/` — Go FCC extension: RFQ intake, matching, TEE-signed settlement submission
- `frontend/` — Next.js taker/maker UI + `scripts/e2e-demo.mts`

## License

MIT — see [`LICENSE`](LICENSE).
