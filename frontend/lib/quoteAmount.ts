/**
 * Must match RfqSettlement.sol's settle() exactly:
 * Math.mulDiv(fill.size, fill.price * 10**quoteDecimals, 10**baseDecimals * 1e18).
 * Shared by the taker and maker pages so an edit to one can't silently desync
 * approve() amounts from what the contract actually pulls.
 */
export function quoteAmount(size: bigint, priceWad: bigint, baseDecimals: number, quoteDecimals: number): bigint {
  return (size * priceWad * 10n ** BigInt(quoteDecimals)) / (10n ** BigInt(baseDecimals) * 10n ** 18n);
}
