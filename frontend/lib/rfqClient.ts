import { eciesEncrypt, bytesToHex, hexToBytes } from "./ecies";
import { EXT_PROXY_URL } from "./contracts";

function opHash(s: string): `0x${string}` {
  const bytes = new TextEncoder().encode(s);
  if (bytes.length > 32) throw new Error(`op string too long: ${s}`);
  const padded = new Uint8Array(32);
  padded.set(bytes);
  return bytesToHex(padded) as `0x${string}`;
}

let cachedPubKey: Uint8Array | null = null;

/** Fetches and caches the TEE's ECIES encryption public key from /info. */
export async function fetchTeePubKey(): Promise<Uint8Array> {
  if (cachedPubKey) return cachedPubKey;
  const resp = await fetch(`${EXT_PROXY_URL}/info`);
  if (!resp.ok) throw new Error(`GET /info failed: ${resp.status}`);
  const info = await resp.json();
  const { x, y } = info.machineData.publicKey as { x: string; y: string };
  const pubKey = new Uint8Array(65);
  pubKey[0] = 0x04;
  pubKey.set(hexToBytes(x), 1);
  pubKey.set(hexToBytes(y), 33);
  cachedPubKey = pubKey;
  return pubKey;
}

interface DirectResult {
  status: number;
  log: string;
  data: string; // hex
}

/**
 * Encrypts `data` to the TEE's pubkey, wraps it with `signature` in the
 * signedEnvelope shape the extension expects (see rfq.go's signedEnvelope[T]),
 * and dispatches it via the proxy's /direct endpoint. Polls for the result —
 * matches the pattern verified live against the real stack (see
 * extension/go/internal/extension/rfq_integration_test.go and
 * frontend/scripts/verify-ecies.mjs).
 */
export async function sendRfqDirect(
  opCommand: "OPEN" | "QUOTE" | "CLOSE",
  payload: { data: unknown; signature: `0x${string}` } | Uint8Array,
): Promise<DirectResult> {
  const pubKey = await fetchTeePubKey();

  let plaintext: Uint8Array;
  if (payload instanceof Uint8Array) {
    // CLOSE sends a plaintext rfqId, not an encrypted envelope — there's no
    // secret in "please close now," matching rfq.go's processRfqClose which
    // reads df.OriginalMessage directly as 32 raw bytes.
    plaintext = payload;
  } else {
    const json = JSON.stringify(payload, (_key, value) =>
      typeof value === "bigint" ? value.toString() : value,
    );
    plaintext = new TextEncoder().encode(json);
  }

  const message =
    opCommand === "CLOSE" ? bytesToHex(plaintext) : bytesToHex(await eciesEncrypt(pubKey, plaintext));

  const directResp = await fetch(`${EXT_PROXY_URL}/direct`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ opType: opHash("RFQ"), opCommand: opHash(opCommand), message }),
  });
  if (!directResp.ok) {
    throw new Error(`POST /direct failed: ${directResp.status} ${await directResp.text()}`);
  }
  const directJson = await directResp.json();
  const actionId: string = directJson.data.id;

  // Direct actions are tagged "submit" (queue.DirectInstructionToAction on
  // the proxy side) — /action/result/{id} defaults to "threshold" if this
  // isn't passed explicitly, which silently returns nothing for a
  // submit-tagged action (found and fixed the hard way in the Go integration
  // test — see BUILD-SPEC.md's changelog).
  for (let i = 0; i < 15; i++) {
    const resp = await fetch(`${EXT_PROXY_URL}/action/result/${actionId}?submissionTag=submit`);
    if (resp.ok) {
      const json = await resp.json();
      if (json.result?.status !== 0 || json.result?.log) {
        return { status: json.result.status, log: json.result.log ?? "", data: json.result.data ?? "0x" };
      }
    }
    await new Promise((r) => setTimeout(r, 2000));
  }
  throw new Error(`timed out polling result for action ${actionId}`);
}
