// contracts/MyContract.sol
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "node_modules/@openzeppelin/contracts/access/Ownable.sol";

contract Contract is Ownable {
    uint public counter;

    function get() public view returns (uint) {
        return counter;
    }

    function set(uint i) public {
        counter = i;
    }
}