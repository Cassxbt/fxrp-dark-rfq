/** viem/wagmi throw rich objects; String() on them yields "[object Object]" and
 *  hides the reason. Prefer viem's shortMessage, then walk the cause chain. */
export function formatError(err: unknown): string {
  const parts: string[] = [];
  let cur: unknown = err;

  for (let depth = 0; cur && depth < 4; depth++) {
    const e = cur as Record<string, unknown>;
    const short = typeof e.shortMessage === "string" ? e.shortMessage : undefined;
    const details = typeof e.details === "string" ? e.details : undefined;
    const msg = typeof e.message === "string" ? e.message : undefined;

    const pick = short ?? details ?? msg;
    if (pick && !parts.includes(pick)) parts.push(pick);
    cur = e.cause;
  }

  if (parts.length) return parts.join(" — ").slice(0, 400);

  if (err instanceof Error) return err.message;
  if (typeof err === "string") return err;
  try {
    return JSON.stringify(err).slice(0, 400);
  } catch {
    return String(err);
  }
}
