# FCC extension — FXRP Dark RFQ matcher

This directory is the Flare Confidential Compute (FCC) extension that matches
sealed RFQs for [FXRP Dark RFQ](../README.md). The matcher itself is
[`go/internal/extension/rfq.go`](go/internal/extension/rfq.go) — RFQ intake,
sealed-bid winner selection, and TEE-signed settlement submission. Runs in
**simulated TEE mode**; `/direct` accepts `RFQ` op types only (`KEY`
management was locked to the on-chain `InstructionSender` path after a
security review — see [`../docs/TRUST.md`](../docs/TRUST.md)).

```bash
cd go && go test ./...
```

This directory started from Flare's `fce-sign` scaffold (a private-key-manager
example unrelated to this submission's RFQ logic). That scaffold's own
deploy, registration, and troubleshooting tooling — which this build actually
uses to get the Go extension onto Coston2 — is kept separate in
[`ops.md`](ops.md), so this page stays about the product rather than mixing
in an unrelated example's runbook.
