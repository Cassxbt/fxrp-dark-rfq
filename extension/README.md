# FCC extension — FXRP Dark RFQ matcher

This directory is the Flare Confidential Compute (FCC) extension that matches
sealed RFQs for [FXRP Dark RFQ](../README.md). The matcher itself is
[`go/internal/extension/rfq.go`](go/internal/extension/rfq.go) — RFQ intake,
sealed-bid winner selection, and TEE-signed settlement submission. Runs in
**simulated TEE mode**; `/direct` accepts `RFQ` op types only (`KEY`
management was locked to the on-chain `InstructionSender` path after a
security review — see [`../docs/TRUST.md`](../docs/TRUST.md)).

This directory started from Flare's `fce-sign` scaffold (a private-key-manager
example unrelated to this submission's RFQ logic). What follows below is that
scaffold's own deploy, registration, and troubleshooting tooling, which this
project's build actually uses to get the Go extension onto Coston2 — kept
because it's a real operational runbook, not because the example it
originally documented is part of this submission.

> **Scaffold's own warning**, preserved as-is: storing encrypted secrets
> on-chain is not advisable in production — on-chain data is public and
> encryption can be broken over time. A production extension should use
> off-chain channels for secret delivery. (Not applicable to the RFQ matcher
> above, which stores no secrets on-chain.)

## Layout & deployable surface

This deployment always builds [`go/`](go/) via [`Dockerfile`](Dockerfile) —
bit-for-bit reproducible across machines. `LANGUAGE=go` in `.env.<chain>`
selects it; the scaffold originally also shipped Python/TypeScript variants
of the example extension, removed from this repo since they were never used
here.

```bash
# .env.coston2
LANGUAGE=go
```

`scripts/start-services.sh` maps `LANGUAGE` to the right Dockerfile via
`EXTENSION_DOCKERFILE`, which `docker-compose.yaml` then uses for the
`extension-tee` build. The on-chain registration tooling under `go/tools/`
runs on the developer machine, never inside the TEE.

See [`REPRODUCIBILITY.md`](REPRODUCIBILITY.md) for `SOURCE_DATE_EPOCH` and
reproducible-build details.

### Running the Go tests

```bash
cd go && go test ./...
```

## Shared contract

`contracts/InstructionSender.sol` is shared across all implementations. It
declares `OP_TYPE_KEY = bytes32("KEY")`, `OP_COMMAND_UPDATE = bytes32("UPDATE")`
and `OP_COMMAND_SIGN = bytes32("SIGN")` and exposes `updateKey(bytes)` and
`sign(bytes)` entry points that route through the Flare TEE Manager diamond.

## Deploying and Testing

The full testnet flow (Coston/Coston2 with a devops-hosted Confidential Space
VM) is documented in [`TESTNET_DEPLOYMENT.md`](TESTNET_DEPLOYMENT.md). The
short version:

```bash
bash ./scripts/use-chain.sh coston2       # or coston
bash ./scripts/full-setup.sh --chain coston2 --test
```

For local development against a Hardhat devnet + locally-built Docker stack:

```bash
bash ./scripts/use-chain.sh local
bash ./scripts/full-setup.sh --test       # defaults to --chain local
```

Each phase can also be run individually:

```bash
./scripts/pre-build.sh         # 1. Deploy contract + register extension → config/extension.env
./scripts/start-services.sh    # 2. Docker compose up (redis + ext-proxy + extension-tee)
./scripts/post-build.sh        # 3. Allow TEE version + register TEE machine on-chain
./scripts/test.sh              # 4. End-to-end UPDATE/SIGN test against the running TEE
./scripts/stop-services.sh     # Tear down
```

To build a hand-off image for a devops-hosted TEE (instead of the local stack),
use `./scripts/build-image.sh` — it builds the `LANGUAGE` from `.env`, verifies
`MODE=0`, and saves a tar. See [`DEPLOYMENT_STEPS.md`](DEPLOYMENT_STEPS.md).

### Prerequisites

- **Docker Desktop** (Linux containers) — for the local stack
- **Go 1.25.1+** — for the deploy + registration tools in `go/tools/`
- **Foundry** (`forge`, `cast`, `jq`) — for Solidity compilation and contract bindings
- **Bash** — Git Bash works on Windows
- **No sibling repos needed** — `tee-node` and `tee-proxy` are fetched from the
  public `github.com/flare-foundation` repos at build time (Go modules pinned in
  `go.sum`; the proxy image `git clone`s them). A flat checkout of just this repo
  is enough.
- **A funded private key** for the target chain. Set as `DEPLOYMENT_PRIVATE_KEY`
  in `.env.<chain>` (no `0x` prefix). Fund at
  [`faucet.flare.network`](https://faucet.flare.network/).
- For Coston/Coston2 deploys: a devops contact who'll run the TEE on a real
  GCP Confidential Space VM. See `TESTNET_DEPLOYMENT.md` for the full handoff.

### Chain selection

`.env` is a per-chain file. `scripts/use-chain.sh <chain> [language]` copies the
active chain's template (`.env.coston` or `.env.coston2`) over `.env`, optionally
setting `LANGUAGE` (only `go` is used by this submission). Use `--list` to see
available chains, or `--help` for usage. All scripts then source `.env`
automatically.

| Chain     | `.env.<chain>`  | Addresses file                          | RPC                                              |
| --------- | --------------- | --------------------------------------- | ------------------------------------------------ |
| local     | `.env.example`  | `e2e/docker/sim_dump/deployed-addresses.json` (auto-detected) | `http://127.0.0.1:8545`                          |
| coston    | `.env.coston`   | `config/coston/deployed-addresses.json` | `https://coston-api.flare.network/ext/C/rpc`     |
| coston2   | `.env.coston2`  | `config/coston2/deployed-addresses.json`| `https://coston2-api.flare.network/ext/C/rpc`    |

### Generated artifacts

`pre-build.sh` writes the new `EXTENSION_ID` and `INSTRUCTION_SENDER` to
`config/extension.env`. Every subsequent script (`start-services`, `post-build`,
`test`) reads this file automatically — no manual `.env` edits required.

## Reproducible builds

The Go `Dockerfile` is bit-for-bit reproducible: same source + same
`SOURCE_DATE_EPOCH` yields an identical image on any host. See
[`REPRODUCIBILITY.md`](REPRODUCIBILITY.md).

## Troubleshooting

See `TESTNET_DEPLOYMENT.md` § Troubleshooting for the full catalogue. Common
issues:

- **`connect: connection refused` from ext-proxy** — your indexer DB is
  unreachable. Check your indexer host/creds and network connectivity.
- **`Verification.TeeNotFound`** — `NORMAL_PROXY_URL` is pointed at the wrong
  chain's FTDC proxy.
- **`Verification.ChallengeExpired`** — re-run `post-build.sh`; the patched
  `register-tee` already passes `-command rRap` for fresh attestation.
- **`code hashes do not match`** — `SIMULATED_TEE` and the TEE's `MODE` env
  disagree. Both must point at "real" for testnet (`SIMULATED_TEE=false`,
  `MODE=0`).
- **On-chain instructions never reach the extension** — the extension's
  current TEE identity may not be registered on-chain (a fresh container
  build produces a new simulated TEE ID). Diagnose with
  `go run ./cmd/check-tee-state -a <addresses-file> -c <chain-url> -p
  <proxy-url> -instructionSender <address>`; fix by re-running
  `post-build.sh` against the running extension.

## Related docs

| Doc                                                  | What it covers                                     |
| ----------------------------------------------------- | --------------------------------------------------- |
| [`TESTNET_DEPLOYMENT.md`](TESTNET_DEPLOYMENT.md)       | End-to-end Coston/Coston2 deploy with devops handoff |
| [`REPRODUCIBILITY.md`](REPRODUCIBILITY.md)             | `SOURCE_DATE_EPOCH` and reproducible image builds    |
| [`go/`](go/)                                          | Go extension binary, RFQ matcher, deploy/registration tooling |
