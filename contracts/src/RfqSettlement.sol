// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {EIP712} from "@openzeppelin/contracts/utils/cryptography/EIP712.sol";
import {ECDSA} from "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {IERC20Metadata} from "@openzeppelin/contracts/token/ERC20/extensions/IERC20Metadata.sol";
import {SafeERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import {Ownable} from "@openzeppelin/contracts/access/Ownable.sol";
import {Math} from "@openzeppelin/contracts/utils/math/Math.sol";

/// @dev Minimal Flare FTSOv2 feed-read interface. Feed id for Coston2 XRP/USD
/// must be confirmed live before deployment — see BUILD-SPEC.md §2.1 / §4.1b.
interface IFtsoV2 {
    function getFeedById(bytes21 feedId) external view returns (uint256 value, int8 decimals, uint64 timestamp);
}

/// @title RfqSettlement
/// @notice Atomic settlement for a sealed-bid FXRP/quoteToken RFQ matched off-chain
///         inside a Flare Confidential Compute TEE extension. See BUILD-SPEC.md for
///         the full design, threat model, and disclosed limitations.
///
/// @dev Trust model, stated plainly (BUILD-SPEC.md §2.1 round 3):
///      - This contract verifies only the TEE's own attestation signature over `Fill`.
///        Taker/maker intent signatures (`RfqIntent`, `Quote`) are verified off-chain,
///        inside the Go extension, and are NOT re-checked here — a deliberate scope
///        choice under the SIMULATED_TEE trust model, not an oversight.
///      - `isAttestedSigner` is an MVP owner-controlled allowlist. Production would
///        check the signer against FlareTeeManager's live registry for our extension
///        ID; that registry's exact ABI is unconfirmed against the deployed scaffold
///        as of this commit (BUILD-SPEC.md §4.1b) — do not assume this is the final
///        integration point.
///      - Holds no funds. Settlement is `transferFrom`-based and best-effort, not a
///        binding commitment — see the non-binding-fill disclosure in the spec.
contract RfqSettlement is EIP712, Ownable {
    using SafeERC20 for IERC20;
    using ECDSA for bytes32;

    enum Side {
        TakerBuy,
        TakerSell
    }

    struct Fill {
        bytes32 rfqId;
        address taker;
        address maker;
        Side side;
        uint256 size; // base-token units; decimals read live off baseToken, never hardcoded
        uint256 price; // quote per base, WAD (1e18)
        uint256 expiry;
    }

    bytes32 private constant FILL_TYPEHASH =
        keccak256("Fill(bytes32 rfqId,address taker,address maker,uint8 side,uint256 size,uint256 price,uint256 expiry)");

    /// @notice FXRP (or whatever base asset this deployment settles), registry-resolved
    /// at deploy time — never a hardcoded/guessed address.
    IERC20 public immutable baseToken;

    /// @notice mUSD (or USDT0 if swapped in per BUILD-SPEC.md §2.1) — the quote asset.
    IERC20 public immutable quoteToken;

    /// @notice Optional FTSO price bound. Set to address(0) to disable (e.g. in tests
    /// that don't want to depend on a live feed) — but per spec, if this check isn't
    /// wired up for the real deployment, the FTSO differentiator claim comes out of
    /// the README (BUILD-SPEC.md §2.1).
    IFtsoV2 public ftso;
    bytes21 public ftsoFeedId;
    uint256 public ftsoToleranceBps; // e.g. 1000 = 10%, per BUILD-SPEC.md round 2 §3.2
    uint256 public ftsoMaxStaleness; // seconds

    /// @dev MVP placeholder for the real TeeExtensionRegistry check — see contract-level NatSpec.
    mapping(address => bool) public isAttestedSigner;

    mapping(bytes32 => bool) public settled;

    event Filled(bytes32 indexed rfqId, address indexed taker, address indexed maker, Side side, uint256 size, uint256 price);
    event AttestedSignerSet(address indexed signer, bool allowed);
    event FtsoBoundSet(address ftso, bytes21 feedId, uint256 toleranceBps, uint256 maxStaleness);

    error AlreadySettled(bytes32 rfqId);
    error Expired(uint256 expiry, uint256 nowTs);
    error UntrustedSigner(address signer);
    error PriceOutOfBounds(uint256 price, uint256 refPrice, uint256 toleranceBps);
    error StaleFeed(uint64 feedTimestamp, uint256 nowTs, uint256 maxStaleness);
    error ZeroAmount(uint256 size, uint256 quoteAmount);
    error InvalidToleranceBps(uint256 toleranceBps);
    error FeedDecimalsOutOfRange(int8 refDecimals);

    constructor(IERC20 _baseToken, IERC20 _quoteToken, address initialOwner)
        EIP712("RfqSettlement", "1")
        Ownable(initialOwner)
    {
        baseToken = _baseToken;
        quoteToken = _quoteToken;
    }

    function setAttestedSigner(address signer, bool allowed) external onlyOwner {
        isAttestedSigner[signer] = allowed;
        emit AttestedSignerSet(signer, allowed);
    }

    function setFtsoBound(IFtsoV2 _ftso, bytes21 _feedId, uint256 _toleranceBps, uint256 _maxStaleness) external onlyOwner {
        if (_toleranceBps > 10_000) revert InvalidToleranceBps(_toleranceBps);
        ftso = _ftso;
        ftsoFeedId = _feedId;
        ftsoToleranceBps = _toleranceBps;
        ftsoMaxStaleness = _maxStaleness;
        emit FtsoBoundSet(address(_ftso), _feedId, _toleranceBps, _maxStaleness);
    }

    /// @notice Domain-separated hash of a Fill, for the TEE (or a test harness) to sign.
    function hashFill(Fill calldata fill) public view returns (bytes32) {
        bytes32 structHash = keccak256(
            abi.encode(FILL_TYPEHASH, fill.rfqId, fill.taker, fill.maker, uint8(fill.side), fill.size, fill.price, fill.expiry)
        );
        return _hashTypedDataV4(structHash);
    }

    /// @notice Settle a matched RFQ. Reverts (not silently no-ops) on every failure
    /// path so the frontend can surface a specific reason per BUILD-SPEC.md §2.3.
    function settle(Fill calldata fill, bytes calldata attestationSig) external {
        if (settled[fill.rfqId]) revert AlreadySettled(fill.rfqId);
        if (block.timestamp > fill.expiry) revert Expired(fill.expiry, block.timestamp);

        address signer = hashFill(fill).recover(attestationSig);
        if (!isAttestedSigner[signer]) revert UntrustedSigner(signer);

        _checkFtsoBound(fill.price);

        settled[fill.rfqId] = true;

        uint8 quoteDecimals = IERC20Metadata(address(quoteToken)).decimals();
        uint8 baseDecimals = IERC20Metadata(address(baseToken)).decimals();
        // fill.size * fill.price must not be computed as a plain checked-arithmetic
        // argument — that multiplication alone can overflow-revert before mulDiv ever
        // runs, defeating the point (audit finding). Scaling fill.price by
        // 10**quoteDecimals first is safe in comparison: quoteDecimals is a small
        // token-metadata value (6-18), not an attacker-influenced multiplicand, so
        // mulDiv's own protected size*scaledPrice multiplication is what actually
        // guards against overflow from an oversized fill.size or fill.price.
        uint256 quoteAmount =
            Math.mulDiv(fill.size, fill.price * (10 ** quoteDecimals), (10 ** baseDecimals) * 1e18);

        // Defense-in-depth against a degenerate Fill (buggy/compromised attested
        // signer — the contract otherwise trusts it fully per the disclosed trust
        // model). Floor division can zero out quoteAmount for a tiny price while
        // `size` still transfers in full; caught empirically by the code-review
        // pass, not by the original test suite. size==0 is symmetric nonsense.
        if (fill.size == 0 || quoteAmount == 0) revert ZeroAmount(fill.size, quoteAmount);

        (address baseFrom, address baseTo, address quoteFrom, address quoteTo) = fill.side == Side.TakerBuy
            ? (fill.maker, fill.taker, fill.taker, fill.maker)
            : (fill.taker, fill.maker, fill.maker, fill.taker);

        baseToken.safeTransferFrom(baseFrom, baseTo, fill.size);
        quoteToken.safeTransferFrom(quoteFrom, quoteTo, quoteAmount);

        emit Filled(fill.rfqId, fill.taker, fill.maker, fill.side, fill.size, fill.price);
    }

    /// @dev BUILD-SPEC.md §2.1: 10% tolerance (not the originally-proposed 2%, too
    /// tight for a 60-120s expiry window), plus a staleness check. Skipped entirely
    /// if `ftso` is unset — tests and early iteration don't need a live feed.
    function _checkFtsoBound(uint256 price) internal view {
        if (address(ftso) == address(0)) return;

        (uint256 refValue, int8 refDecimals, uint64 refTimestamp) = ftso.getFeedById(ftsoFeedId);

        // Bound-check the oracle-supplied exponent before using it — a wrong feed
        // id or an oracle bug returning an out-of-range value would otherwise let
        // `10 ** uint256(wadExponent)` overflow uint256's checked-math revert,
        // bricking every settle() until the owner notices (round-2 review finding).
        if (refDecimals < -30 || refDecimals > 30) revert FeedDecimalsOutOfRange(refDecimals);

        if (block.timestamp > refTimestamp && block.timestamp - refTimestamp > ftsoMaxStaleness) {
            revert StaleFeed(refTimestamp, block.timestamp, ftsoMaxStaleness);
        }

        // Normalize the feed value to WAD (1e18): actual price = value * 10^decimals,
        // so WAD value = value * 10^(decimals + 18). For any realistic FTSO decimals
        // (small positive or negative int8), decimals+18 stays positive.
        int256 wadExponent = int256(refDecimals) + 18;
        uint256 refWad = wadExponent >= 0
            ? refValue * (10 ** uint256(wadExponent))
            : refValue / (10 ** uint256(-wadExponent));

        uint256 lower = refWad - (refWad * ftsoToleranceBps) / 10_000;
        uint256 upper = refWad + (refWad * ftsoToleranceBps) / 10_000;

        if (price < lower || price > upper) revert PriceOutOfBounds(price, refWad, ftsoToleranceBps);
    }
}
