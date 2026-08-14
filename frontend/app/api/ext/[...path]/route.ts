// Server-side proxy to the FCC extension's tunnel.
//
// The browser cannot call the tunnel directly, for two reasons that only show
// up in a real browser and so were invisible to the Node e2e script:
//
//  1. The ext-proxy sends no Access-Control-Allow-Origin and answers OPTIONS
//     with 405. Our POST uses Content-Type: application/json, which is not a
//     CORS-safelisted type, so it requires a preflight — and that preflight
//     fails. This blocked the browser from localhost just as much as from a
//     deployed origin.
//  2. ngrok's free tier serves an HTML interstitial to anything with a
//     browser User-Agent, so fetch() got a warning page instead of JSON.
//
// Proxying server-side removes both: this request is same-origin from the
// browser's point of view, and server-to-server fetch has no CORS at all.
// It also keeps the tunnel URL out of the client bundle, so rotating the
// tunnel is an env change rather than a rebuild.

const EXT_ORIGIN = process.env.EXT_PROXY_ORIGIN ?? "https://unappliably-unphased-josef.ngrok-free.dev";

async function forward(req: Request, path: string[]) {
  const url = new URL(req.url);
  const target = `${EXT_ORIGIN}/${path.join("/")}${url.search}`;

  const init: RequestInit = {
    method: req.method,
    headers: {
      "Content-Type": "application/json",
      // Any value works; its presence is what suppresses the interstitial.
      "ngrok-skip-browser-warning": "true",
    },
    // Never cache: /action/result is polled and must return fresh state.
    cache: "no-store",
  };
  if (req.method !== "GET" && req.method !== "HEAD") {
    init.body = await req.text();
  }

  try {
    const resp = await fetch(target, init);
    const body = await resp.text();
    return new Response(body, {
      status: resp.status,
      headers: { "Content-Type": resp.headers.get("content-type") ?? "application/json" },
    });
  } catch (err) {
    // The tunnel is down (laptop asleep, ngrok restarted). Say so plainly
    // rather than surfacing an opaque network error in the UI.
    return Response.json(
      {
        error: "extension unreachable",
        detail: err instanceof Error ? err.message : String(err),
        hint: "The FCC extension is served through a tunnel from a dev machine. If it is down, the on-chain fill linked from the homepage remains the evidence of record.",
      },
      { status: 502 },
    );
  }
}

export async function GET(req: Request, ctx: { params: Promise<{ path: string[] }> }) {
  return forward(req, (await ctx.params).path);
}

export async function POST(req: Request, ctx: { params: Promise<{ path: string[] }> }) {
  return forward(req, (await ctx.params).path);
}
