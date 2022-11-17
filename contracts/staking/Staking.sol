// SPDX-License-Identifier: MIT
pragma solidity ^0.8.9;

import "node_modules/@openzeppelin/contracts/utils/Address.sol";

contract Staking {
    using Address for address;

    // Parameters
    uint256 public constant DEFAULT_STAKING_THRESHOLD = 1 ether;
    uint256 public constant DEFAULT_MIN_SLASH_PERCENTAGE = 1;
    string public  constant NODE_SEQUENCER = "sequencer";
    string public  constant NODE_WATCHTOWER = "watchtower";
    string public  constant NODE_VALIDATOR = "validator";

    string[] public AVAILABLE_NODE_TYPES = ["sequencer", "watchtower", "validator"];

    // Properties
    uint256 public _minStakingThreshold;
    uint256 public _slashPercentage;

    address[] public _participants;
    mapping(address => bool) public _addressToIsParticipant;
    mapping(address => uint256) public _addressToStakedAmount;
    mapping(address => uint256) public _addressToParticipantIndex;
    mapping(address => string) public _addressToNodeType;
    uint256 public _stakedAmount;

    uint256 public _minimumNumParticipants;
    uint256 public _maximumNumParticipants;
    uint256 public _minimumNumSequencers;
    uint256 public _minimumNumValidators;
    uint256 public _minimumNumWatchtowers;
    uint256 public _maximumNumSequencers;
    uint256 public _maximumNumValidators;
    uint256 public _maximumNumWatchtowers;

    // For now, we are going to have 3 separated array storages for each node type
    address[] public _sequencers;
    mapping(address => bool) public _addressToIsSequencer;
    mapping(address => uint256) public _addressToSequencerIndex;

    address[] public _sequencers_in_probation;
    mapping(address => bool) public _addressToIsSequencerInProbationAddr;
    mapping(address => uint256) public _addressToSequencerInProbationIndex;

    address[] public _watchtowers;
    mapping(address => bool) public _addressToIsWatchtower;
    mapping(address => uint256) public _addressToWatchtowerIndex;

    address[] public _validators;
    mapping(address => bool) public _addressToIsValidator;
    mapping(address => uint256) public _addressToValidatorIndex;

    // Events
    event Staked(address indexed account, uint256 amount);
    event Unstaked(address indexed account, uint256 amount);
    event Slashed(address indexed account, uint256 newAmount, uint256 slashedAmount);

    // Fraud Dispute Resolution Events
    event DisputeResolutionBegan(address indexed account);
    event DisputeResolutionEnded(address indexed account);

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

    constructor(uint256 minNumParticipants, uint256 maxNumParticipants) {
        require(
            minNumParticipants <= maxNumParticipants,
            "Min participants number can not be greater than max num of participants"
        );
        _minimumNumParticipants = minNumParticipants;
        _maximumNumParticipants = maxNumParticipants;
    }

    // PARTICIPANT LIMITER SETTERS-GETTERS

    function SetMinNumParticipants(uint256 minimumNumParticipants) public returns (uint256) {
        _minimumNumParticipants = minimumNumParticipants;
        return _minimumNumParticipants;
    }

    function SetMaxNumParticipants(uint256 maximumNumParticipants) public returns (uint256) {
        _maximumNumParticipants = maximumNumParticipants;
        return _maximumNumParticipants;
    }

    function SetMinNumSequencers(uint256 minimumNumSequencers) public returns (uint256) {
        _minimumNumSequencers = minimumNumSequencers;
        return _minimumNumSequencers;
    }

    function SetMaxNumSequencers(uint256 maximumNumSequencers) public returns (uint256) {
        _maximumNumSequencers = maximumNumSequencers;
        return _maximumNumSequencers;
    }

    function SetMinNumValidators(uint256 minimumNumValidators) public returns (uint256) {
        _minimumNumValidators = minimumNumValidators;
        return _minimumNumValidators;
    }

    function SetMaxNumValidators(uint256 maximumNumValidators) public returns (uint256) {
        _maximumNumValidators = maximumNumValidators;
        return _maximumNumValidators;
    }

    function SetMinNumWatchtowers(uint256 minimumNumWatchtowers) public returns (uint256) {
        _minimumNumWatchtowers = minimumNumWatchtowers;
        return _minimumNumWatchtowers;
    }

    function SetMaxNumWatchtowers(uint256 maximumNumWatchtowers) public returns (uint256) {
        _maximumNumWatchtowers = maximumNumWatchtowers;
        return _maximumNumWatchtowers;
    }

    // VIEW FUNCTIONS 

    function GetSlashPercentage() public view returns (uint256) {
        return _getSlashPercentage();
    }

    function GetMinNumSequencers() public view returns (uint256) {
        return _minimumNumSequencers;
    }

    function GetMaxNumSequencers() public view returns (uint256) {
        return _maximumNumSequencers;
    }

    function GetMinNumValidators() public view returns (uint256) {
        return _minimumNumValidators;
    }

    function GetMaxNumValidators() public view returns (uint256) {
        return _maximumNumValidators;
    }

    function GetMinNumWatchtowers() public view returns (uint256) {
        return _minimumNumWatchtowers;
    }

    function GetMaxNumWatchtowers() public view returns (uint256) {
        return _maximumNumWatchtowers;
    }

    function GetMinNumParticipants() public view returns (uint256) {
        return _minimumNumParticipants;
    }

    function GetMaxNumParticipants() public view returns (uint256) {
        return _maximumNumParticipants;
    }

    function GetAvailableNodeTypes() public view returns (string[] memory) {
        return AVAILABLE_NODE_TYPES;
    }

    function GetCurrentParticipants() public view returns (address[] memory) {
        return _participants;
    }

    function GetCurrentSequencers() public view returns (address[] memory) {
        return _sequencers;
    }

    function GetCurrentWatchtowers() public view returns (address[] memory) {
        return _watchtowers;
    }

    function GetCurrentValidators() public view returns (address[] memory) {
        return _validators;
    }

    function GetCurrentStakedAmount() public view returns (uint256) {
        return _stakedAmount;
    }

    function GetCurrentAccountStakedAmount(address addr) public view returns (uint256) {
        return _addressToStakedAmount[addr];
    }

    function IsSequencer(address addr) public view returns (bool) {
        return _addressToIsSequencer[addr];
    }

    function IsWatchtower(address addr) public view returns (bool) {
        return _addressToIsWatchtower[addr];
    }

    function IsValidator(address addr) public view returns (bool) {
        return _addressToIsValidator[addr];
    }

    function GetCurrentStakingThreshold() public view returns (uint256) {
        return _getStakingThreshold();
    }

    // -- END VIEW FUNCTIONS 


    // PUBLIC FRAUD DISPUTE RESOLUTION FUNCTIONS

    function GetCurrentSequencersInProbation() public view returns (address[] memory) {
        return _sequencers_in_probation;
    }

    function GetIsSequencerInProbation(address sequencerAddr) public view returns (bool) {
        return _isSequencerInProbation(sequencerAddr);
    }

    function BeginDisputeResolution(address sequencerAddr) public {
        _appendToSequencersInProbationSet(sequencerAddr);
        emit DisputeResolutionBegan(sequencerAddr);
    }

    function EndDisputeResolution(address sequencerAddr) public {
        _deleteFromSequencersInProbationSet(sequencerAddr);
        emit DisputeResolutionEnded(sequencerAddr);
    }

    // -- END FRAUD DISPUTE RESOLUTION FUNCTIONS

    // PUBLIC STAKING FUNCTIONS

    function SetSlashPercentage(uint256 newPercentage) public onlyEOA {
        _slashPercentage = newPercentage;
    }

    function SetStakingMinThreshold(uint256 newThreshold) public onlyEOA {
        _minStakingThreshold = newThreshold;
    }

    function stake(string memory nodeType) public payable onlyEOA {
        _stake(nodeType);
    }

    function unstake() public onlyEOA onlyStaker {
        _unstake();
    }

    // TODO: Cannot be only staker but only watchtower for example
    function slash(address slashAddr, uint256 slashAmount) public onlyEOA onlyStaker {
        _slash(slashAddr, slashAmount);
    }

    // -- END PUBLIC STAKING FUNCTIONS

    // PRIVATE FUNCTIONS

    function _stake(string memory nodeType) private {
        require(
            _isNodeTypeArgumentSequencer(nodeType) || _isNodeTypeArgumentWatchtower(nodeType) || _isNodeTypeArgumentValidator(nodeType),
            "Provided node type has to match available node types"
        );

        require(
            msg.value >= _getStakingThreshold(),
            "Insuficient staking amount provided. Has to be larger than staking minimum threshold"
        );

        require(
            _canBecomeSequencer(msg.sender, nodeType),
            "Sender is already sequencer. Rejecting stake request."
        );

        require(
            _canBecomeWatchTower(msg.sender, nodeType),
            "Sender is already watchtower. Rejecting stake request."
        );

        require(
            _canBecomeValidator(msg.sender, nodeType),
            "Sender is already validator. Rejecting stake request."
        );

        _stakedAmount += msg.value;
        _addressToStakedAmount[msg.sender] += msg.value;

        // Function will ensure that participant is appended only once.
        _appendToParticipantSet(msg.sender);

        if (_isNodeTypeArgumentSequencer(nodeType)) {
            _appendToSequencerSet(msg.sender);
        } else if (_isNodeTypeArgumentWatchtower(nodeType)) {
            _appendToWatchtowerSet(msg.sender);
        } else if (_isNodeTypeArgumentValidator(nodeType)) {
            _appendToValidatorSet(msg.sender);
        }

        emit Staked(msg.sender, msg.value);
    }

    function _unstake() private {
        require(
            _isParticipant(msg.sender),
            "Sender has to be part of staking poll in order to unstake its share. Unstake rejected."
        );

        uint256 amount = _addressToStakedAmount[msg.sender];

        _addressToStakedAmount[msg.sender] = 0;
        _stakedAmount -= amount;

        // Delete from participants
        _deleteFromParticipants(msg.sender);

        // If possible, delete from the sequencers list
        if(_isSequencer(msg.sender)) {
            _deleteFromSequencers(msg.sender);
        }

        // If possible, delete from the watchtower list
        if(_isWatchtower(msg.sender)) {
            _deleteFromWatchtowers(msg.sender);
        }
        
        // If possible, delete from the validator list
        if(_isValidator(msg.sender)) {
            _deleteFromValidators(msg.sender);
        }

        payable(msg.sender).transfer(amount);
        emit Unstaked(msg.sender, amount);
    }

    // TODO:
    // - Make sure slashing amount is properly transferred to appropriate participants
    // - Append slashed with time interval into the mapping so next sequencer set won't include it.
    //
    function _slash(address slashAddr, uint256 slashAmount) private {
        uint256 amount = _addressToStakedAmount[msg.sender];

        require(
            slashAmount != 0,
            "Slash amount needs to be provided."
        );

        require(
            slashAmount < amount,
            "Slash amount cannot be greater than staked amount"
        );

        uint256 newStakedAmount = amount - slashAmount;
        _addressToStakedAmount[msg.sender] = newStakedAmount;
        _stakedAmount -= newStakedAmount;

        payable(msg.sender).transfer(slashAmount);
        emit Slashed(msg.sender, newStakedAmount, slashAmount);

        if(_isSequencerInProbation(slashAddr)) {
            _deleteFromSequencersInProbationSet(slashAddr);
            emit DisputeResolutionEnded(slashAddr);
        }
    }

    function _deleteFromParticipants(address staker) private {
        require(
            _participants.length > _minimumNumParticipants,
            "Staking participants can't be less than the minimum required participant num"
        );

        require(
            _addressToParticipantIndex[staker] < _participants.length,
            "participant index out of range in mapping"
        );

        // index of removed address
        uint256 index = _addressToParticipantIndex[staker];
        uint256 lastIndex = _participants.length - 1;

        if (index != lastIndex) {
            // exchange between the element and last to pop for delete
            address lastAddr = _participants[lastIndex];
            _participants[index] = lastAddr;
            _addressToParticipantIndex[lastAddr] = index;
        }

        _addressToIsParticipant[staker] = false;
        _addressToParticipantIndex[staker] = 0;
        _participants.pop();
    }


    function _deleteFromSequencers(address staker) private {
        require(
            _addressToSequencerIndex[staker] < _sequencers.length,
            "sequencer index out of range in mapping"
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

    function _deleteFromWatchtowers(address staker) private {
        require(
            _addressToWatchtowerIndex[staker] < _watchtowers.length,
            "watchtower index out of range in mapping"
        );

        // index of removed address
        uint256 index = _addressToWatchtowerIndex[staker];
        uint256 lastIndex = _watchtowers.length - 1;

        if (index != lastIndex) {
            // exchange between the element and last to pop for delete
            address lastAddr = _watchtowers[lastIndex];
            _watchtowers[index] = lastAddr;
            _addressToWatchtowerIndex[lastAddr] = index;
        }

        _addressToIsWatchtower[staker] = false;
        _addressToWatchtowerIndex[staker] = 0;
        _watchtowers.pop();
    }

    function _deleteFromValidators(address staker) private {
        require(
            _addressToValidatorIndex[staker] < _validators.length,
            "validator index out of range in mapping"
        );

        // index of removed address
        uint256 index = _addressToValidatorIndex[staker];
        uint256 lastIndex = _validators.length - 1;

        if (index != lastIndex) {
            // exchange between the element and last to pop for delete
            address lastAddr = _validators[lastIndex];
            _validators[index] = lastAddr;
            _addressToValidatorIndex[lastAddr] = index;
        }

        _addressToIsValidator[staker] = false;
        _addressToValidatorIndex[staker] = 0;
        _validators.pop();
    }

    function _appendToSequencersInProbationSet(address newMaliciousAddr) private {
        _addressToIsSequencerInProbationAddr[newMaliciousAddr] = true;
        _addressToSequencerInProbationIndex[newMaliciousAddr] = _sequencers_in_probation.length;
        _sequencers_in_probation.push(newMaliciousAddr);
    }

    function _deleteFromSequencersInProbationSet(address participant) private {
        require(
            _isSequencerInProbation(participant),
            "Address has to be in probation in order to delete it from the probation sequencers list."
        );

        require(
            _addressToSequencerInProbationIndex[participant] < _sequencers_in_probation.length,
            "malicious participant index out of range in mapping"
        );

        // index of removed address
        uint256 index = _addressToSequencerInProbationIndex[participant];
        uint256 lastIndex = _sequencers_in_probation.length - 1;

        if (index != lastIndex) {
            // exchange between the element and last to pop for delete
            address lastAddr = _sequencers_in_probation[lastIndex];
            _sequencers_in_probation[index] = lastAddr;
            _addressToSequencerInProbationIndex[lastAddr] = index;
        }

        _addressToIsSequencerInProbationAddr[participant] = false;
        _addressToSequencerInProbationIndex[participant] = 0;
        _sequencers_in_probation.pop();
    }

    // Append to participant set only if participant is not already set.
    // Due to possibility to be multi-node participant, we need to make this check.
    function _appendToParticipantSet(address newParticipant) private {
        if(!_addressToIsParticipant[newParticipant]) {
            _addressToIsParticipant[newParticipant] = true;
            _addressToParticipantIndex[newParticipant] = _participants.length;
            _participants.push(newParticipant);
        }
    }

    function _appendToSequencerSet(address newSequencer) private {
        _addressToIsSequencer[newSequencer] = true;
        _addressToSequencerIndex[newSequencer] = _sequencers.length;
        _sequencers.push(newSequencer);
    }

    function _appendToWatchtowerSet(address newWatchtower) private {
        _addressToIsWatchtower[newWatchtower] = true;
        _addressToWatchtowerIndex[newWatchtower] = _watchtowers.length;
        _watchtowers.push(newWatchtower);
    }

    function _appendToValidatorSet(address newValidator) private {
        _addressToIsValidator[newValidator] = true;
        _addressToValidatorIndex[newValidator] = _validators.length;
        _validators.push(newValidator);
    }

    function _isSequencerInProbation(address account) private view returns (bool) {
        return _addressToIsSequencerInProbationAddr[account];
    }

    function _isParticipant(address account) private view returns (bool) {
        return _addressToIsParticipant[account];
    }

    function _isSequencer(address account) private view returns (bool) {
        return _addressToIsSequencer[account];
    }
    
    function _isWatchtower(address account) private view returns (bool) {
        return _addressToIsWatchtower[account];
    }

    function _isValidator(address account) private view returns (bool) {
        return _addressToIsValidator[account];
    }

    // Check if it is already sequencer as participant cannot become twice sequencer
    function _canBecomeSequencer(address account, string memory nodeType) private view returns (bool)  {
        if(_isNodeTypeArgumentSequencer(nodeType) && _addressToIsSequencer[account]) {
           return false; 
        }

        return true;
    }

    // Check if it is already watchtower as participant cannot become twice watchtower
    function _canBecomeWatchTower(address account, string memory nodeType) private view returns (bool)  {
        if(_isNodeTypeArgumentWatchtower(nodeType) && _addressToIsWatchtower[account]) {
           return false; 
        }

        return true;
    }

    // Check if it is already validator as participant cannot become twice validator
    function _canBecomeValidator(address account, string memory nodeType) private view returns (bool)  {
        if(_isNodeTypeArgumentValidator(nodeType) && _addressToIsValidator[account]) {
           return false; 
        }

        return true;
    }


    function _isNodeTypeArgumentSequencer(string memory _nodeType) private pure returns (bool) {
        // Compare string keccak256 hashes to check equality
        if (keccak256(abi.encodePacked(NODE_SEQUENCER)) == keccak256(abi.encodePacked(_nodeType))) {
            return true;
        }

        return false;
    }

    function _isNodeTypeArgumentWatchtower(string memory _nodeType) private pure returns (bool) {
        // Compare string keccak256 hashes to check equality
        if (keccak256(abi.encodePacked(NODE_WATCHTOWER)) == keccak256(abi.encodePacked(_nodeType))) {
            return true;
        }

        return false;
    }

    function _isNodeTypeArgumentValidator(string memory _nodeType) private pure returns (bool) {
        // Compare string keccak256 hashes to check equality
        if (keccak256(abi.encodePacked(NODE_VALIDATOR)) == keccak256(abi.encodePacked(_nodeType))) {
            return true;
        }

        return false;
    }

    function _getStakingThreshold() private view returns (uint256) {
        if (_minStakingThreshold > 0) {
            return _minStakingThreshold;
        } else {
            return DEFAULT_STAKING_THRESHOLD;
        }
    }

    function _getSlashPercentage() private view returns (uint256) {
        if (_slashPercentage <= 0) {
            return DEFAULT_MIN_SLASH_PERCENTAGE;
        } else {
            return _slashPercentage;
        }
    }

    // -- END PRIVATE FUNCTIONS
}