// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {EIP712} from "@openzeppelin/contracts/utils/cryptography/EIP712.sol";
import {ECDSA} from "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {IERC20Metadata} from "@openzeppelin/contracts/token/ERC20/extensions/IERC20Metadata.sol";
import {SafeERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import {Ownable} from "@openzeppelin/contracts/access/Ownable.sol";
import {Math} from "@openzeppelin/contracts/utils/math/Math.sol";

/// @dev Minimal FTSOv2 feed-read interface.
interface IFtsoV2 {
    function getFeedById(bytes21 feedId) external view returns (uint256 value, int8 decimals, uint64 timestamp);
}

/// @title RfqSettlement
/// @notice Atomic settlement for a sealed-bid RFQ matched off-chain inside a Flare
///         Confidential Compute TEE extension.
/// @dev Verifies only the TEE's attestation over `Fill`; `RfqIntent` and `Quote`
///      are checked in the extension and deliberately not re-checked here. Holds no
///      funds — settlement is transferFrom-based and can fail. See docs/TRUST.md.
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

    /// @notice Base asset, registry-resolved at deploy time.
    IERC20 public immutable baseToken;

    /// @notice The quote asset for this deployment (USDT0 on Coston2).
    IERC20 public immutable quoteToken;

    /// @notice Optional FTSO price bound. Unset on the live deployment, which
    /// disables the check entirely.
    IFtsoV2 public ftso;
    bytes21 public ftsoFeedId;
    uint256 public ftsoToleranceBps; // e.g. 1000 = 10%
    uint256 public ftsoMaxStaleness; // seconds

    /// @dev Owner allowlist, standing in for a TeeExtensionRegistry check.
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

    /// @notice Settle a matched RFQ. Reverts on every failure path so callers can
    /// surface a specific reason.
    function settle(Fill calldata fill, bytes calldata attestationSig) external {
        if (settled[fill.rfqId]) revert AlreadySettled(fill.rfqId);
        if (block.timestamp > fill.expiry) revert Expired(fill.expiry, block.timestamp);

        address signer = hashFill(fill).recover(attestationSig);
        if (!isAttestedSigner[signer]) revert UntrustedSigner(signer);

        _checkFtsoBound(fill.price);

        settled[fill.rfqId] = true;

        uint8 quoteDecimals = IERC20Metadata(address(quoteToken)).decimals();
        uint8 baseDecimals = IERC20Metadata(address(baseToken)).decimals();
        // size * price must happen inside mulDiv: as a plain argument it can
        // overflow-revert before mulDiv's guard ever runs. Scaling price by the
        // small token-metadata exponent first is safe.
        uint256 quoteAmount =
            Math.mulDiv(fill.size, fill.price * (10 ** quoteDecimals), (10 ** baseDecimals) * 1e18);

        // Floor division can zero quoteAmount for a tiny price while size still
        // transfers in full. The signer is otherwise trusted completely.
        if (fill.size == 0 || quoteAmount == 0) revert ZeroAmount(fill.size, quoteAmount);

        (address baseFrom, address baseTo, address quoteFrom, address quoteTo) = fill.side == Side.TakerBuy
            ? (fill.maker, fill.taker, fill.taker, fill.maker)
            : (fill.taker, fill.maker, fill.maker, fill.taker);

        baseToken.safeTransferFrom(baseFrom, baseTo, fill.size);
        quoteToken.safeTransferFrom(quoteFrom, quoteTo, quoteAmount);

        emit Filled(fill.rfqId, fill.taker, fill.maker, fill.side, fill.size, fill.price);
    }

    /// @dev 10% tolerance; 2% proved too tight for a 60-120s expiry window.
    function _checkFtsoBound(uint256 price) internal view {
        if (address(ftso) == address(0)) return;

        (uint256 refValue, int8 refDecimals, uint64 refTimestamp) = ftso.getFeedById(ftsoFeedId);

        // Bound-check the oracle-supplied exponent before using it — a wrong feed
        // id or an oracle bug returning an out-of-range value would otherwise let
        // `10 ** uint256(wadExponent)` overflow uint256's checked-math revert,
        // bricking every settle() until the owner notices.
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
