# Polygon Avail - Settlement Layer

## Background

Polygon Avail provides a solid Data Availability layer for rollups, but it lacks a settlement layer that provides communication channel for rollups.

Original plan was to adapt [Optimism](https://www.optimism.io/) for this, but after learning the plans for the new [Bedrock design](https://dev.optimism.io/introducing-optimism-bedrock/) and the new challenging scheme for fault proof generation, it was deemed to be more reasonable to start from scratch.

## Iteration 0 - research

Start by studying [minigeth](https://github.com/ethereum-optimism/minigeth/) code, if it could be adapted so that blocks are stored in Avail and the node would run sort of stand-alone.

### Notes from `minigeth`

- In `go-ethereum` `ethdb` provides abstraction over actual DB backend implementations.
    - In `minigeth` this has been mostly stripped off and the state gets handled via `trie` directly.


### Notes from Polygon Edge

- Modular EVM compatible blockchain framework written in Go.
- _Might_ have tight coupling between execution, p2p and consensus layer.
- Meet on Wednesday 22nd June.

### High Overview Architecture Diagram

#### Basic Sequence Diagram

Idea is to clearly outline process of the 0th iteration from the end customer to data being written and consumed to/from Avail.

Example uses a simple increment smart contract where we can execute Get(int) and Set(int).

```sequence
Participant Customer
Participant EVM Node
Participant Avail

Customer->EVM Node: Contract Deployment
EVM Node->Avail: ???
Note right of Avail: Do we need to do anything\nat this stage?
Avail->EVM Node: ???
EVM Node->Customer: Contract Deployed
Customer->EVM Node: Sends Set(1) Transaction
EVM Node->Avail: Writes Set() Transaction
Avail->EVM Node: ACK
Note right of Avail: How do we read that data \nis being written properly?
EVM Node->Customer: ACK
Note right of EVM Node: Customer successfully notified
```

## High Level Plan - Tasks to work on

- Evaluate base technology to build on
    - `minigeth` -> Seems like the best shot right now.
    - Polygon Edge -> Meet on Wednesday.
    - Anything else?
- Build small PoC pieces of code to validate our evaluation.
    - Deploy a contract on EVM.
    - Execute a transaction.
    - Load & Store the account state.
    - Load & Store block from Avail.
- Figure out blockchain bootstrap process
    - How to generate the genesis block in our case.
- Figure out the block production process
    - Right now BABE looks like a good option for this.
    - :exclamation: Try to avoid single-node sequencer design from the start.
- Block finalization process shall wait for a bit later stage.
    - GRANDPA'ish seems like a way to go, but do we absolutely need it?
- 

## Requirements

- EVM compatible smart contracts for inter-application communication.
- Polygon Avail as Data Availability solution.

## :question: Questions to answer - Problems to solve

- How to finalize a block?
- How to manage blockchain upgrades?
    - i.e. block structure changes, process changes, node upgrades...


## :game_die: Random Notes

Here are some random notes while studying / researching through various components:

- Avail kind of consensus
    - Blockchain can fork; Honest node follows the right path.
- Account Data should probably be stored out-of-chain / out-of-Avail.
    - Bootstrapping new node from scratch is easier when a snapshot of state can be shipped and it can then catch up the rest by replaying from Avail.
    - If Avail had all state, we'd need to replay a long tail of history in order to catch with present day -> inconvenient.
- We can probably use [GSRPC](https://github.com/centrifuge/go-substrate-rpc-client) for Avail JSON-RPC connection.

## :warning: Threat Modeling

Here are some notes along the project, to take into account from threat modelling perspective.

- DDoS - Somehow able to predict the next block generator and stalling the progress, similar to Solana incident?
