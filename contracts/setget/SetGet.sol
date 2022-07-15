// SPDX-License-Identifier: MIT
pragma solidity ^0.8.13;

contract SetGet {
    uint public counter;

    function get() public view returns (uint) {
        return counter;
    }

    function set(uint i) public {
        counter = i;
    }
}