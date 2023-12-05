# Optimistic EVM Rollup

OpEVM is a sovereign EVM-compatible optimistic rollup construction designed for efficient and secure transaction processing. It provides a decentralized infrastructure for running a layer-2 (L2) blockchain and enables high-throughput, low-latency transaction processing. OpEVM is built on top of the [Avail](https://www.availproject.org/) and offers advanced features for block validation, fraud-proof detection, and transaction verification.

## Features

- Sovereign: OpEVM is uniquely built to provide an working optimistic rollup design without access to a base layer which supports execution. This means there is no need for a smart contract to perform fraud-proof checks to determine the canonical state of the chain. OpEVM completely relies on the node operators to determine the state of the chain, making it completely sovereign, while still inheriting the security of the base layer. 
- Optimistic Rollup: OpEVM is built as an optimistic rollup solution, leveraging Layer 2 scalability techniques to achieve high transaction throughput while maintaining the security guarantees of the underlying blockchain.
- Block Validation: OpEVM ensures that incoming blocks conform to the specified structure and contain valid transaction data. It extends [Polygon Edge](https://github.com/0xPolygon/polygon-edge) framework as the blockchain engine.
- Security: OpEVM relies on honest minority assumption. Under the assumption there is a watchtower which catches invalid blocks and produces fraud-proof, the system inherits the security of the base layer. 
- Fraud-proof Detection: The system is equipped with fraud-proof detection mechanisms to identify and handle malicious blocks.
- Transaction Verification: OpEVM verifies the integrity and correctness of transactions, ensuring the accuracy of settlement processes.
- High Throughput: The system is optimized for high throughput, enabling fast and efficient transaction processing.
- Low Latency: OpEVM minimizes transaction confirmation times, reducing latency and enabling near-instantaneous settlement.
- Decentralized Infrastructure: OpEVM operates on a decentralized infrastructure, ensuring the security and resilience of the settlement system. This includes support for a decentralized sequencer set, creating way for much stronger decentralization and censorship-resistant properties. 


## Components

### Bootstrap Sequencer

The Bootstrap Sequencer component is responsible for bootstrapping the OpEVM. It initializes the necessary parameters, establishes the initial block structure, and sets up the initial state of the system.

### Sequencer

The Sequencer component is the main transaction processor in OpEVM. It receives incoming transactions, orders them, and includes them in the blocks to be added to the blockchain. The Sequencer plays a crucial role in maintaining the integrity and consistency of the OpEVM rollup.

### WatchTower

The WatchTower component is responsible for block validation, fraud-proof detection, and transaction verification. It ensures the integrity of incoming blocks and identifies potential fraud or malicious activities.

### Staking

The Staking component handles the staking mechanisms within OpEVM. It manages stakeholder addresses, tracks staked amounts, and facilitates dispute resolution processes.


## Getting Started

### Demo Deployment

For detailed instructions on installation, configuration, and usage, refer to the [DEMO DEPLOYMENT](/docs/demo.md).

### Devnet/Testnet Deployment

To deploy a devnet or a testnet in AWS using terraform follow the instructions [here](/deployment/readme.md).

## Testing Fraudproof

Testing fraud-proof processing is relatively straightforward. Sequencer implementation contains so called fraud server, which provides an HTTP interface which can be used to trigger a one time fraud construction into next produced block. Watchtower will then catch this and produce a fraud-proof block, which leads to dispute resolution process.

### To Evaluate the Fraud-proof Mechanism, Execute the Following Procedures:

1. Initiate a minimum of one sequencer with the fraud server function activated.
   - This can be accomplished by launching the sequencer with the `--fraud-srv-listen-addr "<address>:<port>"` command (for instance: `op-evm server --fraud-srv-listen-addr ":9990"`).
2. Initiate at least one watchtower that possesses staked assets.
3. Optionally, you may monitor rollup blocks from Avail to facilitate a clear understanding of the process. This can be done by executing the `op-evm tail --jsonrpc-addr "<sequencer's JSON-RPC address>"` command.
4. Generate an HTTP request to the fraud server:
   - For instance, `curl http://localhost:9990/fraud/prime`.

The sequencer, acting in a "malicious" capacity, will attempt to incorporate a begin dispute resolution transaction into the block, without chain inclusion. The watchtower will detect this action and penalize the sequencer accordingly.

Following the production of a fraudulent block by the "malicious" sequencer, normal operations will be resumed until the fraud server is _primed_ once more.

## Limitations

A list of limitations is present in the [issues](https://github.com/availproject/op-evm/issues). However, here are a few core limitations of this prototype:
- Transaction replay: If a valid fraud proof arrives and rolls back the chain, the transactions that are rolled back are not replayed. Transactions from invalid block are essentially invalidated without any notification to the user.
- Light client support: The prototype does not support execution light clients (LC), as it is expensive for LCs to re-execute blocks. However, in the future, we can produce validity proof of correct block execution as fraud proofs which will allow LCs to follow chain very easily. NOTE: This is not about DA light clients. The design does support Avail LCs and may be included in the prototype in a future iteration. 
- Bridging: The rollup uses native tokens for staking and user transactions. The operators use AVL to submit blocks to Avail. The Sovereign nature of the rollup does not allow a clear bi-directional bridge design from the rollup to another chain.
- Data compression: The prototype already implements state diffs propagation on the p2p in the optimistic case. However, the blocks that are posted on Avail are not compressed. There is a big scope for cost optimization there.

If you feel some important limitations are not covered, please check out our [Contributing section](https://github.com/availproject/op-evm#contributing) and open an issue or a PR. 

## Contributing

We welcome contributions to the OpEVM project. If you find any issues, have suggestions for improvements, or would like to contribute new features, please open a GitHub issue or submit a pull request.

## Contributors

OpEVM was built in collaboration between [Avail](https://www.availproject.org/) and [Equilibrium Group](https://www.eqg.co/) ([Equilibrium Labs](https://equilibrium.co/) & [Eiger](https://www.eiger.co/)).

[Avail](https://www.availproject.org/) creates the base layer for future blockchains, enabling developers to build rollups and appchains with scalability, flexibility, and ease.

[Equilibrium Group](https://www.eqg.co/), a blockchain powerhouse founded in 2018, is composed of three entities:
- [Equilibrium Labs](https://equilibrium.co/) designs & builds decentralized infrastructure in collaboration with industry pioneers and as in-house ventures.
- [Eiger](https://www.eiger.co/) provides high-value add engineering services to accelerate web3 mass adoption.
- Membrane Finance is the issuer of [EUROe](https://www.euroe.com/), the first EU-regulated euro-backed stablecoin.


## Warning
This is a prototype. It contains known vulnerabilities and missing essential features. It should not be used in production, in its current form, under any circumstances.

## License

OpEVM is released under the Apache 2.0 License. See the [LICENSE](LICENSE) file for more details.
