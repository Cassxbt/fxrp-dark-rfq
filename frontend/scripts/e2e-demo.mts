// Drives the real OPEN -> QUOTE x2 -> CLOSE -> Filled flow against the live
// Coston2 contract and the running FCC extension, using three funded burner
// keys (never committed — see .env.e2e.local, gitignored). Reuses the same
// lib/ code the actual UI calls (eip712, rfqClient, quoteAmount) so this
// proves the real code paths, not a reimplementation that could drift.
//
// Run: npx tsx --env-file=.env.e2e.local scripts/e2e-demo.mts
import { createPublicClient, createWalletClient, http, parseUnits, decodeEventLog } from "viem";
import { privateKeyToAccount } from "viem/accounts";
import { defineChain } from "viem";
import {
  RFQ_SETTLEMENT_ADDRESS,
  FXRP_ADDRESS,
  USDT0_ADDRESS,
  ERC20_ABI,
  RFQ_SETTLEMENT_ABI,
} from "../lib/contracts";
import { EIP712_DOMAIN, RFQ_INTENT_TYPES, QUOTE_TYPES, SIDE_TAKER_BUY, SIDE_TAKER_SELL } from "../lib/eip712";
import { sendRfqDirect } from "../lib/rfqClient";
import { hexToBytes } from "../lib/ecies";
import { quoteAmount } from "../lib/quoteAmount";

const FXRP_DECIMALS = 6;
const USDT0_DECIMALS = 6;

const coston2 = defineChain({
  id: 114,
  name: "Coston2",
  nativeCurrency: { name: "C2FLR", symbol: "C2FLR", decimals: 18 },
  rpcUrls: { default: { http: ["https://coston2-api.flare.network/ext/C/rpc"] } },
});

function requireEnv(name: string): `0x${string}` {
  const v = process.env[name];
  if (!v) throw new Error(`missing env ${name}`);
  return v as `0x${string}`;
}

const takerAccount = privateKeyToAccount(requireEnv("TAKER_PRIVATE_KEY"));
const maker1Account = privateKeyToAccount(requireEnv("MAKER1_PRIVATE_KEY"));
const maker2Account = privateKeyToAccount(requireEnv("MAKER2_PRIVATE_KEY"));

const publicClient = createPublicClient({ chain: coston2, transport: http() });
const takerWallet = createWalletClient({ account: takerAccount, chain: coston2, transport: http() });
const maker1Wallet = createWalletClient({ account: maker1Account, chain: coston2, transport: http() });
const maker2Wallet = createWalletClient({ account: maker2Account, chain: coston2, transport: http() });

function log(step: string, detail?: unknown) {
  console.log(`\n[${new Date().toISOString().slice(11, 19)}] ${step}`);
  if (detail !== undefined) console.log(JSON.stringify(detail, (_k, v) => (typeof v === "bigint" ? v.toString() : v), 2));
}

async function checkBalances() {
  log("Checking balances/gas for all three accounts...");
  for (const [name, account] of [
    ["taker", takerAccount],
    ["maker1", maker1Account],
    ["maker2", maker2Account],
  ] as const) {
    const gas = await publicClient.getBalance({ address: account.address });
    const fxrp = await publicClient.readContract({
      address: FXRP_ADDRESS,
      abi: ERC20_ABI,
      functionName: "balanceOf",
      args: [account.address],
    });
    const usdt0 = await publicClient.readContract({
      address: USDT0_ADDRESS,
      abi: ERC20_ABI,
      functionName: "balanceOf",
      args: [account.address],
    });
    log(`${name} (${account.address})`, {
      gas: gas.toString(),
      fxrp: fxrp.toString(),
      usdt0: usdt0.toString(),
    });
    if (gas === 0n) throw new Error(`${name} has no C2FLR for gas`);
  }
}

async function main() {
  await checkBalances();

  // SIDE=sell exercises the other half of the matcher (taker sells FXRP,
  // highest qualifying quote wins) and rebalances wallets that have drifted
  // one way after repeated same-side runs.
  const sell = process.env.SIDE === "sell";
  const size = "1";
  const limitPrice = sell ? "2.50" : "3.00";
  const sizeUnits = parseUnits(size, FXRP_DECIMALS);
  const limitPriceWad = parseUnits(limitPrice, 18);
  const expiry = BigInt(Math.floor(Date.now() / 1000) + 900);
  const rfqNonce = BigInt(Date.now()) * 1000n + BigInt(Math.floor(Math.random() * 1000));

  // Buy: approve USDT0 at the worst case (the taker's own limit).
  // Sell: approve the exact FXRP being sold.
  const approveAmount = sell
    ? sizeUnits
    : quoteAmount(sizeUnits, limitPriceWad, FXRP_DECIMALS, USDT0_DECIMALS);
  log(`Taker: approving ${sell ? "FXRP" : "USDT0"}`, { approveAmount: approveAmount.toString() });
  const approveTakerTx = await takerWallet.writeContract({
    address: sell ? FXRP_ADDRESS : USDT0_ADDRESS,
    abi: ERC20_ABI,
    functionName: "approve",
    args: [RFQ_SETTLEMENT_ADDRESS, approveAmount],
  });
  await publicClient.waitForTransactionReceipt({ hash: approveTakerTx });
  log("Taker: approve confirmed", { tx: approveTakerTx });

  const intent = {
    side: sell ? SIDE_TAKER_SELL : SIDE_TAKER_BUY,
    size: sizeUnits,
    limitPrice: limitPriceWad,
    taker: takerAccount.address,
    expiry,
    rfqNonce,
  };
  log(`Taker: signing RfqIntent (${sell ? "sell" : "buy"} ${size} FXRP @ limit ${limitPrice} USDT0)`);
  const takerSig = await takerWallet.signTypedData({
    domain: EIP712_DOMAIN,
    types: RFQ_INTENT_TYPES,
    primaryType: "RfqIntent",
    message: intent,
  });

  log("Taker: sending OPEN to the TEE");
  const openResult = await sendRfqDirect("OPEN", { data: intent, signature: takerSig });
  if (openResult.status !== 1) throw new Error(`OPEN failed: ${openResult.log}`);
  const opened = JSON.parse(new TextDecoder().decode(hexToBytes(openResult.data)));
  const rfqId = opened.rfqId as `0x${string}`;
  log("OPEN succeeded", { rfqId });

  // Two makers, two different prices, so the winner is actually a selection
  // decision made inside the TEE, not the only option on the table.
  // On a sell the taker wants the HIGHEST price, so the winner flips.
  const makerQuotes: { wallet: typeof maker1Wallet; account: typeof maker1Account; price: string }[] = [
    { wallet: maker1Wallet, account: maker1Account, price: sell ? "2.55" : "2.95" },
    { wallet: maker2Wallet, account: maker2Account, price: sell ? "2.60" : "2.99" },
  ];

  for (const [i, { wallet, account, price }] of makerQuotes.entries()) {
    const priceWad = parseUnits(price, 18);
    const quoteExpiry = BigInt(Math.floor(Date.now() / 1000) + 300);

    // Buy: makers deliver FXRP. Sell: makers pay USDT0 at their own price.
    const makerApprove = sell
      ? quoteAmount(sizeUnits, priceWad, FXRP_DECIMALS, USDT0_DECIMALS)
      : sizeUnits;
    log(`Maker${i + 1} (${account.address}): approving ${sell ? "USDT0" : "FXRP"}`);
    const approveTx = await wallet.writeContract({
      address: sell ? USDT0_ADDRESS : FXRP_ADDRESS,
      abi: ERC20_ABI,
      functionName: "approve",
      args: [RFQ_SETTLEMENT_ADDRESS, makerApprove],
    });
    await publicClient.waitForTransactionReceipt({ hash: approveTx });

    const quote = { rfqId, price: priceWad, maker: account.address, expiry: quoteExpiry };
    log(`Maker${i + 1}: signing Quote @ ${price}`);
    const quoteSig = await wallet.signTypedData({
      domain: EIP712_DOMAIN,
      types: QUOTE_TYPES,
      primaryType: "Quote",
      message: quote,
    });

    log(`Maker${i + 1}: sending QUOTE to the TEE`);
    const quoteResult = await sendRfqDirect("QUOTE", { data: quote, signature: quoteSig });
    if (quoteResult.status !== 1) throw new Error(`QUOTE (maker${i + 1}) failed: ${quoteResult.log}`);
    log(`Maker${i + 1}: QUOTE accepted`);
  }

  const expectedWinner = sell ? maker2Account : maker1Account;
  const expectedPrice = sell ? "2.60" : "2.95";
  log(`Taker: sending CLOSE (expect ${sell ? "maker2 @ 2.60 — highest wins on a sell" : "maker1 @ 2.95 — lowest wins on a buy"})`);
  const closeResult = await sendRfqDirect("CLOSE", hexToBytes(rfqId));
  if (closeResult.status !== 1) throw new Error(`CLOSE failed: ${closeResult.log}`);
  const closed = JSON.parse(new TextDecoder().decode(hexToBytes(closeResult.data)));
  log("CLOSE result", closed);
  if (!closed.matched) throw new Error("CLOSE reported no match — nothing to settle");

  log("Polling settled() for up to ~60s...");
  let settledOk = false;
  for (let i = 0; i < 30; i++) {
    await new Promise((r) => setTimeout(r, 2000));
    const isSettled = await publicClient.readContract({
      address: RFQ_SETTLEMENT_ADDRESS,
      abi: RFQ_SETTLEMENT_ABI,
      functionName: "settled",
      args: [rfqId],
    });
    if (isSettled) {
      settledOk = true;
      break;
    }
  }
  if (!settledOk) throw new Error("settled() never flipped true within 60s");
  log("settled() == true, on-chain settlement confirmed");

  log("Fetching Filled event for this rfqId...");
  // Coston2's public RPC caps eth_getLogs at 30 blocks per request, so scan
  // recent history rather than from genesis.
  const latestBlock = await publicClient.getBlockNumber();
  const logs = await publicClient.getContractEvents({
    address: RFQ_SETTLEMENT_ADDRESS,
    abi: RFQ_SETTLEMENT_ABI,
    eventName: "Filled",
    args: { rfqId },
    fromBlock: latestBlock - 25n,
    toBlock: "latest",
  });
  if (logs.length === 0) throw new Error("settled() is true but no Filled event found");
  const fill = logs[logs.length - 1];
  log("Filled event", { txHash: fill.transactionHash, ...fill.args });
  log(
    `Explorer: https://coston2-explorer.flare.network/tx/${fill.transactionHash}`,
  );

  const winnerOk = (fill.args as { maker?: string }).maker?.toLowerCase() === expectedWinner.address.toLowerCase();
  log(winnerOk
    ? `PASS: best-priced maker won at ${expectedPrice}, as expected`
    : "NOTE: winner was not the expected best-priced maker — check selection logic");

  await checkBalances();
  log("E2E DEMO COMPLETE");
}

main().catch((e) => {
  console.error("\nE2E DEMO FAILED:", e);
  process.exit(1);
});
