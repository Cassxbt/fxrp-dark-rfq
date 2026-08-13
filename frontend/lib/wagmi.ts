import { http, createConfig } from "wagmi";
import { injected } from "wagmi/connectors";
import { coston2 } from "./chain";

// injected() covers MetaMask and any other browser-extension wallet — no
// WalletConnect project ID to manage for an MVP demo.
export const wagmiConfig = createConfig({
  chains: [coston2],
  connectors: [injected()],
  transports: {
    [coston2.id]: http(),
  },
});
