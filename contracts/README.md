# RfqSettlement contracts

Foundry project for the on-chain half of the FXRP Dark RFQ. Trust model and
known limitations: see [`../README.md`](../README.md).

## Setup

Dependencies (`forge-std`, `openzeppelin-contracts`) are gitignored — install them fresh:

```bash
forge install foundry-rs/forge-std --no-git
forge install OpenZeppelin/openzeppelin-contracts --no-git
```

## Build & test

```bash
forge build
forge test -vv
```

10 tests, all passing: the two decimal-math tests required by the spec (buy/sell,
1 FXRP at 2.00 mUSD, verifying the corrected formula lands on exactly 2e18 and not
the round-1 off-by-10¹² result), a third decimal test against the actual production
6/6 FXRP/USDT0 pair (`test_TakerBuy_SixAndSix_MatchesLiveFillShape`, pinned to the
same 2.95 USDT0 result as the real funded fill — the other two use a 6/18 mock pair
for historical reasons and don't exercise production's token shape),
replay/expiry/untrusted-signer/zero-amount guards, and three FTSO-bound tests
against a mocked feed (within tolerance, outside tolerance, stale).

## Known non-issues, hardened but not redeployed

`settle`'s `quoteAmount` calculation previously computed `fill.size * fill.price` as
a plain checked-arithmetic argument before passing it into `Math.mulDiv` — that
multiplication could overflow-revert on its own, without ever reaching mulDiv's
own overflow-safe path (an internal audit finding). Fixed in source by scaling
`fill.price` by `10**quoteDecimals` first instead, so `mulDiv` protects the actual
`size * price` multiplication. Not a live bug: no realistic `size`/`price` pair
from a real fill can trigger it, and the currently deployed contract instance
predates this hardening — it wasn't redeployed for this fix, since it changes no
behavior for any value that could actually occur. Flagging explicitly so this
isn't mistaken for "fixed on-chain."

## What's deliberately not verified on-chain

`settle` checks only the TEE's own attestation signature over `Fill`. It does not
re-verify the taker's `RfqIntent` or the maker's `Quote` signatures — that
verification happens once, inside the Go extension (see `../extension/`). This is
a disclosed scope choice under the `SIMULATED_TEE=true` trust model, not an
oversight — see BUILD-SPEC.md §2.1's "trust-model disclosure" note.

`isAttestedSigner` is an MVP owner-controlled allowlist, not a live
`TeeExtensionRegistry` check — that registry's exact ABI is unconfirmed against the
deployed FCC scaffold as of this commit. Swapping in the real check is the next
integration step once confirmed.
