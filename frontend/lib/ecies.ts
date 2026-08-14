/**
 * Browser-safe port of go-ethereum's crypto/ecies package (AES128_SHA256
 * params) — this is what the TEE extension decrypts with (see
 * extension/go/internal/extension/utils.go's decryptViaNode call chain).
 *
 * Do not swap this for a generic ECIES library. go-ethereum's construction is
 * specific: NIST SP 800-56 concat-KDF (not HKDF), AES-128-CTR (not GCM/CBC),
 * and an HMAC-SHA256 tag over the ciphertext only (not AAD-bound). A
 * "standard" ECIES implementation will not produce compatible ciphertext.
 *
 * Verified byte-for-byte against the live deployed TEE before this file was
 * used anywhere near the UI. secp256k1 ECDH
 * uses @noble/curves because Web Crypto's ECDH only supports NIST P-curves,
 * not secp256k1; everything else (SHA-256, HMAC, AES-CTR) uses the Web Crypto
 * API directly since it's natively available in every target browser.
 */
import { secp256k1 } from "@noble/curves/secp256k1.js";

export function padTo32(bytes: Uint8Array): Uint8Array {
  if (bytes.length === 32) return bytes;
  if (bytes.length > 32) throw new Error(`unexpected length > 32: ${bytes.length}`);
  const out = new Uint8Array(32);
  out.set(bytes, 32 - bytes.length);
  return out;
}

function concatBytes(...arrays: Uint8Array[]): Uint8Array {
  const total = arrays.reduce((sum, a) => sum + a.length, 0);
  const out = new Uint8Array(total);
  let offset = 0;
  for (const a of arrays) {
    out.set(a, offset);
    offset += a.length;
  }
  return out;
}

async function sha256(data: Uint8Array): Promise<Uint8Array> {
  const digest = await crypto.subtle.digest("SHA-256", data.slice().buffer as ArrayBuffer);
  return new Uint8Array(digest);
}

// NIST SP 800-56 concat KDF, SHA-256, s1 (shared info) empty — matches
// go-ethereum's concatKDF exactly, including the big-endian uint32 counter.
async function concatKDF(z: Uint8Array, kdLen: number): Promise<Uint8Array> {
  const hashLen = 32;
  const chunks: Uint8Array[] = [];
  let produced = 0;
  for (let counter = 1; produced < kdLen; counter++) {
    const counterBuf = new Uint8Array(4);
    new DataView(counterBuf.buffer).setUint32(0, counter, false); // big-endian
    const h = await sha256(concatBytes(counterBuf, z));
    chunks.push(h);
    produced += hashLen;
  }
  return concatBytes(...chunks).slice(0, kdLen);
}

async function deriveKeys(z: Uint8Array, keyLen: number): Promise<{ Ke: Uint8Array; Km: Uint8Array }> {
  const K = await concatKDF(z, 2 * keyLen);
  const Ke = K.slice(0, keyLen);
  const KmRaw = K.slice(keyLen);
  const Km = await sha256(KmRaw);
  return { Ke, Km };
}

async function aes128Ctr(key: Uint8Array, iv: Uint8Array, data: Uint8Array): Promise<Uint8Array> {
  const cryptoKey = await crypto.subtle.importKey("raw", key.slice().buffer as ArrayBuffer, "AES-CTR", false, ["encrypt"]);
  // Web Crypto's AES-CTR counter is the full 16-byte IV block, matching
  // go-ethereum's cipher.NewCTR(c, iv) with a 16-byte IV.
  const result = await crypto.subtle.encrypt(
    { name: "AES-CTR", counter: iv.slice().buffer as ArrayBuffer, length: 128 },
    cryptoKey,
    data.slice().buffer as ArrayBuffer,
  );
  return new Uint8Array(result);
}

async function hmacSha256(key: Uint8Array, data: Uint8Array): Promise<Uint8Array> {
  const cryptoKey = await crypto.subtle.importKey(
    "raw",
    key.slice().buffer as ArrayBuffer,
    { name: "HMAC", hash: "SHA-256" },
    false,
    ["sign"],
  );
  const sig = await crypto.subtle.sign("HMAC", cryptoKey, data.slice().buffer as ArrayBuffer);
  return new Uint8Array(sig);
}

/**
 * Encrypts plaintext to a 65-byte uncompressed secp256k1 public key
 * (0x04 || X(32) || Y(32)), producing ciphertext shaped exactly like
 * go-ethereum's ecies.Encrypt: ephemeralPubkey(65) || IV(16) || ct || MAC(32).
 */
export async function eciesEncrypt(pubKeyUncompressed: Uint8Array, plaintext: Uint8Array): Promise<Uint8Array> {
  if (pubKeyUncompressed.length !== 65 || pubKeyUncompressed[0] !== 0x04) {
    throw new Error("expected a 65-byte uncompressed secp256k1 public key starting with 0x04");
  }

  const ephemeralPriv = secp256k1.utils.randomSecretKey();
  const ephemeralPub = secp256k1.getPublicKey(ephemeralPriv, false); // uncompressed, 65 bytes

  // ECDH shared point; take the X coordinate, big-endian, zero-padded to 32
  // bytes — matches Go's GenerateShared/x.Bytes() semantics exactly.
  const sharedPoint = secp256k1.getSharedSecret(ephemeralPriv, pubKeyUncompressed, false); // 65 bytes, 0x04||X||Y
  const sharedX = sharedPoint.slice(1, 33);
  const z = padTo32(sharedX);

  const keyLen = 16; // AES-128
  const { Ke, Km } = await deriveKeys(z, keyLen);

  const iv = crypto.getRandomValues(new Uint8Array(16));
  const ciphertext = await aes128Ctr(Ke, iv, plaintext);
  const em = concatBytes(iv, ciphertext);

  const tag = await hmacSha256(Km, em);

  return concatBytes(ephemeralPub, em, tag);
}

export function bytesToHex(bytes: Uint8Array): string {
  return "0x" + Array.from(bytes).map((b) => b.toString(16).padStart(2, "0")).join("");
}

export function hexToBytes(hex: string): Uint8Array {
  const clean = hex.startsWith("0x") ? hex.slice(2) : hex;
  const out = new Uint8Array(clean.length / 2);
  for (let i = 0; i < out.length; i++) {
    out[i] = parseInt(clean.substr(i * 2, 2), 16);
  }
  return out;
}
