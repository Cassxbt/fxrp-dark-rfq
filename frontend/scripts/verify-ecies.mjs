// Verifies the exact algorithm used in frontend/lib/ecies.ts (secp256k1 via
// @noble/curves + Web Crypto for SHA-256/HMAC/AES-CTR) against the live TEE.
// Node 22's globalThis.crypto.subtle is spec-compliant Web Crypto, so this
// exercises the same code path the browser will actually run.
import { secp256k1 } from "@noble/curves/secp256k1.js";

function padTo32(bytes) {
  if (bytes.length === 32) return bytes;
  const out = new Uint8Array(32);
  out.set(bytes, 32 - bytes.length);
  return out;
}
function concatBytes(...arrays) {
  const total = arrays.reduce((s, a) => s + a.length, 0);
  const out = new Uint8Array(total);
  let o = 0;
  for (const a of arrays) { out.set(a, o); o += a.length; }
  return out;
}
async function sha256(data) {
  return new Uint8Array(await crypto.subtle.digest("SHA-256", data.slice().buffer));
}
async function concatKDF(z, kdLen) {
  const chunks = [];
  let produced = 0;
  for (let counter = 1; produced < kdLen; counter++) {
    const cb = new Uint8Array(4);
    new DataView(cb.buffer).setUint32(0, counter, false);
    chunks.push(await sha256(concatBytes(cb, z)));
    produced += 32;
  }
  return concatBytes(...chunks).slice(0, kdLen);
}
async function deriveKeys(z, keyLen) {
  const K = await concatKDF(z, 2 * keyLen);
  const Ke = K.slice(0, keyLen);
  const Km = await sha256(K.slice(keyLen));
  return { Ke, Km };
}
async function aes128Ctr(key, iv, data) {
  const ck = await crypto.subtle.importKey("raw", key.slice().buffer, "AES-CTR", false, ["encrypt"]);
  const r = await crypto.subtle.encrypt({ name: "AES-CTR", counter: iv.slice().buffer, length: 128 }, ck, data.slice().buffer);
  return new Uint8Array(r);
}
async function hmacSha256(key, data) {
  const ck = await crypto.subtle.importKey("raw", key.slice().buffer, { name: "HMAC", hash: "SHA-256" }, false, ["sign"]);
  return new Uint8Array(await crypto.subtle.sign("HMAC", ck, data.slice().buffer));
}
async function eciesEncrypt(pubKeyUncompressed, plaintext) {
  const ephemeralPriv = secp256k1.utils.randomSecretKey();
  const ephemeralPub = secp256k1.getPublicKey(ephemeralPriv, false);
  const sharedPoint = secp256k1.getSharedSecret(ephemeralPriv, pubKeyUncompressed, false);
  const z = padTo32(sharedPoint.slice(1, 33));
  const { Ke, Km } = await deriveKeys(z, 16);
  const iv = crypto.getRandomValues(new Uint8Array(16));
  const ciphertext = await aes128Ctr(Ke, iv, plaintext);
  const em = concatBytes(iv, ciphertext);
  const tag = await hmacSha256(Km, em);
  return concatBytes(ephemeralPub, em, tag);
}

const proxyURL = "https://2971-102-89-68-144.ngrok-free.app";
const info = await (await fetch(proxyURL + "/info")).json();
const { x, y } = info.machineData.publicKey;
const hexToBytes = (h) => Uint8Array.from(Buffer.from(h.replace(/^0x/, ""), "hex"));
const pubKey = concatBytes(new Uint8Array([0x04]), hexToBytes(x), hexToBytes(y));
console.log("TEE pubkey length:", pubKey.length, "(expect 65)");

const plaintext = new TextEncoder().encode(JSON.stringify({ probe: "webcrypto-ecies-compat-check" }));
const ciphertext = await eciesEncrypt(pubKey, plaintext);
console.log("Ciphertext length:", ciphertext.length);

function opHash(s) {
  const padded = new Uint8Array(32);
  padded.set(new TextEncoder().encode(s));
  return "0x" + Buffer.from(padded).toString("hex");
}
const bytesToHex = (b) => "0x" + Buffer.from(b).toString("hex");

const payload = { opType: opHash("RFQ"), opCommand: opHash("OPEN"), message: bytesToHex(ciphertext) };
const directResp = await fetch(proxyURL + "/direct", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify(payload),
});
const directJson = await directResp.json();
console.log("POST /direct status:", directResp.status);
const actionId = directJson.data.id;

await new Promise((r) => setTimeout(r, 3000));
const resultJson = await (await fetch(proxyURL + "/action/result/" + actionId + "?submissionTag=submit")).json();
console.log("Result:", JSON.stringify(resultJson.result, null, 2));

const log = resultJson.result.log || "";
if (log.includes("can not decrypt")) {
  console.log("\nFAIL: Web Crypto + @noble/curves ECIES implementation is NOT compatible.");
  process.exit(1);
} else {
  console.log("\nPASS: Web Crypto + @noble/curves ECIES implementation IS compatible — safe to use in the browser frontend.");
}
