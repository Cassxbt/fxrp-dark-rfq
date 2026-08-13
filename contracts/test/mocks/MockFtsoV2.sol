// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {IFtsoV2} from "../../src/RfqSettlement.sol";

contract MockFtsoV2 is IFtsoV2 {
    uint256 public value;
    int8 public decimalsValue;
    uint64 public timestamp;

    function setFeed(uint256 _value, int8 _decimals, uint64 _timestamp) external {
        value = _value;
        decimalsValue = _decimals;
        timestamp = _timestamp;
    }

    function getFeedById(bytes21) external view returns (uint256, int8, uint64) {
        return (value, decimalsValue, timestamp);
    }
}
