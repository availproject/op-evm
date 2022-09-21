// SPDX-License-Identifier: MIT
pragma solidity ^0.8.7;

import "node_modules/@openzeppelin/contracts/utils/Address.sol";


contract Staking {
    using Address for address;

    // Parameters
    uint128 public constant STAKING_THRESHOLD = 1 ether;

    // Properties
    address[] public _sequencers;
    mapping(address => bool) public _addressToIsSequencer;
    mapping(address => uint256) public _addressToStakedAmount;
    mapping(address => uint256) public _addressToSequencerIndex;
    uint256 public _stakedAmount;
    uint256 public _minimumNumSequencers;
    uint256 public _maximumNumSequencers;

    // Events
    event Staked(address indexed account, uint256 amount);

    event Unstaked(address indexed account, uint256 amount);

    // Modifiers
    modifier onlyEOA() {
        require(!msg.sender.isContract(), "Only EOA can call function");
        _;
    }

    modifier onlyStaker() {
        require(
            _addressToStakedAmount[msg.sender] > 0,
            "Only staker can call function"
        );
        _;
    }

    constructor(uint256 minNumSequencers, uint256 maxNumSequencers) {
        require(
            minNumSequencers <= maxNumSequencers,
            "Min sequencers num can not be greater than max num of sequencers"
        );
        _minimumNumSequencers = minNumSequencers;
        _maximumNumSequencers = maxNumSequencers;
    }

    // View functions
    function CurrentStakedAmount() public view returns (uint256) {
        return _stakedAmount;
    }

    function CurrentSequencers() public view returns (address[] memory) {
        return _sequencers;
    }

    function IsSequencer(address addr) public view returns (bool) {
        return _addressToIsSequencer[addr];
    }

    function AccountStake(address addr) public view returns (uint256) {
        return _addressToStakedAmount[addr];
    }

    function MinNumSequencers() public view returns (uint256) {
        return _minimumNumSequencers;
    }

    function MaxNumSequencers() public view returns (uint256) {
        return _maximumNumSequencers;
    }

    // Public functions
    receive() external payable onlyEOA {
        _stake();
    }

    function stake() public payable onlyEOA {
        _stake();
    }

    function unstake() public onlyEOA onlyStaker {
        _unstake();
    }

    // Private functions
    function _stake() private {
        _stakedAmount += msg.value;
        _addressToStakedAmount[msg.sender] += msg.value;

        if (_canBecomeSequencer(msg.sender)) {
            _appendToSequencerSet(msg.sender);
        }

        emit Staked(msg.sender, msg.value);
    }

    function _unstake() private {
        uint256 amount = _addressToStakedAmount[msg.sender];

        _addressToStakedAmount[msg.sender] = 0;
        _stakedAmount -= amount;

        if (_isSequencer(msg.sender)) {
            _deleteFromSequencers(msg.sender);
        }

        payable(msg.sender).transfer(amount);
        emit Unstaked(msg.sender, amount);
    }

    function _deleteFromSequencers(address staker) private {
        require(
            _sequencers.length > _minimumNumSequencers,
            "Sequencers can't be less than the minimum required sequencer num"
        );

        require(
            _addressToSequencerIndex[staker] < _sequencers.length,
            "index out of range"
        );

        // index of removed address
        uint256 index = _addressToSequencerIndex[staker];
        uint256 lastIndex = _sequencers.length - 1;

        if (index != lastIndex) {
            // exchange between the element and last to pop for delete
            address lastAddr = _sequencers[lastIndex];
            _sequencers[index] = lastAddr;
            _addressToSequencerIndex[lastAddr] = index;
        }

        _addressToIsSequencer[staker] = false;
        _addressToSequencerIndex[staker] = 0;
        _sequencers.pop();
    }

    function _appendToSequencerSet(address newSequencer) private {
        require(
            _sequencers.length < _maximumNumSequencers,
            "Sequencer set has reached full capacity"
        );

        _addressToIsSequencer[newSequencer] = true;
        _addressToSequencerIndex[newSequencer] = _sequencers.length;
        _sequencers.push(newSequencer);
    }

    function _isSequencer(address account) private view returns (bool) {
        return _addressToIsSequencer[account];
    }

    function _canBecomeSequencer(address account) private view returns (bool) {
        return
            !_isSequencer(account) &&
            _addressToStakedAmount[account] >= STAKING_THRESHOLD;
    } 
}