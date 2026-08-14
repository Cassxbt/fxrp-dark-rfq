import { eciesEncrypt, bytesToHex, hexToBytes, padTo32 } from "./ecies";
import { EXT_PROXY_URL } from "./contracts";

function opHash(s: string): `0x${string}` {
  const bytes = new TextEncoder().encode(s);
  if (bytes.length > 32) throw new Error(`op string too long: ${s}`);
  const padded = new Uint8Array(32);
  padded.set(bytes);
  return bytesToHex(padded) as `0x${string}`;
}

/** Go decodes a plain `[]byte` field as base64, not hex. Sending wagmi's hex
 *  signature through unconverted yields 99 garbage bytes and fails every
 *  signature check. */
function hexSignatureToBase64(hexSig: `0x${string}`): string {
  const bytes = hexToBytes(hexSig);
  let binary = "";
  for (const b of bytes) binary += String.fromCharCode(b);
  return btoa(binary);
}

/** Go's big.Int requires bare JSON numbers and rejects quoted strings. WAD
 *  values exceed MAX_SAFE_INTEGER, so this round-trips through a sentinel
 *  rather than Number(). */
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
  // /info's hex is not fixed-width; an unpadded coordinate misaligns the layout.
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

/** Encrypts to the TEE's pubkey, wraps with the signature, dispatches via
 *  /direct and polls for the result. */
export async function sendRfqDirect(
  opCommand: "OPEN" | "QUOTE" | "CLOSE",
  payload: { data: unknown; signature: `0x${string}` } | Uint8Array,
): Promise<DirectResult> {
  const pubKey = await fetchTeePubKey();

  let plaintext: Uint8Array;
  if (payload instanceof Uint8Array) {
    // CLOSE carries no secret, so it goes as a plaintext 32-byte rfqId.
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

  // Direct actions are tagged "submit"; without this the endpoint defaults to
  // "threshold" and silently returns nothing.
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
