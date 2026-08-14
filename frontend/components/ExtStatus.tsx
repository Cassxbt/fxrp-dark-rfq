"use client";

import { useEffect, useState } from "react";

type State = "checking" | "up" | "down";

/**
 * Live enclave status. A hardcoded green dot is decoration; this one actually
 * probes /api/ext/info, which is the same path the taker and maker screens
 * depend on. If the dev machine behind the tunnel is asleep this reads "enclave
 * offline" instead of implying everything is fine.
 */
export function ExtStatus() {
  const [state, setState] = useState<State>("checking");

  useEffect(() => {
    let alive = true;
    const probe = async () => {
      try {
        const r = await fetch("/api/ext/info", { cache: "no-store" });
        const ok = r.ok && (await r.json())?.teeInfo != null;
        if (alive) setState(ok ? "up" : "down");
      } catch {
        if (alive) setState("down");
      }
    };
    probe();
    const t = setInterval(probe, 30_000);
    return () => {
      alive = false;
      clearInterval(t);
    };
  }, []);

  const { dot, label, title } = {
    checking: { dot: "bg-faint", label: "Coston2", title: "Checking the enclave…" },
    up: { dot: "bg-positive", label: "Coston2", title: "Enclave reachable" },
    down: {
      dot: "bg-negative",
      label: "enclave offline",
      title:
        "The extension runs on a dev machine behind a tunnel and is not reachable right now. The settled fills on the homepage are unaffected.",
    },
  }[state];

  return (
    <span className="flex items-center gap-1.5 text-[11px] text-faint" title={title}>
      <span className={`h-1.5 w-1.5 rounded-full ${dot}`} />
      {label}
    </span>
  );
}
