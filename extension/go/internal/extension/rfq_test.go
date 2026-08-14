package extension

import (
	"math/big"
	"testing"
	"time"

	dcrsecp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// privKeyToAddress derives the Ethereum address for a decred secp256k1 key by
// round-tripping its uncompressed public key through go-ethereum's crypto
// package — the two libraries use the same curve but different types.
func privKeyToAddress(t *testing.T, key *dcrsecp256k1.PrivateKey) common.Address {
	t.Helper()
	pub, err := crypto.UnmarshalPubkey(key.PubKey().SerializeUncompressed())
	if err != nil {
		t.Fatalf("unmarshaling pubkey: %v", err)
	}
	return crypto.PubkeyToAddress(*pub)
}

func TestRecoverSigner_RoundTrip(t *testing.T) {
	keyBytes := make([]byte, 32)
	keyBytes[31] = 0x42
	key := dcrsecp256k1.PrivKeyFromBytes(keyBytes)
	expectedAddr := privKeyToAddress(t, key)

	digest := common.HexToHash("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")

	sig, err := signDigest(key, digest.Bytes())
	if err != nil {
		t.Fatalf("signDigest: %v", err)
	}

	recovered, err := recoverSigner(digest, sig)
	if err != nil {
		t.Fatalf("recoverSigner: %v", err)
	}
	if recovered != expectedAddr {
		t.Fatalf("recovered %s, want %s", recovered.Hex(), expectedAddr.Hex())
	}
}

func TestRecoverSigner_WrongSignerDoesNotMatch(t *testing.T) {
	key1Bytes := make([]byte, 32)
	key1Bytes[31] = 0x01
	key1 := dcrsecp256k1.PrivKeyFromBytes(key1Bytes)

	key2Bytes := make([]byte, 32)
	key2Bytes[31] = 0x02
	key2 := dcrsecp256k1.PrivKeyFromBytes(key2Bytes)
	key2Addr := privKeyToAddress(t, key2)

	digest := common.HexToHash("0xcafebabecafebabecafebabecafebabecafebabecafebabecafebabecafebabe")
	sig, err := signDigest(key1, digest.Bytes())
	if err != nil {
		t.Fatalf("signDigest: %v", err)
	}

	recovered, err := recoverSigner(digest, sig)
	if err != nil {
		t.Fatalf("recoverSigner: %v", err)
	}
	if recovered == key2Addr {
		t.Fatalf("recovered address incorrectly matched a different key's address")
	}
}

func mkQuote(maker common.Address, price int64, receivedAt time.Time) quoteEntry {
	return quoteEntry{
		Quote:      quote{Price: big.NewInt(price), Maker: maker, Expiry: uint64(time.Now().Add(time.Hour).Unix())},
		ReceivedAt: receivedAt,
	}
}

func TestSelectWinner_TakerBuy_PicksLowestQualifyingPrice(t *testing.T) {
	now := time.Now()
	makerA := common.HexToAddress("0xaaaa000000000000000000000000000000aaaa")
	makerB := common.HexToAddress("0xbbbb000000000000000000000000000000bbbb")
	makerC := common.HexToAddress("0xcccc000000000000000000000000000000cccc") // does not qualify

	intent := rfqIntent{Side: sideTakerBuy, LimitPrice: big.NewInt(200), Expiry: uint64(now.Add(time.Hour).Unix())}
	quotes := map[common.Address]quoteEntry{
		makerA: mkQuote(makerA, 150, now),
		makerB: mkQuote(makerB, 120, now.Add(time.Second)), // lowest, should win
		makerC: mkQuote(makerC, 999, now),                  // above limit, must not qualify
	}

	winner, ok := selectWinner(intent, quotes, now)
	if !ok {
		t.Fatal("expected a winner")
	}
	if winner.Quote.Maker != makerB {
		t.Fatalf("expected makerB (lowest qualifying price) to win, got %s", winner.Quote.Maker.Hex())
	}
}

func TestSelectWinner_TakerSell_PicksHighestQualifyingPrice(t *testing.T) {
	now := time.Now()
	makerA := common.HexToAddress("0xaaaa000000000000000000000000000000aaaa")
	makerB := common.HexToAddress("0xbbbb000000000000000000000000000000bbbb")

	intent := rfqIntent{Side: sideTakerSell, LimitPrice: big.NewInt(100), Expiry: uint64(now.Add(time.Hour).Unix())}
	quotes := map[common.Address]quoteEntry{
		makerA: mkQuote(makerA, 110, now),
		makerB: mkQuote(makerB, 150, now.Add(time.Second)), // highest, should win
	}

	winner, ok := selectWinner(intent, quotes, now)
	if !ok {
		t.Fatal("expected a winner")
	}
	if winner.Quote.Maker != makerB {
		t.Fatalf("expected makerB (highest qualifying price) to win, got %s", winner.Quote.Maker.Hex())
	}
}

func TestSelectWinner_TieBrokenByEarliestReceived(t *testing.T) {
	now := time.Now()
	makerEarly := common.HexToAddress("0xe000000000000000000000000000000000e000")
	makerLate := common.HexToAddress("0x1000000000000000000000000000000000000")

	intent := rfqIntent{Side: sideTakerBuy, LimitPrice: big.NewInt(200), Expiry: uint64(now.Add(time.Hour).Unix())}
	quotes := map[common.Address]quoteEntry{
		makerLate:  mkQuote(makerLate, 100, now.Add(5*time.Second)),
		makerEarly: mkQuote(makerEarly, 100, now), // same price, earlier — should win
	}

	winner, ok := selectWinner(intent, quotes, now)
	if !ok {
		t.Fatal("expected a winner")
	}
	if winner.Quote.Maker != makerEarly {
		t.Fatalf("expected earliest-received quote to win a tie, got %s", winner.Quote.Maker.Hex())
	}
}

func TestSelectWinner_NoQualifyingQuote_NoMatch(t *testing.T) {
	now := time.Now()
	maker := common.HexToAddress("0xaaaa000000000000000000000000000000aaaa")

	intent := rfqIntent{Side: sideTakerBuy, LimitPrice: big.NewInt(100), Expiry: uint64(now.Add(time.Hour).Unix())}
	quotes := map[common.Address]quoteEntry{
		maker: mkQuote(maker, 200, now), // above limit
	}

	_, ok := selectWinner(intent, quotes, now)
	if ok {
		t.Fatal("expected no match — the only quote is above the taker's limit")
	}
}

func TestSelectWinner_ExpiredQuoteExcluded(t *testing.T) {
	now := time.Now()
	maker := common.HexToAddress("0xaaaa000000000000000000000000000000aaaa")

	intent := rfqIntent{Side: sideTakerBuy, LimitPrice: big.NewInt(200), Expiry: uint64(now.Add(time.Hour).Unix())}
	quotes := map[common.Address]quoteEntry{
		maker: {Quote: quote{Price: big.NewInt(100), Maker: maker, Expiry: uint64(now.Add(-time.Minute).Unix())}, ReceivedAt: now},
	}

	_, ok := selectWinner(intent, quotes, now)
	if ok {
		t.Fatal("expected no match — the only quote is already expired")
	}
}

// Regression test: selectWinner previously only
// checked each quote's own expiry, never the taker's RfqIntent expiry — a
// late CLOSE could still pick a winner and sign+submit a doomed on-chain tx.
func TestSelectWinner_ExpiredIntentExcludesEvenAValidQuote(t *testing.T) {
	now := time.Now()
	maker := common.HexToAddress("0xaaaa000000000000000000000000000000aaaa")

	intent := rfqIntent{Side: sideTakerBuy, LimitPrice: big.NewInt(200), Expiry: uint64(now.Add(-time.Second).Unix())}
	quotes := map[common.Address]quoteEntry{
		maker: mkQuote(maker, 100, now), // well within limit, not itself expired
	}

	_, ok := selectWinner(intent, quotes, now)
	if ok {
		t.Fatal("expected no match — the RfqIntent itself has already expired, even though the quote qualifies")
	}
}

func TestValidateRfqIntent_RejectsNilFields(t *testing.T) {
	incomplete := rfqIntent{Side: sideTakerBuy, Taker: common.HexToAddress("0x1"), Expiry: uint64(time.Now().Add(time.Hour).Unix())}
	if err := validateRfqIntent(incomplete); err == nil {
		t.Fatal("expected an error for nil Size/LimitPrice/RfqNonce, got nil")
	}
}

func TestValidateRfqIntent_RejectsNegativeValues(t *testing.T) {
	negative := rfqIntent{
		Side: sideTakerBuy, Size: big.NewInt(-1), LimitPrice: big.NewInt(100), RfqNonce: big.NewInt(1),
		Taker: common.HexToAddress("0x1"), Expiry: uint64(time.Now().Add(time.Hour).Unix()),
	}
	if err := validateRfqIntent(negative); err == nil {
		t.Fatal("expected an error for negative Size, got nil")
	}
}

func TestValidateQuote_RejectsNegativePrice(t *testing.T) {
	negative := quote{Price: big.NewInt(-1), Maker: common.HexToAddress("0x1"), Expiry: uint64(time.Now().Add(time.Hour).Unix())}
	if err := validateQuote(negative); err == nil {
		t.Fatal("expected an error for negative Price, got nil")
	}
}
