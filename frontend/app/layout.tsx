import type { Metadata } from "next";
import { Geist, Geist_Mono, Cormorant_Garamond } from "next/font/google";
import "./globals.css";
import { Providers } from "./providers";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

// Display only — headings and editorial ledes. Never the order tickets.
const cormorant = Cormorant_Garamond({
  variable: "--font-cormorant",
  subsets: ["latin"],
  weight: ["400", "600"],
  style: ["normal", "italic"],
});

export const metadata: Metadata = {
  title: "FXRP Dark RFQ",
  description: "Sealed-bid RFQ desk for FXRP, matched inside a Flare Confidential Compute TEE.",
};

// Dark-only surface: tells the browser to render scrollbars and native
// controls to match, instead of light chrome around a dark page.
export const viewport = { colorScheme: "dark" as const, themeColor: "#070709" };

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html
      lang="en"
      className={`${geistSans.variable} ${geistMono.variable} ${cormorant.variable} antialiased`}
    >
      {/* No flex here on purpose — pages own their own full-height layout, and a
          flex body silently breaks `mx-auto max-w-*` children by shrink-wrapping
          them (auto margins beat align-items: stretch on the cross axis). */}
      <body>
        <Providers>{children}</Providers>
      </body>
    </html>
  );
}
