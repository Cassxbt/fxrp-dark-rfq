package extension

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"sort"
	"sync"
	"time"

	"sign-extension/internal/extension/rfqcontract"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/flare-foundation/go-flare-common/pkg/logger"
	"github.com/flare-foundation/go-flare-common/pkg/tee/instruction"
	teetypes "github.com/flare-foundation/tee-node/pkg/types"

	dcrsecp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// side mirrors RfqSettlement.sol's Side enum — must stay numerically aligned.
type side uint8

const (
	sideTakerBuy  side = 0
	sideTakerSell side = 1
)

// rfqIntent mirrors the taker's signed EIP-712 RfqIntent, per BUILD-SPEC.md §2.1/§2.2.
// Field order and types must match the Solidity/JS encodings exactly, or the
// recovered signer will not match what the client believes it signed.
type rfqIntent struct {
	Side       side           `json:"side"`
	Size       *big.Int       `json:"size"`
	LimitPrice *big.Int       `json:"limitPrice"`
	Taker      common.Address `json:"taker"`
	Expiry     uint64         `json:"expiry"`
	RfqNonce   *big.Int       `json:"rfqNonce"`
}

// quote mirrors the maker's signed EIP-712 Quote. RfqID is bound inside the
// signed struct itself (round-3 fix) so a captured quote can't be replayed
// onto a different RFQ.
type quote struct {
	RfqID  common.Hash    `json:"rfqId"`
	Price  *big.Int       `json:"price"`
	Maker  common.Address `json:"maker"`
	Expiry uint64         `json:"expiry"`
}

// signedEnvelope is the plaintext structure recovered after decrypting a
// /direct RFQ payload: the typed struct plus its 65-byte [r,s,v] signature.
type signedEnvelope[T any] struct {
	Data      T      `json:"data"`
	Signature []byte `json:"signature"`
}

// rfqState is the extension's in-memory record of one open RFQ. Deliberately
// not persisted — a restart wipes the book, disclosed in BUILD-SPEC.md §2.2.
type rfqState struct {
	ID       common.Hash
	Intent   rfqIntent
	Quotes   map[common.Address]quoteEntry // at most one live quote per maker
	OpenedAt time.Time
}

type quoteEntry struct {
	Quote      quote
	ReceivedAt time.Time
}

// rfqBook holds all state for the RFQ extension: open RFQs, and a nonce-reuse
// guard that's independent of the RFQ map so a replayed intent is rejected
// even after the RFQ it originally opened has already closed and been
// forgotten (round-3 rfqNonce fix).
//
// usedNonces maps nonce -> the expiry of the intent that used it, not just a
// bool. An intent can never succeed again once its own Expiry has passed
// (processRfqOpen already rejects expired intents independently), so once
// that time is reached the entry is safe to prune — this is what bounds the
// map's growth (code-review finding: it was unbounded before).
type rfqBook struct {
	mu         sync.Mutex
	open       map[common.Hash]*rfqState
	usedNonces map[common.Address]map[string]uint64 // taker -> nonce.String() -> intent expiry
}

func newRfqBook() *rfqBook {
	return &rfqBook{
		open:       make(map[common.Hash]*rfqState),
		usedNonces: make(map[common.Address]map[string]uint64),
	}
}

// pruneExpired evicts open RFQs and used-nonce records whose expiry has
// passed. Called opportunistically on open/close rather than on a timer —
// simplest fix that bounds memory growth without adding a background
// goroutine (code-review finding on both b.open and b.usedNonces). Caller
// must hold b.mu.
func (b *rfqBook) pruneExpired(now uint64) {
	for id, rfq := range b.open {
		if now >= rfq.Intent.Expiry {
			delete(b.open, id)
		}
	}
	for taker, nonces := range b.usedNonces {
		for nonce, expiry := range nonces {
			if now >= expiry {
				delete(nonces, nonce)
			}
		}
		if len(nonces) == 0 {
			delete(b.usedNonces, taker)
		}
	}
}

// validateRfqIntent rejects a decoded-but-unchecked RfqIntent before it's
// hashed or stored. Required: json.Unmarshal leaves *big.Int fields nil when
// the client omits them, and .Bytes() on a nil *big.Int panics (code-review
// finding) — reject cleanly here instead of crashing the request goroutine.
// Negative values are also rejected: big.Int.Bytes() encodes only the
// magnitude, but the on-chain Fill encodes the same field via two's-complement
// ABI packing — a negative value would hash differently in Go than what
// actually gets ABI-encoded on-chain, silently breaking the signature
// (code-review finding).
func validateRfqIntent(i rfqIntent) error {
	if i.Size == nil || i.LimitPrice == nil || i.RfqNonce == nil {
		return fmt.Errorf("size, limitPrice, and rfqNonce must all be present")
	}
	if i.Size.Sign() < 0 || i.LimitPrice.Sign() < 0 || i.RfqNonce.Sign() < 0 {
		return fmt.Errorf("size, limitPrice, and rfqNonce must be non-negative")
	}
	if i.Side != sideTakerBuy && i.Side != sideTakerSell {
		return fmt.Errorf("side must be 0 (buy) or 1 (sell), got %d", i.Side)
	}
	return nil
}

// validateQuote is validateRfqIntent's counterpart for Quote — same rationale.
func validateQuote(q quote) error {
	if q.Price == nil {
		return fmt.Errorf("price must be present")
	}
	if q.Price.Sign() < 0 {
		return fmt.Errorf("price must be non-negative")
	}
	return nil
}

// EIP-712 domain and type hashes, matching RfqSettlement.sol exactly.
//
// contracts/src/RfqSettlement.sol:
//   EIP712("RfqSettlement", "1")
//   FILL_TYPEHASH = keccak256("Fill(bytes32 rfqId,address taker,address maker,uint8 side,uint256 size,uint256 price,uint256 expiry)")
//
// RfqIntent and Quote are verified only here in the extension (not re-checked
// on-chain — the disclosed trust-model choice in BUILD-SPEC.md §2.1), but they
// use the *same* domain as Fill so all three are bound to this contract and
// can't be replayed against a different one.

const (
	eip712DomainTypehash = "EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"
	rfqIntentTypehash    = "RfqIntent(uint8 side,uint256 size,uint256 limitPrice,address taker,uint256 expiry,uint256 rfqNonce)"
	quoteTypehash        = "Quote(bytes32 rfqId,uint256 price,address maker,uint256 expiry)"
	fillTypehash         = "Fill(bytes32 rfqId,address taker,address maker,uint8 side,uint256 size,uint256 price,uint256 expiry)"
)

// eip712Config is resolved once at startup from the environment — the
// verifying contract address changes per deploy, so it must never be hardcoded.
type eip712Config struct {
	chainID           *big.Int
	verifyingContract common.Address
}

func loadEip712Config() (eip712Config, error) {
	chainIDStr := os.Getenv("CHAIN_ID")
	if chainIDStr == "" {
		return eip712Config{}, fmt.Errorf("CHAIN_ID not set")
	}
	chainID, ok := new(big.Int).SetString(chainIDStr, 10)
	if !ok {
		return eip712Config{}, fmt.Errorf("invalid CHAIN_ID: %s", chainIDStr)
	}

	addrStr := os.Getenv("RFQ_SETTLEMENT_ADDRESS")
	if addrStr == "" {
		return eip712Config{}, fmt.Errorf("RFQ_SETTLEMENT_ADDRESS not set")
	}

	return eip712Config{
		chainID:           chainID,
		verifyingContract: common.HexToAddress(addrStr),
	}, nil
}

func (c eip712Config) domainSeparator() common.Hash {
	nameHash := keccak256([]byte("RfqSettlement"))
	versionHash := keccak256([]byte("1"))

	buf := make([]byte, 0, 32*5)
	buf = append(buf, keccak256([]byte(eip712DomainTypehash))...)
	buf = append(buf, nameHash...)
	buf = append(buf, versionHash...)
	buf = append(buf, padLeft(c.chainID.Bytes(), 32)...)
	buf = append(buf, padLeft(c.verifyingContract.Bytes(), 32)...)

	return common.BytesToHash(keccak256(buf))
}

// typedDataDigest computes the final EIP-712 digest: keccak256("\x19\x01" || domainSeparator || structHash).
func (c eip712Config) typedDataDigest(structHash []byte) common.Hash {
	domainSep := c.domainSeparator()
	buf := make([]byte, 0, 2+32+32)
	buf = append(buf, 0x19, 0x01)
	buf = append(buf, domainSep.Bytes()...)
	buf = append(buf, structHash...)
	return common.BytesToHash(keccak256(buf))
}

func hashRfqIntent(i rfqIntent) []byte {
	buf := make([]byte, 0, 32*7)
	buf = append(buf, keccak256([]byte(rfqIntentTypehash))...)
	buf = append(buf, padLeft(big.NewInt(int64(i.Side)).Bytes(), 32)...)
	buf = append(buf, padLeft(i.Size.Bytes(), 32)...)
	buf = append(buf, padLeft(i.LimitPrice.Bytes(), 32)...)
	buf = append(buf, padLeft(i.Taker.Bytes(), 32)...)
	buf = append(buf, padLeft(new(big.Int).SetUint64(i.Expiry).Bytes(), 32)...)
	buf = append(buf, padLeft(i.RfqNonce.Bytes(), 32)...)
	return keccak256(buf)
}

func hashQuote(q quote) []byte {
	buf := make([]byte, 0, 32*5)
	buf = append(buf, keccak256([]byte(quoteTypehash))...)
	buf = append(buf, q.RfqID.Bytes()...)
	buf = append(buf, padLeft(q.Price.Bytes(), 32)...)
	buf = append(buf, padLeft(q.Maker.Bytes(), 32)...)
	buf = append(buf, padLeft(new(big.Int).SetUint64(q.Expiry).Bytes(), 32)...)
	return keccak256(buf)
}

func hashFillForSigning(rfqID common.Hash, taker, maker common.Address, s side, size, price *big.Int, expiry uint64) []byte {
	buf := make([]byte, 0, 32*8)
	buf = append(buf, keccak256([]byte(fillTypehash))...)
	buf = append(buf, rfqID.Bytes()...)
	buf = append(buf, padLeft(taker.Bytes(), 32)...)
	buf = append(buf, padLeft(maker.Bytes(), 32)...)
	buf = append(buf, padLeft(big.NewInt(int64(s)).Bytes(), 32)...)
	buf = append(buf, padLeft(size.Bytes(), 32)...)
	buf = append(buf, padLeft(price.Bytes(), 32)...)
	buf = append(buf, padLeft(new(big.Int).SetUint64(expiry).Bytes(), 32)...)
	return keccak256(buf)
}

// recoverSigner recovers the address that produced a 65-byte [r,s,v] signature
// (v = 27/28) over the given EIP-712 digest. Never trust a client-supplied
// address field instead of this — the round-2 critical fix this whole file
// exists to implement correctly.
func recoverSigner(digest common.Hash, sig []byte) (common.Address, error) {
	if len(sig) != 65 {
		return common.Address{}, fmt.Errorf("signature must be 65 bytes, got %d", len(sig))
	}
	normalized := make([]byte, 65)
	copy(normalized, sig)
	if normalized[64] >= 27 {
		normalized[64] -= 27
	}

	pubkey, err := crypto.SigToPub(digest.Bytes(), normalized)
	if err != nil {
		return common.Address{}, fmt.Errorf("recovering pubkey: %w", err)
	}
	return crypto.PubkeyToAddress(*pubkey), nil
}

// Handlers below are dispatched from processAction in extension.go.

func (e *Extension) processRfqOpen(action teetypes.Action, df *instruction.DataFixed) teetypes.ActionResult {
	if len(df.OriginalMessage) == 0 {
		return buildResult(action, df, nil, 0, fmt.Errorf("originalMessage is empty"))
	}

	plaintext, err := decryptViaNode(e.signPort, df.OriginalMessage)
	if err != nil {
		return buildResult(action, df, nil, 0, fmt.Errorf("decrypting RfqIntent: %w", err))
	}

	var env signedEnvelope[rfqIntent]
	if err := json.Unmarshal(plaintext, &env); err != nil {
		return buildResult(action, df, nil, 0, fmt.Errorf("parsing RfqIntent envelope: %w", err))
	}
	if err := validateRfqIntent(env.Data); err != nil {
		return buildResult(action, df, nil, 0, fmt.Errorf("invalid RfqIntent: %w", err))
	}

	eip712, err := loadEip712Config()
	if err != nil {
		return buildResult(action, df, nil, 0, fmt.Errorf("loading EIP-712 config: %w", err))
	}

	digest := eip712.typedDataDigest(hashRfqIntent(env.Data))
	recovered, err := recoverSigner(digest, env.Signature)
	if err != nil {
		return buildResult(action, df, nil, 0, fmt.Errorf("recovering taker signature: %w", err))
	}
	if recovered != env.Data.Taker {
		return buildResult(action, df, nil, 0, fmt.Errorf(
			"signature does not match claimed taker: recovered %s, claimed %s", recovered.Hex(), env.Data.Taker.Hex()))
	}
	if env.Data.Expiry <= uint64(time.Now().Unix()) {
		return buildResult(action, df, nil, 0, fmt.Errorf("RfqIntent already expired"))
	}

	rfqID, err := e.book.openRfq(eip712, env.Data)
	if err != nil {
		return buildResult(action, df, nil, 0, err)
	}

	respData, err := json.Marshal(map[string]any{
		"rfqId": rfqID.Hex(),
		"side":  env.Data.Side,
		"size":  env.Data.Size.String(),
	})
	if err != nil {
		return buildResult(action, df, nil, 0, fmt.Errorf("marshaling response: %w", err))
	}
	logger.Infof("RFQ opened: %s (taker %s)", rfqID.Hex(), env.Data.Taker.Hex())
	return buildResult(action, df, respData, 1, nil)
}

// openRfq validates the nonce hasn't been used, computes the deterministic
// rfqId (round-3 fix: keccak256(taker, nonce, verifyingContract), not random —
// so the on-chain settled[] guard still catches a replay after an in-memory
// restart), and stores the RFQ.
func (b *rfqBook) openRfq(eip712 eip712Config, intent rfqIntent) (common.Hash, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := uint64(time.Now().Unix())
	b.pruneExpired(now)

	nonceKey := intent.RfqNonce.String()
	if _, used := b.usedNonces[intent.Taker][nonceKey]; used {
		return common.Hash{}, fmt.Errorf("rfqNonce %s already used by %s", nonceKey, intent.Taker.Hex())
	}

	rfqID := common.BytesToHash(keccak256(append(append(
		intent.Taker.Bytes(),
		padLeft(intent.RfqNonce.Bytes(), 32)...),
		eip712.verifyingContract.Bytes()...)))

	if b.usedNonces[intent.Taker] == nil {
		b.usedNonces[intent.Taker] = make(map[string]uint64)
	}
	b.usedNonces[intent.Taker][nonceKey] = intent.Expiry

	b.open[rfqID] = &rfqState{
		ID:       rfqID,
		Intent:   intent,
		Quotes:   make(map[common.Address]quoteEntry),
		OpenedAt: time.Now(),
	}
	return rfqID, nil
}

func (e *Extension) processRfqQuote(action teetypes.Action, df *instruction.DataFixed) teetypes.ActionResult {
	if len(df.OriginalMessage) == 0 {
		return buildResult(action, df, nil, 0, fmt.Errorf("originalMessage is empty"))
	}

	plaintext, err := decryptViaNode(e.signPort, df.OriginalMessage)
	if err != nil {
		return buildResult(action, df, nil, 0, fmt.Errorf("decrypting Quote: %w", err))
	}

	var env signedEnvelope[quote]
	if err := json.Unmarshal(plaintext, &env); err != nil {
		return buildResult(action, df, nil, 0, fmt.Errorf("parsing Quote envelope: %w", err))
	}
	if err := validateQuote(env.Data); err != nil {
		return buildResult(action, df, nil, 0, fmt.Errorf("invalid Quote: %w", err))
	}

	eip712, err := loadEip712Config()
	if err != nil {
		return buildResult(action, df, nil, 0, fmt.Errorf("loading EIP-712 config: %w", err))
	}

	digest := eip712.typedDataDigest(hashQuote(env.Data))
	recovered, err := recoverSigner(digest, env.Signature)
	if err != nil {
		return buildResult(action, df, nil, 0, fmt.Errorf("recovering maker signature: %w", err))
	}
	if recovered != env.Data.Maker {
		return buildResult(action, df, nil, 0, fmt.Errorf(
			"signature does not match claimed maker: recovered %s, claimed %s", recovered.Hex(), env.Data.Maker.Hex()))
	}
	if env.Data.Expiry <= uint64(time.Now().Unix()) {
		return buildResult(action, df, nil, 0, fmt.Errorf("Quote already expired"))
	}

	if err := e.book.submitQuote(env.Data); err != nil {
		return buildResult(action, df, nil, 0, err)
	}

	logger.Infof("Quote received for RFQ %s from maker %s", env.Data.RfqID.Hex(), env.Data.Maker.Hex())
	return buildResult(action, df, nil, 1, nil)
}

func (b *rfqBook) submitQuote(q quote) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	rfq, exists := b.open[q.RfqID]
	if !exists {
		return fmt.Errorf("no open RFQ with id %s", q.RfqID.Hex())
	}
	if uint64(time.Now().Unix()) >= rfq.Intent.Expiry {
		return fmt.Errorf("RFQ %s has expired", q.RfqID.Hex())
	}

	// At most one live quote per maker — a resubmission replaces the earlier one.
	rfq.Quotes[q.Maker] = quoteEntry{Quote: q, ReceivedAt: time.Now()}
	return nil
}

// processRfqClose runs winner selection and, on a match, builds, signs, and
// submits the settlement transaction. Triggered by the taker (or a client-side
// timer) — see BUILD-SPEC.md §2.2's six-step winner-selection spec, implemented
// exactly in selectWinner below.
func (e *Extension) processRfqClose(action teetypes.Action, df *instruction.DataFixed) teetypes.ActionResult {
	if len(df.OriginalMessage) != 32 {
		return buildResult(action, df, nil, 0, fmt.Errorf("expected a 32-byte rfqId, got %d bytes", len(df.OriginalMessage)))
	}
	rfqID := common.BytesToHash(df.OriginalMessage)

	rfq, winner, ok := e.book.closeAndSelectWinner(rfqID)
	if !ok {
		return buildResult(action, df, []byte(`{"matched":false}`), 1, nil)
	}

	e.mu.RLock()
	signerKey := e.privateKey
	e.mu.RUnlock()
	if signerKey == nil {
		return buildResult(action, df, nil, 0, fmt.Errorf("no attested signer key stored — send an UPDATE instruction first"))
	}

	eip712, err := loadEip712Config()
	if err != nil {
		return buildResult(action, df, nil, 0, fmt.Errorf("loading EIP-712 config: %w", err))
	}

	fillExpiry := minUint64(rfq.Intent.Expiry, winner.Quote.Expiry, uint64(time.Now().Unix())+120)

	fillDigest := eip712.typedDataDigest(hashFillForSigning(
		rfqID, rfq.Intent.Taker, winner.Quote.Maker, rfq.Intent.Side,
		rfq.Intent.Size, winner.Quote.Price, fillExpiry,
	))

	attestationSig, err := signAttestation(signerKey, fillDigest)
	if err != nil {
		return buildResult(action, df, nil, 0, fmt.Errorf("signing Fill: %w", err))
	}

	// Submitted asynchronously, not awaited here — discovered live, not by
	// inspection: the framework's own action-response cycle has a hardcoded,
	// non-configurable 2-second timeout (tee-node internal/settings.ProxyTimeout),
	// and dialing the chain, fetching the chain ID, and broadcasting a tx
	// reliably takes longer than that. The correct place to confirm settlement
	// is the on-chain Filled event, not this HTTP response anyway — a frontend
	// trusting a synchronous "settled" claim from the extension rather than
	// watching the chain would be trusting the TEE's word over the ledger,
	// which is backwards for a system whose whole point is on-chain
	// verifiability. This handler reports a match; the chain reports settlement.
	go func() {
		txHash, err := submitSettle(rfqID, rfq.Intent, winner.Quote, fillExpiry, attestationSig)
		if err != nil {
			logger.Errorf("RFQ %s matched but settle submission failed: %v", rfqID.Hex(), err)
			return
		}
		logger.Infof("RFQ %s matched: maker %s, settle tx %s (confirm via the Filled event, not this log)", rfqID.Hex(), winner.Quote.Maker.Hex(), txHash)
	}()

	respData, _ := json.Marshal(map[string]any{
		"matched": true,
		"maker":   winner.Quote.Maker.Hex(),
		"price":   winner.Quote.Price.String(),
		"note":    "settlement submitted asynchronously — watch the RfqSettlement contract's Filled event on-chain to confirm, do not trust this response alone",
	})
	logger.Infof("RFQ %s matched: maker %s, settlement submission in progress", rfqID.Hex(), winner.Quote.Maker.Hex())
	return buildResult(action, df, respData, 1, nil)
}

// closeAndSelectWinner removes the RFQ from the open book (it's terminal
// either way — matched or not, we don't leave it open for a second close
// attempt) and returns the winning quote, if any.
func (b *rfqBook) closeAndSelectWinner(rfqID common.Hash) (rfqState, quoteEntry, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	rfq, exists := b.open[rfqID]
	if !exists {
		return rfqState{}, quoteEntry{}, false
	}
	delete(b.open, rfqID)

	winner, ok := selectWinner(rfq.Intent, rfq.Quotes, time.Now())
	if !ok {
		return *rfq, quoteEntry{}, false
	}
	return *rfq, winner, true
}

// selectWinner implements BUILD-SPEC.md §2.2's six-step spec exactly:
//  1. Drop expired/malformed quotes (malformed ones never made it into the map
//     at all — processRfqQuote already rejected them before storage)
//  2. At most one live quote per maker (already enforced by map key)
//  3. TAKER_BUY: qualifying quotes have price <= limitPrice; winner is lowest
//  4. TAKER_SELL: qualifying quotes have price >= limitPrice; winner is highest
//  5. Ties broken by earliest-received quote
//  6. No qualifying quote -> no match
func selectWinner(intent rfqIntent, quotes map[common.Address]quoteEntry, now time.Time) (quoteEntry, bool) {
	// The RFQ itself can have expired even if some individual quote hasn't
	// (code-review finding: this was previously only checked per-quote, never
	// against the taker's own intent — a late CLOSE could still pick a winner
	// and sign+submit a Fill guaranteed to revert on-chain, burning gas for
	// nothing since the tx had already been reported as matched).
	if uint64(now.Unix()) >= intent.Expiry {
		return quoteEntry{}, false
	}

	var qualifying []quoteEntry
	for _, q := range quotes {
		if uint64(now.Unix()) >= q.Quote.Expiry {
			continue
		}
		switch intent.Side {
		case sideTakerBuy:
			if q.Quote.Price.Cmp(intent.LimitPrice) <= 0 {
				qualifying = append(qualifying, q)
			}
		case sideTakerSell:
			if q.Quote.Price.Cmp(intent.LimitPrice) >= 0 {
				qualifying = append(qualifying, q)
			}
		}
	}
	if len(qualifying) == 0 {
		return quoteEntry{}, false
	}

	sort.Slice(qualifying, func(i, j int) bool {
		cmp := qualifying[i].Quote.Price.Cmp(qualifying[j].Quote.Price)
		if cmp == 0 {
			return qualifying[i].ReceivedAt.Before(qualifying[j].ReceivedAt)
		}
		if intent.Side == sideTakerBuy {
			return cmp < 0 // lowest price wins
		}
		return cmp > 0 // highest price wins
	})
	return qualifying[0], true
}

func minUint64(a, b, c uint64) uint64 {
	m := a
	if b < m {
		m = b
	}
	if c < m {
		m = c
	}
	return m
}

// signAttestation signs an EIP-712 digest directly (no Keccak-of-message step —
// the digest already is the EIP-712 hash) with the extension's stored key,
// returning a 65-byte [r,s,v] signature matching what RfqSettlement.sol's
// ECDSA.recover expects.
func signAttestation(key *dcrsecp256k1.PrivateKey, digest common.Hash) ([]byte, error) {
	return signDigest(key, digest.Bytes())
}

// settler caches the chain client, contract binding, and hot key across
// calls instead of redialing and re-fetching the chain ID on every single
// RFQ close (code-review finding). sendMu serializes the fetch-nonce-then-send
// sequence: bind.NewKeyedTransactorWithChainID pulls a pending nonce from the
// client internally, and two concurrent Settle calls racing that step could
// be assigned the same nonce (code-review finding — actionHandler has no
// serialization of its own beyond an assumption documented, not enforced, in
// extension.go). This mutex is what actually enforces it for this code path.
type settler struct {
	client   *ethclient.Client
	contract *rfqcontract.RfqSettlement
	hotKey   *ecdsa.PrivateKey
	chainID  *big.Int
	sendMu   sync.Mutex
}

var (
	settlerOnce     sync.Once
	settlerInstance *settler
	settlerInitErr  error
)

func loadSettler() (*settler, error) {
	settlerOnce.Do(func() {
		rpcURL := os.Getenv("CHAIN_URL")
		hotKeyHex := os.Getenv("RFQ_HOT_KEY")
		contractAddr := os.Getenv("RFQ_SETTLEMENT_ADDRESS")
		if rpcURL == "" || hotKeyHex == "" || contractAddr == "" {
			settlerInitErr = fmt.Errorf("CHAIN_URL, RFQ_HOT_KEY, and RFQ_SETTLEMENT_ADDRESS must all be set")
			return
		}

		client, err := ethclient.Dial(rpcURL)
		if err != nil {
			settlerInitErr = fmt.Errorf("connecting to chain: %w", err)
			return
		}

		hotKey, err := crypto.HexToECDSA(trimHexPrefix(hotKeyHex))
		if err != nil {
			settlerInitErr = fmt.Errorf("parsing RFQ_HOT_KEY: %w", err)
			return
		}

		chainID, err := client.ChainID(context.Background())
		if err != nil {
			settlerInitErr = fmt.Errorf("fetching chain ID: %w", err)
			return
		}

		contract, err := rfqcontract.NewRfqSettlement(common.HexToAddress(contractAddr), client)
		if err != nil {
			settlerInitErr = fmt.Errorf("binding contract: %w", err)
			return
		}

		settlerInstance = &settler{client: client, contract: contract, hotKey: hotKey, chainID: chainID}
	})
	return settlerInstance, settlerInitErr
}

// submitSettle builds the on-chain Fill transaction and submits it using the
// extension's gas-paying hot key (RFQ_HOT_KEY env var) — a distinct key from
// the attested signer per BUILD-SPEC.md §2.1's "three keys" note: msg.sender
// need not be the attested signer, the contract only checks the signature.
func submitSettle(rfqID common.Hash, intent rfqIntent, winningQuote quote, expiry uint64, attestationSig []byte) (string, error) {
	s, err := loadSettler()
	if err != nil {
		return "", err
	}

	fill := rfqcontract.RfqSettlementFill{
		RfqId:  rfqID,
		Taker:  intent.Taker,
		Maker:  winningQuote.Maker,
		Side:   uint8(intent.Side),
		Size:   intent.Size,
		Price:  winningQuote.Price,
		Expiry: new(big.Int).SetUint64(expiry),
	}

	s.sendMu.Lock()
	defer s.sendMu.Unlock()

	auth, err := bind.NewKeyedTransactorWithChainID(s.hotKey, s.chainID)
	if err != nil {
		return "", fmt.Errorf("building transactor: %w", err)
	}

	tx, err := s.contract.Settle(auth, fill, attestationSig)
	if err != nil {
		return "", fmt.Errorf("sending settle tx: %w", err)
	}
	return tx.Hash().Hex(), nil
}

func trimHexPrefix(s string) string {
	if len(s) >= 2 && s[0] == '0' && (s[1] == 'x' || s[1] == 'X') {
		return s[2:]
	}
	return s
}
