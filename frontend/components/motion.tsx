"use client";

import { useEffect, useRef, useState } from "react";

/** Fade-and-rise, once. Starts visible and only arms after mount, so without
 *  JS the page reads finished rather than blank. */
export function Reveal({
  children,
  delay = 0,
  className = "",
}: {
  children: React.ReactNode;
  delay?: number;
  className?: string;
}) {
  const ref = useRef<HTMLDivElement>(null);
  const [armed, setArmed] = useState(false);
  const [shown, setShown] = useState(false);

  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) return;

    setArmed(true);

    // Anything already on screen reveals immediately; waiting for a scroll
    // that never comes would leave the hero blank.
    const io = new IntersectionObserver(
      ([e]) => {
        if (e.isIntersecting) {
          setShown(true);
          io.disconnect();
        }
      },
      { rootMargin: "0px 0px -12% 0px", threshold: 0.05 },
    );
    io.observe(el);
    return () => io.disconnect();
  }, []);

  return (
    <div
      ref={ref}
      className={`${armed ? "reveal" : ""} ${shown ? "is-in" : ""} ${className}`}
      style={delay ? ({ "--reveal-delay": `${delay}ms` } as React.CSSProperties) : undefined}
    >
      {children}
    </div>
  );
}

/** Covers the webfont swap on first paint. Leaves on a timer whatever else
 *  happens, and is skipped entirely under reduced motion. */
export function Curtain() {
  const [leaving, setLeaving] = useState(false);
  const [gone, setGone] = useState(false);

  useEffect(() => {
    if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) {
      setGone(true);
      return;
    }
    // Hard cap: a slow font CDN must not hold the page hostage.
    const ready = Promise.race([
      (document as Document & { fonts?: FontFaceSet }).fonts?.ready ?? Promise.resolve(),
      new Promise((r) => setTimeout(r, 900)),
    ]);
    let t2: ReturnType<typeof setTimeout>;
    ready.then(() => {
      setLeaving(true);
      t2 = setTimeout(() => setGone(true), 440);
    });
    return () => clearTimeout(t2);
  }, []);

  if (gone) return null;

  return (
    <div className="curtain" data-leaving={leaving} aria-hidden="true">
      <div className="flex flex-col items-center gap-4">
        <div className="flex items-center gap-2.5">
          <span className="h-2.5 w-2.5 rounded-[2px] bg-accent" />
          <span className="display text-[19px] tracking-[0.02em] text-ink">FXRP Dark RFQ</span>
        </div>
        <span className="curtain-rule" />
      </div>
    </div>
  );
}
