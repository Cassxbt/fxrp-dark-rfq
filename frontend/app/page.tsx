import Link from "next/link";

export default function Home() {
  return (
    <main className="mx-auto flex max-w-xl flex-1 flex-col items-center justify-center gap-8 p-8 text-center">
      <div>
        <h1 className="text-2xl font-semibold">FXRP Dark RFQ</h1>
        <p className="mt-2 text-neutral-600">
          Sealed-bid RFQ desk for FXRP. Matching happens privately inside a Flare Confidential Compute TEE — the
          chain only sees the fill.
        </p>
      </div>
      <div className="flex gap-4">
        <Link href="/taker" className="rounded bg-black px-6 py-3 text-white hover:bg-neutral-800">
          I&apos;m a Taker
        </Link>
        <Link href="/maker" className="rounded border px-6 py-3 hover:bg-neutral-50">
          I&apos;m a Maker
        </Link>
      </div>
    </main>
  );
}
