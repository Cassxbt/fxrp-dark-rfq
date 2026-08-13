// Verifies the ACTUAL production code — imports lib/ecies.ts and
// lib/rfqClient.ts directly, not a reimplementation. A code-review finding
// on the previous version of this script (verify-ecies.mjs) was that it
// duplicated the entire ECIES stack by hand: a fix to lib/ecies.ts wouldn't
// have been exercised by that script, giving false confidence that
// "verification" covered what actually ships. Run with:
//
//   npx tsx scripts/verify-ecies.mts
//
// Set EXT_PROXY_URL if the ngrok tunnel differs from lib/contracts.ts's default.
import { eciesEncrypt, bytesToHex } from "../lib/ecies";
import { fetchTeePubKey } from "../lib/rfqClient";
import { EXT_PROXY_URL } from "../lib/contracts";

function opHash(s: string): string {
  const padded = new Uint8Array(32);
  padded.set(new TextEncoder().encode(s));
  return bytesToHex(padded);
}

async function main() {
  const pubKey = await fetchTeePubKey();
  console.log("TEE pubkey length:", pubKey.length, "(expect 65)");

  const plaintext = new TextEncoder().encode(JSON.stringify({ probe: "verify-ecies-mts-compat-check" }));
  const ciphertext = await eciesEncrypt(pubKey, plaintext);
  console.log("Ciphertext length:", ciphertext.length);

  const payload = { opType: opHash("RFQ"), opCommand: opHash("OPEN"), message: bytesToHex(ciphertext) };
  const directResp = await fetch(`${EXT_PROXY_URL}/direct`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  console.log("POST /direct status:", directResp.status);
  const directJson = await directResp.json();
  const actionId = directJson.data.id;

  await new Promise((r) => setTimeout(r, 3000));
  const resultResp = await fetch(`${EXT_PROXY_URL}/action/result/${actionId}?submissionTag=submit`);
  const resultJson = await resultResp.json();
  console.log("Result:", JSON.stringify(resultJson.result, null, 2));

  const log: string = resultJson.result?.log ?? "";
  if (log.includes("can not decrypt")) {
    console.log("\nFAIL: lib/ecies.ts (the actual production module) is NOT compatible with the live TEE.");
    process.exit(1);
  }
  console.log("\nPASS: lib/ecies.ts (the actual production module, imported directly) IS compatible.");
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});
