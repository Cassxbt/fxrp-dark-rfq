import { eciesEncrypt, bytesToHex, hexToBytes, padTo32 } from "./ecies";
import { EXT_PROXY_URL } from "./contracts";

function opHash(s: string): `0x${string}` {
  const bytes = new TextEncoder().encode(s);
  if (bytes.length > 32) throw new Error(`op string too long: ${s}`);
  const padded = new Uint8Array(32);
  padded.set(bytes);
  return bytesToHex(padded) as `0x${string}`;
}

/**
 * Go's encoding/json decodes a plain `[]byte` field as base64 by default,
 * not hex — the extension's signedEnvelope[T]{ Signature []byte } (rfq.go)
 * uses exactly that. wagmi's signTypedData returns a hex string; sending it
 * through unconverted would make Go's json.Unmarshal base64-decode "0x1a2b..."
 * into 99 garbage bytes instead of the real 65-byte signature, and
 * recoverSigner's explicit len(sig) != 65 check would reject every OPEN and
 * QUOTE (code-review finding — this broke the entire happy path, and the
 * project's own Go-to-Go integration test never exposed it because both
 * sides there used Go's default []byte marshaling consistently).
 */
function hexSignatureToBase64(hexSig: `0x${string}`): string {
  const bytes = hexToBytes(hexSig);
  let binary = "";
  for (const b of bytes) binary += String.fromCharCode(b);
  return btoa(binary);
}

/**
 * JSON.stringify with bigint fields emitted as raw, unquoted number literals
 * instead of strings. Go's math/big.Int.UnmarshalJSON requires a bare JSON
 * number ("size":1000000) and rejects a quoted string ("size":"1000000")
 * outright — found live via verify-signature-encoding.mts, not by
 * inspection: "math/big: cannot unmarshal \"1000000\" into a *big.Int".
 * Values like WAD prices (1e18 scale) exceed Number.MAX_SAFE_INTEGER, so a
 * plain `Number(x)` round-trip would lose precision — this uses a sentinel
 * string during JSON.stringify, then strips the surrounding quotes with a
 * regex pass afterward, preserving full precision.
 */
function stringifyWithRawBigInts(value: unknown): string {
  const marker = "__BIGINT__";
  const json = JSON.stringify(value, (_key, v) => (typeof v === "bigint" ? `${marker}${v.toString()}${marker}` : v));
  return json.replace(new RegExp(`"${marker}(\\d+)${marker}"`, "g"), "$1");
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
  // Zero-pad each coordinate to 32 bytes — the hex from /info is not
  // guaranteed fixed-width (a coordinate with a leading zero byte serializes
  // shorter), and writing it unpadded would misalign the byte layout
  // (code-review finding, ~1/256 chance per coordinate but real).
  pubKey.set(padTo32(hexToBytes(x)), 1);
  pubKey.set(padTo32(hexToBytes(y)), 33);
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
    const encodedPayload = { data: payload.data, signature: hexSignatureToBase64(payload.signature) };
    plaintext = new TextEncoder().encode(stringifyWithRawBigInts(encodedPayload));
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
