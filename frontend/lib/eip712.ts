import { RFQ_SETTLEMENT_ADDRESS } from "./contracts";

// Must match extension/go/internal/extension/rfq.go's eip712Config and
// contracts/src/RfqSettlement.sol's EIP712("RfqSettlement", "1") exactly —
// domain mismatch means a signature that looks valid here recovers to the
// wrong address on the Go side. Field order and types below also must match
// rfqIntentTypehash / quoteTypehash exactly, same reason.
export const EIP712_DOMAIN = {
  name: "RfqSettlement",
  version: "1",
  chainId: 114,
  verifyingContract: RFQ_SETTLEMENT_ADDRESS,
} as const;

export const RFQ_INTENT_TYPES = {
  RfqIntent: [
    { name: "side", type: "uint8" },
    { name: "size", type: "uint256" },
    { name: "limitPrice", type: "uint256" },
    { name: "taker", type: "address" },
    { name: "expiry", type: "uint256" },
    { name: "rfqNonce", type: "uint256" },
  ],
} as const;

export const QUOTE_TYPES = {
  Quote: [
    { name: "rfqId", type: "bytes32" },
    { name: "price", type: "uint256" },
    { name: "maker", type: "address" },
    { name: "expiry", type: "uint256" },
  ],
} as const;

export const SIDE_TAKER_BUY = 0;
export const SIDE_TAKER_SELL = 1;
