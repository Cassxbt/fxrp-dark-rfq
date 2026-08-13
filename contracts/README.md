# RfqSettlement contracts

Foundry project for the on-chain half of the FXRP Dark RFQ. Full design and trust
model: see [`../BUILD-SPEC.md`](../BUILD-SPEC.md) §2.1.

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

8 tests, all passing: the two decimal-math tests required by the spec (buy/sell,
1 FXRP at 2.00 mUSD, verifying the corrected formula lands on exactly 2e18 and not
the round-1 off-by-10¹² result), replay/expiry/untrusted-signer guards, and three
FTSO-bound tests against a mocked feed (within tolerance, outside tolerance, stale).

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
