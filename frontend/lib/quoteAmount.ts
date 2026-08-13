/**
 * Must match contracts/src/RfqSettlement.sol's settle() exactly:
 * Math.mulDiv(fill.size * fill.price, 10**quoteDecimals, 10**baseDecimals * 1e18).
 * Previously duplicated between taker and maker pages (code-review finding) —
 * a future edit to one copy without the other would silently desync approve()
 * amounts from what the contract actually pulls.
 */
export function quoteAmount(size: bigint, priceWad: bigint, baseDecimals: number, quoteDecimals: number): bigint {
  return (size * priceWad * 10n ** BigInt(quoteDecimals)) / (10n ** BigInt(baseDecimals) * 10n ** 18n);
}
