// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {Test} from "forge-std/Test.sol";
import {RfqSettlement, IFtsoV2} from "../src/RfqSettlement.sol";
import {MockERC20} from "./mocks/MockERC20.sol";
import {MockFtsoV2} from "./mocks/MockFtsoV2.sol";

/// @dev Required per BUILD-SPEC.md §2.1: buy 1e6 FXRP-units (1 FXRP, 6 decimals)
/// at 2.00 mUSD, and the equivalent sell — plus FTSO-bound coverage, mocked per
/// the spec's own instruction ("tests must mock the FTSO feed").
contract RfqSettlementTest is Test {
    MockERC20 fxrp; // 6 decimals, matches FXRP's confirmed on-chain decimals
    MockERC20 musd; // 18 decimals, our choice
    MockFtsoV2 ftso;
    RfqSettlement rfq;

    uint256 signerPk = 0xA11CE;
    address signer;
    address owner = address(0xEE);
    address taker = address(0x1);
    address maker = address(0x2);

    function setUp() public {
        signer = vm.addr(signerPk);
        fxrp = new MockERC20("Test FXRP", "FXRP", 6);
        musd = new MockERC20("Mock USD", "mUSD", 18);
        ftso = new MockFtsoV2();

        rfq = new RfqSettlement(fxrp, musd, owner);

        vm.prank(owner);
        rfq.setAttestedSigner(signer, true);

        fxrp.mint(maker, 100e6);
        fxrp.mint(taker, 100e6);
        musd.mint(taker, 1_000e18);
        musd.mint(maker, 1_000e18);

        vm.prank(taker);
        musd.approve(address(rfq), type(uint256).max);
        vm.prank(taker);
        fxrp.approve(address(rfq), type(uint256).max);
        vm.prank(maker);
        fxrp.approve(address(rfq), type(uint256).max);
        vm.prank(maker);
        musd.approve(address(rfq), type(uint256).max);
    }

    function _sign(RfqSettlement.Fill memory fill) internal view returns (bytes memory) {
        bytes32 digest = rfq.hashFill(fill);
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(signerPk, digest);
        return abi.encodePacked(r, s, v);
    }

    /// Buy 1 FXRP (1e6 units, 6 decimals) at 2.00 mUSD (2e18 WAD) — decimal math
    /// per BUILD-SPEC.md §2.1: quoteAmount = size*price*10^quoteDec/(10^baseDec*1e18)
    /// = 1e6 * 2e18 * 1e18 / (1e6 * 1e18) = 2e18 (i.e. exactly 2.00 mUSD).
    function test_TakerBuy_Settles_CorrectDecimalMath() public {
        RfqSettlement.Fill memory fill = RfqSettlement.Fill({
            rfqId: keccak256("rfq-buy-1"),
            taker: taker,
            maker: maker,
            side: RfqSettlement.Side.TakerBuy,
            size: 1e6,
            price: 2e18,
            expiry: block.timestamp + 120
        });

        uint256 takerFxrpBefore = fxrp.balanceOf(taker);
        uint256 makerMusdBefore = musd.balanceOf(maker);

        rfq.settle(fill, _sign(fill));

        assertEq(fxrp.balanceOf(taker), takerFxrpBefore + 1e6, "taker should receive exactly 1 FXRP");
        assertEq(musd.balanceOf(maker), makerMusdBefore + 2e18, "maker should receive exactly 2.00 mUSD, not 2e-12");
        assertTrue(rfq.settled(fill.rfqId));
    }

    /// Equivalent sell: taker sends 1 FXRP, receives 2.00 mUSD from maker.
    function test_TakerSell_Settles_CorrectDecimalMath() public {
        RfqSettlement.Fill memory fill = RfqSettlement.Fill({
            rfqId: keccak256("rfq-sell-1"),
            taker: taker,
            maker: maker,
            side: RfqSettlement.Side.TakerSell,
            size: 1e6,
            price: 2e18,
            expiry: block.timestamp + 120
        });

        uint256 makerFxrpBefore = fxrp.balanceOf(maker);
        uint256 takerMusdBefore = musd.balanceOf(taker);

        rfq.settle(fill, _sign(fill));

        assertEq(fxrp.balanceOf(maker), makerFxrpBefore + 1e6, "maker should receive exactly 1 FXRP");
        assertEq(musd.balanceOf(taker), takerMusdBefore + 2e18, "taker should receive exactly 2.00 mUSD");
    }

    function test_RevertsOnReplay() public {
        RfqSettlement.Fill memory fill = RfqSettlement.Fill({
            rfqId: keccak256("rfq-replay"),
            taker: taker,
            maker: maker,
            side: RfqSettlement.Side.TakerBuy,
            size: 1e6,
            price: 2e18,
            expiry: block.timestamp + 120
        });
        bytes memory sig = _sign(fill);
        rfq.settle(fill, sig);

        // Sign computed before expectRevert: vm.expectRevert only guards the very
        // next external call, and _sign() itself calls the view fn hashFill() first —
        // inlining it here would make expectRevert catch that harmless call instead.
        vm.expectRevert(abi.encodeWithSelector(RfqSettlement.AlreadySettled.selector, fill.rfqId));
        rfq.settle(fill, sig);
    }

    function test_RevertsOnExpiry() public {
        RfqSettlement.Fill memory fill = RfqSettlement.Fill({
            rfqId: keccak256("rfq-expired"),
            taker: taker,
            maker: maker,
            side: RfqSettlement.Side.TakerBuy,
            size: 1e6,
            price: 2e18,
            expiry: block.timestamp + 1
        });
        bytes memory sig = _sign(fill);
        vm.warp(block.timestamp + 2);

        vm.expectRevert(abi.encodeWithSelector(RfqSettlement.Expired.selector, fill.expiry, block.timestamp));
        rfq.settle(fill, sig);
    }

    function test_RevertsOnUntrustedSigner() public {
        uint256 badPk = 0xBAD;
        RfqSettlement.Fill memory fill = RfqSettlement.Fill({
            rfqId: keccak256("rfq-bad-signer"),
            taker: taker,
            maker: maker,
            side: RfqSettlement.Side.TakerBuy,
            size: 1e6,
            price: 2e18,
            expiry: block.timestamp + 120
        });
        bytes32 digest = rfq.hashFill(fill);
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(badPk, digest);
        bytes memory badSig = abi.encodePacked(r, s, v);

        vm.expectRevert(abi.encodeWithSelector(RfqSettlement.UntrustedSigner.selector, vm.addr(badPk)));
        rfq.settle(fill, badSig);
    }

    /// FTSO bound: mocked feed per BUILD-SPEC.md's own instruction (the live XRP/USD
    /// price is not $2.00, so an unmocked test would fail this check). Price within
    /// the 10% tolerance settles normally.
    function test_FtsoBound_WithinTolerance_Settles() public {
        vm.prank(owner);
        rfq.setFtsoBound(IFtsoV2(address(ftso)), bytes21(0), 1000, 300); // 10% bps, 300s staleness
        ftso.setFeed(2 * 10 ** 5, -5, uint64(block.timestamp)); // 2.00 in 5-decimal feed units -> 2e18 WAD

        RfqSettlement.Fill memory fill = RfqSettlement.Fill({
            rfqId: keccak256("rfq-ftso-ok"),
            taker: taker,
            maker: maker,
            side: RfqSettlement.Side.TakerBuy,
            size: 1e6,
            price: 2e18,
            expiry: block.timestamp + 120
        });

        rfq.settle(fill, _sign(fill));
        assertTrue(rfq.settled(fill.rfqId));
    }

    /// Price far outside the FTSO bound reverts — this is the check that stops a
    /// joke quote from winning, per BUILD-SPEC.md §2.1.
    function test_FtsoBound_OutsideTolerance_Reverts() public {
        vm.prank(owner);
        rfq.setFtsoBound(IFtsoV2(address(ftso)), bytes21(0), 1000, 300);
        ftso.setFeed(2 * 10 ** 5, -5, uint64(block.timestamp)); // reference 2.00

        RfqSettlement.Fill memory fill = RfqSettlement.Fill({
            rfqId: keccak256("rfq-ftso-bad"),
            taker: taker,
            maker: maker,
            side: RfqSettlement.Side.TakerBuy,
            size: 1e6,
            price: 1e18, // 1.00 — 50% away from the 2.00 reference, well outside 10%
            expiry: block.timestamp + 120
        });

        bytes memory sig = _sign(fill);
        vm.expectRevert();
        rfq.settle(fill, sig);
    }

    function test_FtsoBound_StaleFeed_Reverts() public {
        vm.prank(owner);
        rfq.setFtsoBound(IFtsoV2(address(ftso)), bytes21(0), 1000, 300);
        ftso.setFeed(2 * 10 ** 5, -5, uint64(block.timestamp));
        vm.warp(block.timestamp + 301);

        RfqSettlement.Fill memory fill = RfqSettlement.Fill({
            rfqId: keccak256("rfq-ftso-stale"),
            taker: taker,
            maker: maker,
            side: RfqSettlement.Side.TakerBuy,
            size: 1e6,
            price: 2e18,
            expiry: block.timestamp + 120
        });

        bytes memory sig = _sign(fill);
        vm.expectRevert();
        rfq.settle(fill, sig);
    }
}
