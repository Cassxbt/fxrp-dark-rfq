package extension

// Live integration test against the actual running Docker stack (ext-proxy +
// extension-tee), reached over the network via the ngrok-tunneled proxy URL.
// Reuses this package's real hashRfqIntent/hashQuote/eip712Config code
// directly instead of a separate tool reimplementing it — a duplicate
// implementation could silently drift from what's actually running.
//
// Guarded behind RFQ_INTEGRATION_TEST=1 so `go test ./...` stays hermetic for
// anyone (a judge included) who clones the repo without the stack running.
// Run with:
//
//	RFQ_INTEGRATION_TEST=1 EXT_PROXY_URL=https://<ngrok>.ngrok-free.app \
//	  RFQ_SETTLEMENT_ADDRESS=0x... CHAIN_ID=114 go test ./internal/extension/... -run TestRfqEndToEnd -v

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/json"
	"math/big"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	gethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/ecies"
	"github.com/ethereum/go-ethereum/ethclient"

	"fxrp-dark-rfq-extension/internal/extension/rfqcontract"

	dcrsecp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// Message/Data use hexutil.Bytes, not plain []byte — Go's default JSON
// encoding for []byte is base64, but the server (types.DirectInstruction /
// types.ActionResult) expects hex strings. Mismatched here first, caught by
// a 400 from the real server rather than a passing-but-wrong test.
type directInstruction struct {
	OPType    common.Hash   `json:"opType"`
	OPCommand common.Hash   `json:"opCommand"`
	Message   hexutil.Bytes `json:"message"`
}

type directResponse struct {
	Data struct {
		ID common.Hash `json:"id"`
	} `json:"data"`
}

type actionResultResponse struct {
	Result struct {
		Status uint8         `json:"status"`
		Log    string        `json:"log"`
		Data   hexutil.Bytes `json:"data"`
	} `json:"result"`
}

func opHash(s string) common.Hash {
	var h common.Hash
	copy(h[:], []byte(s))
	return h
}

func postDirect(t *testing.T, proxyURL, opType, opCommand string, message []byte) common.Hash {
	t.Helper()
	body, err := json.Marshal(directInstruction{OPType: opHash(opType), OPCommand: opHash(opCommand), Message: message})
	if err != nil {
		t.Fatalf("marshaling DirectInstruction: %v", err)
	}

	resp, err := http.Post(proxyURL+"/direct", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /direct: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /direct returned %d", resp.StatusCode)
	}

	var dr directResponse
	if err := json.NewDecoder(resp.Body).Decode(&dr); err != nil {
		t.Fatalf("decoding /direct response: %v", err)
	}
	return dr.Data.ID
}

func pollResult(t *testing.T, proxyURL string, actionID common.Hash) actionResultResponse {
	t.Helper()
	// Direct actions are enqueued with SubmissionTag "submit" (queue.DirectInstructionToAction),
	// but /action/result/{id} defaults to "threshold" if submissionTag isn't
	// passed explicitly — querying the default tag for a submit-tagged action
	// silently returns nothing, which looks identical to "not ready yet" and
	// was the actual cause of a timeout here even though the log showed the
	// action succeeding server-side almost instantly.
	var last actionResultResponse
	for i := 0; i < 15; i++ {
		resp, err := http.Get(proxyURL + "/action/result/" + actionID.Hex() + "?submissionTag=submit")
		if err == nil {
			// Close unconditionally on every non-nil response, not just the
			// happy-path branch below — a transient non-200 across 15 poll
			// iterations previously leaked one socket per bad response.
			if resp.StatusCode == http.StatusOK {
				decErr := json.NewDecoder(resp.Body).Decode(&last)
				resp.Body.Close()
				if decErr == nil && last.Result.Status != 0 {
					return last
				}
			} else {
				resp.Body.Close()
			}
		}
		time.Sleep(2 * time.Second)
	}
	t.Fatalf("timed out polling result for action %s (last log: %q)", actionID.Hex(), last.Result.Log)
	return last
}

func fetchTeePubkey(t *testing.T, proxyURL string) *ecies.PublicKey {
	t.Helper()
	resp, err := http.Get(proxyURL + "/info")
	if err != nil {
		t.Fatalf("GET /info: %v", err)
	}
	defer resp.Body.Close()

	// machineData.publicKey is {x, y} hex fields, not a serialized blob —
	// discovered by curling /info directly rather than assuming the shape.
	var info struct {
		MachineData struct {
			PublicKey struct {
				X string `json:"x"`
				Y string `json:"y"`
			} `json:"publicKey"`
		} `json:"machineData"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		t.Fatalf("decoding /info: %v", err)
	}

	x, ok := new(big.Int).SetString(trimHexPrefix(info.MachineData.PublicKey.X), 16)
	if !ok {
		t.Fatalf("parsing pubkey X %q as hex", info.MachineData.PublicKey.X)
	}
	y, ok := new(big.Int).SetString(trimHexPrefix(info.MachineData.PublicKey.Y), 16)
	if !ok {
		t.Fatalf("parsing pubkey Y %q as hex", info.MachineData.PublicKey.Y)
	}
	return &ecies.PublicKey{X: x, Y: y, Curve: ecies.DefaultCurve, Params: ecies.ECIES_AES128_SHA256}
}

func signEnvelope[T any](t *testing.T, data T, digest common.Hash, key *dcrsecp256k1.PrivateKey) []byte {
	t.Helper()
	sig, err := signDigest(key, digest.Bytes())
	if err != nil {
		t.Fatalf("signing envelope: %v", err)
	}
	b, err := json.Marshal(signedEnvelope[T]{Data: data, Signature: sig})
	if err != nil {
		t.Fatalf("marshaling envelope: %v", err)
	}
	return b
}

func encryptToTee(t *testing.T, pub *ecies.PublicKey, plaintext []byte) []byte {
	t.Helper()
	ct, err := ecies.Encrypt(rand.Reader, pub, plaintext, nil, nil)
	if err != nil {
		t.Fatalf("ECIES encrypt: %v", err)
	}
	return ct
}

func TestRfqEndToEnd(t *testing.T) {
	if os.Getenv("RFQ_INTEGRATION_TEST") != "1" {
		t.Skip("set RFQ_INTEGRATION_TEST=1 to run this against a live stack")
	}
	proxyURL := os.Getenv("EXT_PROXY_URL")
	if proxyURL == "" {
		t.Fatal("EXT_PROXY_URL must be set")
	}

	eip712, err := loadEip712Config()
	if err != nil {
		t.Fatalf("loadEip712Config (need CHAIN_ID and RFQ_SETTLEMENT_ADDRESS set): %v", err)
	}

	pubkey := fetchTeePubkey(t, proxyURL)

	// Prerequisite, not automated here: the attested signer key must already be
	// loaded via the on-chain InstructionSender path (e.g. `go run
	// ./tools/cmd/run-test`, or any KEY/UPDATE sent as a real transaction), then
	// its address whitelisted on-chain via setAttestedSigner.
	//
	// KEY/UPDATE is deliberately NOT dispatchable here via /direct — that was
	// wrong in an earlier version of this test: /direct has no
	// auth at this layer, so loading arbitrary keys through it would let anyone
	// who can reach the tunnel overwrite the extension's signing key. extension.go's
	// processAction now hard-rejects any OPType other than RFQ on the Direct path;
	// this test respects that boundary rather than routing around it.
	// If CLOSE below fails with "no attested signer key stored," that means this
	// prerequisite wasn't met — set it up first, this test won't do it for you.

	// Step 1: open an RFQ as a fresh taker.
	takerKeyBytes := make([]byte, 32)
	takerKeyBytes[31] = 0x11
	takerKey := dcrsecp256k1.PrivKeyFromBytes(takerKeyBytes)
	takerAddr := gethcrypto.PubkeyToAddress(*mustPubkey(t, takerKey))

	intent := rfqIntent{
		Side:       sideTakerBuy,
		Size:       big.NewInt(1_000_000),                 // 1 FXRP (6 decimals)
		LimitPrice: big.NewInt(3_000_000_000_000_000_000), // generous limit, 3.00 WAD
		Taker:      takerAddr,
		Expiry:     uint64(time.Now().Add(2 * time.Minute).Unix()),
		RfqNonce:   big.NewInt(time.Now().UnixNano()), // unique per run
	}
	intentDigest := eip712.typedDataDigest(hashRfqIntent(intent))
	openCiphertext := encryptToTee(t, pubkey, signEnvelope(t, intent, intentDigest, takerKey))

	openID := postDirect(t, proxyURL, "RFQ", "OPEN", openCiphertext)
	openResult := pollResult(t, proxyURL, openID)
	if openResult.Result.Status != 1 {
		t.Fatalf("RFQ/OPEN failed: %s", openResult.Result.Log)
	}

	var openResp struct {
		RfqID string `json:"rfqId"`
	}
	if err := json.Unmarshal(openResult.Result.Data, &openResp); err != nil {
		t.Fatalf("decoding RFQ/OPEN response %q: %v", string(openResult.Result.Data), err)
	}
	rfqID := common.HexToHash(openResp.RfqID)
	t.Logf("RFQ opened: %s", rfqID.Hex())

	// Step 2: two makers quote — proves the TEE picks a winner, not a 1:1 relay.
	makerAKeyBytes := make([]byte, 32)
	makerAKeyBytes[31] = 0x22
	makerAKey := dcrsecp256k1.PrivKeyFromBytes(makerAKeyBytes)
	makerAAddr := gethcrypto.PubkeyToAddress(*mustPubkey(t, makerAKey))

	makerBKeyBytes := make([]byte, 32)
	makerBKeyBytes[31] = 0x33
	makerBKey := dcrsecp256k1.PrivKeyFromBytes(makerBKeyBytes)
	makerBAddr := gethcrypto.PubkeyToAddress(*mustPubkey(t, makerBKey))

	submitQuote := func(maker common.Address, key *dcrsecp256k1.PrivateKey, price int64) {
		q := quote{RfqID: rfqID, Price: big.NewInt(price), Maker: maker, Expiry: uint64(time.Now().Add(2 * time.Minute).Unix())}
		digest := eip712.typedDataDigest(hashQuote(q))
		ciphertext := encryptToTee(t, pubkey, signEnvelope(t, q, digest, key))
		id := postDirect(t, proxyURL, "RFQ", "QUOTE", ciphertext)
		result := pollResult(t, proxyURL, id)
		if result.Result.Status != 1 {
			t.Fatalf("RFQ/QUOTE from %s failed: %s", maker.Hex(), result.Result.Log)
		}
	}
	submitQuote(makerAAddr, makerAKey, 2_500_000_000_000_000_000) // 2.50 — worse
	submitQuote(makerBAddr, makerBKey, 2_000_000_000_000_000_000) // 2.00 — better, should win
	t.Log("both maker quotes accepted")

	// Step 3: close and expect makerB (lower price on a buy) to win.
	closeID := postDirect(t, proxyURL, "RFQ", "CLOSE", rfqID.Bytes())
	closeResult := pollResult(t, proxyURL, closeID)
	if closeResult.Result.Status != 1 {
		t.Fatalf("RFQ/CLOSE failed: %s", closeResult.Result.Log)
	}

	// Response is intentionally synchronous-match/async-settle (see rfq.go's
	// processRfqClose comment) — no txHash here, the framework's 2-second
	// per-action timeout doesn't fit a chain round trip.
	var closeResp struct {
		Matched bool   `json:"matched"`
		Maker   string `json:"maker"`
		Price   string `json:"price"`
		Note    string `json:"note"`
	}
	if err := json.Unmarshal(closeResult.Result.Data, &closeResp); err != nil {
		t.Fatalf("decoding RFQ/CLOSE response %q: %v", string(closeResult.Result.Data), err)
	}
	t.Logf("RFQ/CLOSE result: %+v", closeResp)

	if !closeResp.Matched {
		t.Fatal("expected a match — two qualifying quotes were submitted")
	}
	if common.HexToAddress(closeResp.Maker) != makerBAddr {
		t.Fatalf("expected the lower-price maker (%s) to win a buy, got %s", makerBAddr.Hex(), closeResp.Maker)
	}
	t.Log("winning maker correctly selected (lowest price on a buy) — this is the wire path and matching logic proven end-to-end through the real /direct endpoint")

	// Informational only: poll the contract's settled() mapping to see whether
	// the async settle tx actually landed. It's expected to revert here —
	// the test taker/maker are throwaway keys with no FXRP/USDT0 balance or
	// allowance, a separate concern from whether OPEN/QUOTE/CLOSE and the
	// signature/matching pipeline work, which is what this test verifies.
	time.Sleep(6 * time.Second)
	settled := checkSettled(t, rfqID)
	t.Logf("on-chain settled(%s) = %v (false is expected/fine without funded test wallets)", rfqID.Hex(), settled)
}

func checkSettled(t *testing.T, rfqID common.Hash) bool {
	t.Helper()
	contractAddr := os.Getenv("RFQ_SETTLEMENT_ADDRESS")
	chainURL := os.Getenv("CHAIN_URL")
	if chainURL == "" {
		chainURL = "https://coston2-api.flare.network/ext/C/rpc"
	}
	client, err := ethclient.Dial(chainURL)
	if err != nil {
		t.Logf("could not dial chain to check settled(): %v", err)
		return false
	}
	defer client.Close()
	contract, err := rfqcontract.NewRfqSettlement(common.HexToAddress(contractAddr), client)
	if err != nil {
		t.Logf("could not bind contract to check settled(): %v", err)
		return false
	}
	settled, err := contract.Settled(nil, rfqID)
	if err != nil {
		t.Logf("could not call settled(): %v", err)
		return false
	}
	return settled
}

func mustPubkey(t *testing.T, key *dcrsecp256k1.PrivateKey) *ecdsa.PublicKey {
	t.Helper()
	pub, err := gethcrypto.UnmarshalPubkey(key.PubKey().SerializeUncompressed())
	if err != nil {
		t.Fatalf("unmarshaling pubkey: %v", err)
	}
	return pub
}
