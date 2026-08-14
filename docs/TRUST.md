# Trust model

What `RfqSettlement.settle()` actually checks, and what it deliberately doesn't.

**Checked on-chain:** only the TEE extension's own attestation signature over
`Fill` (`rfqId`, `taker`, `maker`, `side`, `size`, `price`, `expiry`), recovered
and checked against an owner-controlled allowlist (`isAttestedSigner`).

**Not checked on-chain, by design:** the taker's `RfqIntent` signature and the
winning maker's `Quote` signature. Both are verified once, off-chain, inside
the Go extension (`extension/go/internal/extension/rfq.go`) before a `Fill` is
ever produced. The contract trusts the extension to have done that correctly —
a deliberate scope choice under this deployment's simulated-TEE model, not an
oversight.

**`isAttestedSigner` is an MVP owner allowlist**, not a live read of FCC's
`TeeExtensionRegistry`. That registry's exact ABI (a clean `isSigner(extension,
addr)` call) wasn't confirmed against the deployed scaffold in time for this
submission — the in-repo `ITeeExtensionRegistry` interface only exposes
`sendInstructions` / `getTeeExtensionInstructionsSender`.

**The `/direct` HTTP layer (OPEN/QUOTE/CLOSE) is unauthenticated.** Anyone
reaching the proxy tunnel can call `CLOSE` on any open RFQ ID. `OPEN`'s result
also returns `{rfqId, side, size}` in cleartext over that same layer — not
just via the on-chain `Filled` event. The taker's limit price and every
losing quote never leave the TEE's encrypted boundary; side and size are not
as tightly held as the on-chain summary alone would suggest.

**No FTSO price bound is active on this deployment.** The contract supports
one (`setFtsoBound`, owner-only, 3 passing tests), but `ftso()` reads as the
zero address and the check short-circuits when unset. No fill on this
deployment — including the one linked from the root README — has been
oracle-checked.

Full limitations list, deployed addresses, and the proof-of-fill transaction:
see the [root README](../README.md).
