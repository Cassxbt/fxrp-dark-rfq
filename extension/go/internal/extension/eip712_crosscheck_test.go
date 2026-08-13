package extension

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

// TestFillDigest_MatchesLiveContract cross-checks the Go EIP-712 implementation
// against the actual deployed RfqSettlement contract on Coston2 — not a
// reimplementation trusting itself, an independent on-chain oracle. If this
// ever fails, no signature produced in Go will ever validate on-chain, so
// this is the highest-value test in the whole extension.
//
// Expected digest obtained by calling hashFill directly on the live contract
// at 0xaBf47C48c00DDa806f1d9243c936A8153C7E6FcE on Coston2 (13 Aug 2026) with
// the exact same Fill values below via:
//
//	CALLDATA=$(cast calldata "hashFill((bytes32,address,address,uint8,uint256,uint256,uint256))" \
//	  "(0x1111...1111,0x2222...2222,0x3333...3333,0,1000000,2000000000000000000,9999999999)")
//	cast call 0xaBf47C48c00DDa806f1d9243c936A8153C7E6FcE --data "$CALLDATA" \
//	  --rpc-url https://coston2-api.flare.network/ext/C/rpc
func TestFillDigest_MatchesLiveContract(t *testing.T) {
	eip712 := eip712Config{
		chainID:           big.NewInt(114), // Coston2
		verifyingContract: common.HexToAddress("0xaBf47C48c00DDa806f1d9243c936A8153C7E6FcE"),
	}

	rfqID := common.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111")
	taker := common.HexToAddress("0x2222222222222222222222222222222222222222")
	maker := common.HexToAddress("0x3333333333333333333333333333333333333333")
	size := big.NewInt(1000000)
	price := new(big.Int).SetUint64(2000000000000000000) // 2e18
	expiry := uint64(9999999999)

	got := eip712.typedDataDigest(hashFillForSigning(rfqID, taker, maker, sideTakerBuy, size, price, expiry))

	want := common.HexToHash("0x56d8e79207bf2af713278d755ed767a321716654e27e6fcd79b088a043ba4439")

	if got != want {
		t.Fatalf("Go EIP-712 digest does not match the live contract's hashFill output.\ngot:  %s\nwant: %s\nThis means no signature produced in Go will validate on-chain — fix the encoding before doing anything else.", got.Hex(), want.Hex())
	}
}
