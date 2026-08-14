package extension

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"sign-extension/internal/config"
	"sign-extension/pkg/types"

	"github.com/flare-foundation/go-flare-common/pkg/tee/instruction"
	teetypes "github.com/flare-foundation/tee-node/pkg/types"
	teeutils "github.com/flare-foundation/tee-node/pkg/utils"

	"github.com/flare-foundation/tee-node/pkg/processorutils"

	secp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// Extension holds mutable state for the sign extension. Access is serialized
// by the mutex; the framework dispatches actions serially anyway, but the
// state read in stateHandler is concurrent with action processing.
type Extension struct {
	mu     sync.RWMutex
	Server *http.Server

	// signPort is the TEE node's /decrypt endpoint port, used by handleKeyUpdate.
	signPort int

	// privateKey is the secp256k1 private key delivered via UPDATE_KEY. May be nil
	// before the first successful UPDATE_KEY instruction. Doubles as the RFQ
	// attested signer key — distinct from the extension's separate gas-paying
	// hot key (RFQ_HOT_KEY) and the ECIES encryption key.
	privateKey *secp256k1.PrivateKey

	// book holds in-memory RFQ/quote state. Not persisted — a restart wipes it.
	book *rfqBook
}

// --- DO NOT MODIFY: New(), actionHandler() are boilerplate.
func New(extensionPort, signPort int) *Extension {
	e := &Extension{signPort: signPort, book: newRfqBook()}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /state", e.stateHandler)
	mux.HandleFunc("POST /action", e.actionHandler)

	e.Server = &http.Server{Addr: fmt.Sprintf(":%d", extensionPort), Handler: mux}
	return e
}

// stateHandler reports whether a key is stored, without exposing the key.
func (e *Extension) stateHandler(w http.ResponseWriter, r *http.Request) {
	e.mu.RLock()
	stateResponse := types.StateResponse{
		StateVersion: teeutils.ToHash(config.Version),
		State: types.State{
			HasKey: e.privateKey != nil,
		},
	}
	e.mu.RUnlock()

	err := json.NewEncoder(w).Encode(stateResponse)
	if err != nil {
		http.Error(w, fmt.Sprintf("sending response: %v", err), http.StatusInternalServerError)
		return
	}
}

// processAction parses action.Data.Message into a DataFixed, but the shape of
// that JSON differs by dispatch path — discovered by tracing a live "originalMessage
// is empty" failure back to internal/processors/direct/default.go in the
// tee-node source, not documented anywhere:
//   - On-chain (InstructionSender): already DataFixed-shaped JSON.
//   - Direct (/direct, what RFQ uses): DirectInstruction-shaped JSON
//     ({opType, opCommand, message}), posted to /action completely unwrapped.
//     Parsing that directly as DataFixed silently leaves OriginalMessage empty
//     because the field is named "message", not "originalMessage" — no error,
//     just a wrong-shaped struct. This is why the original "DO NOT MODIFY"
//     boilerplate needed changing: it only handled the on-chain shape.
func (e *Extension) processAction(action teetypes.Action) (int, []byte) {
	var dataFixed *instruction.DataFixed

	if action.Data.Type == teetypes.Direct {
		di, err := processorutils.Parse[teetypes.DirectInstruction](action.Data.Message)
		if err != nil {
			return http.StatusBadRequest, []byte(fmt.Sprintf("decoding direct instruction: %v", err))
		}

		// SECURITY: only RFQ is allowed via the Direct path. Code-review finding —
		// before the message-shape fix above, KEY/UPDATE sent via /direct always
		// failed with an empty OriginalMessage, which accidentally protected the
		// attested-signer key from being overwritten by anyone who could reach the
		// tunnel. Fixing the parsing bug removed that accidental protection: without
		// this check, anyone could GET /info for the TEE's pubkey, encrypt their own
		// chosen private key, and POST {opType:"KEY",opCommand:"UPDATE"} to /direct —
		// overwriting the exact key RfqSettlement trusts without on-chain
		// re-verification. KEY management stays on the on-chain InstructionSender
		// path only, where dispatching it costs real gas and leaves an auditable
		// sender — not this unauthenticated-at-this-layer endpoint.
		if di.OPType != teeutils.ToHash(config.OPTypeRfq) {
			return http.StatusForbidden, []byte(fmt.Sprintf(
				"op type %s is not permitted via the /direct dispatch path — only %s is", di.OPType.Hex(), config.OPTypeRfq))
		}

		dataFixed = &instruction.DataFixed{
			OPType:          di.OPType,
			OPCommand:       di.OPCommand,
			OriginalMessage: di.Message,
		}
	} else {
		var err error
		dataFixed, err = processorutils.Parse[instruction.DataFixed](action.Data.Message)
		if err != nil {
			return http.StatusBadRequest, []byte(fmt.Sprintf("decoding fixed data: %v", err))
		}
	}

	switch {
	case dataFixed.OPType == teeutils.ToHash(config.OPTypeKey):
		return e.processKey(action, dataFixed)

	case dataFixed.OPType == teeutils.ToHash(config.OPTypeRfq):
		return e.processRfq(action, dataFixed)

	default:
		return http.StatusNotImplemented, []byte(fmt.Sprintf(
			"unsupported op type: received %s, expected %s (%s) or %s (%s)",
			dataFixed.OPType.Hex(),
			teeutils.ToHash(config.OPTypeKey).Hex(), config.OPTypeKey,
			teeutils.ToHash(config.OPTypeRfq).Hex(), config.OPTypeRfq,
		))
	}
}

// processKey routes KEY instructions by OPCommand (UPDATE or SIGN).
func (e *Extension) processKey(action teetypes.Action, df *instruction.DataFixed) (int, []byte) {
	switch {
	case df.OPCommand == teeutils.ToHash(config.OPCommandUpdate):
		ar := e.processKeyUpdate(action, df)
		b, _ := json.Marshal(ar)
		return http.StatusOK, b

	case df.OPCommand == teeutils.ToHash(config.OPCommandSign):
		ar := e.processKeySign(action, df)
		b, _ := json.Marshal(ar)
		return http.StatusOK, b

	default:
		return http.StatusNotImplemented, []byte(fmt.Sprintf(
			"unsupported op command: received %s, expected one of [%s (%s), %s (%s)]",
			df.OPCommand.Hex(),
			teeutils.ToHash(config.OPCommandUpdate).Hex(), config.OPCommandUpdate,
			teeutils.ToHash(config.OPCommandSign).Hex(), config.OPCommandSign,
		))
	}
}

// processRfq routes RFQ instructions by OPCommand (OPEN, QUOTE, or CLOSE).
// Dispatched via the proxy's /direct endpoint, not InstructionSender.
func (e *Extension) processRfq(action teetypes.Action, df *instruction.DataFixed) (int, []byte) {
	var ar teetypes.ActionResult
	switch {
	case df.OPCommand == teeutils.ToHash(config.OPCommandRfqOpen):
		ar = e.processRfqOpen(action, df)
	case df.OPCommand == teeutils.ToHash(config.OPCommandRfqQuote):
		ar = e.processRfqQuote(action, df)
	case df.OPCommand == teeutils.ToHash(config.OPCommandRfqClose):
		ar = e.processRfqClose(action, df)
	default:
		return http.StatusNotImplemented, []byte(fmt.Sprintf(
			"unsupported RFQ op command: received %s, expected one of [%s, %s, %s]",
			df.OPCommand.Hex(), config.OPCommandRfqOpen, config.OPCommandRfqQuote, config.OPCommandRfqClose,
		))
	}
	b, _ := json.Marshal(ar)
	return http.StatusOK, b
}

// processKeyUpdate decrypts the original message via the TEE node and stores
// the resulting bytes as a secp256k1 private key.
func (e *Extension) processKeyUpdate(action teetypes.Action, df *instruction.DataFixed) teetypes.ActionResult {
	if len(df.OriginalMessage) == 0 {
		return buildResult(action, df, nil, 0, fmt.Errorf("originalMessage is empty"))
	}

	keyBytes, err := decryptViaNode(e.signPort, df.OriginalMessage)
	if err != nil {
		return buildResult(action, df, nil, 0, fmt.Errorf("decryption failed: %v", err))
	}

	privKey, err := parseSecp256k1PrivateKey(keyBytes)
	if err != nil {
		return buildResult(action, df, nil, 0, fmt.Errorf("invalid private key: %v", err))
	}

	e.mu.Lock()
	e.privateKey = privKey
	e.mu.Unlock()

	return buildResult(action, df, nil, 1, nil)
}

// processKeySign signs the original message with the stored private key.
// Returns ABI-encoded (bytes message, bytes signature) in ActionResult.Data.
func (e *Extension) processKeySign(action teetypes.Action, df *instruction.DataFixed) teetypes.ActionResult {
	e.mu.RLock()
	key := e.privateKey
	e.mu.RUnlock()

	if key == nil {
		return buildResult(action, df, nil, 0, fmt.Errorf("no private key stored"))
	}
	if len(df.OriginalMessage) == 0 {
		return buildResult(action, df, nil, 0, fmt.Errorf("originalMessage is empty"))
	}

	sig, err := signECDSA(key, df.OriginalMessage)
	if err != nil {
		return buildResult(action, df, nil, 0, fmt.Errorf("signing failed: %v", err))
	}

	encoded, err := abiEncodeTwo(df.OriginalMessage, sig)
	if err != nil {
		return buildResult(action, df, nil, 0, fmt.Errorf("ABI encoding failed: %v", err))
	}

	return buildResult(action, df, encoded, 1, nil)
}
