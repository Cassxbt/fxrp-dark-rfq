// Regression check for two bugs found only by testing against the live TEE,
// not by code review alone:
//  1. wagmi's signTypedData returns hex; Go's `Signature []byte` decodes
//     base64 by default. An unconverted hex string decodes to garbage and
//     fails recoverSigner's `len(sig) != 65` check.
//  2. bigint fields were being sent as quoted JSON strings ("1000000"), but
//     Go's math/big.Int.UnmarshalJSON requires a bare number (1000000) and
//     rejects a string outright.
// Uses the actual sendRfqDirect from lib/rfqClient.ts, not a reimplementation
// — the whole point of this script is to catch wire-format bugs in the real
// code path, and a duplicated implementation could pass while the shipped
// code stays broken (this was itself a prior code-review finding).
import { sendRfqDirect } from "../lib/rfqClient";

async function main() {
  // 65 bytes: r(32) + s(32) + v(1) — not a real signature over anything,
  // just the right shape and length to get past the wire-format checks and
  // prove the encoding is correct end to end.
  const fakeSig = ("0x" + "11".repeat(32) + "22".repeat(32) + "1b") as `0x${string}`;

  const intent = {
    side: 0,
    size: 1_000_000n,
    limitPrice: 3_000_000_000_000_000_000n,
    taker: "0x0000000000000000000000000000000000000001",
    expiry: BigInt(Math.floor(Date.now() / 1000) + 300),
    rfqNonce: BigInt(Date.now()),
  };

  const result = await sendRfqDirect("OPEN", { data: intent, signature: fakeSig });
  console.log("Result:", JSON.stringify(result, null, 2));

  if (result.log.includes("must be 65 bytes")) {
    console.log("\nFAIL: signature still not decoding to 65 bytes on the Go side.");
    process.exit(1);
  }
  if (result.log.includes("cannot unmarshal") && result.log.includes("big.Int")) {
    console.log("\nFAIL: bigint fields still not decoding as raw numbers.");
    process.exit(1);
  }
  if (result.log.includes("parsing RfqIntent envelope")) {
    console.log("\nFAIL: envelope still fails to parse:", result.log);
    process.exit(1);
  }
  console.log(
    "\nPASS: signature decodes to 65 bytes and RfqIntent parses cleanly. " +
      "(Remaining error, if any, is about the fake signature not matching the taker address — expected, not a wire-format bug.)",
  );
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});
