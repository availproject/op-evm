# Optimistic EVM Rollup

OpEVM is a blockchain settlement system designed for efficient and secure transaction processing. It provides a decentralized infrastructure for settlement and enables high-throughput, low-latency transaction processing on the blockchain. OpEVM is built on top of the Avail network, extending [Polygon Edge](https://github.com/0xPolygon/polygon-edge) and offers advanced features for block validation, fraudproof detection, and transaction verification.

## Features

- Block Validation: OpEVM ensures that incoming blocks conform to the specified structure and contain valid transaction data.
- Fraudproof Detection: The system is equipped with fraudproof detection mechanisms to identify and handle malicious blocks.
- Transaction Verification: OpEVM verifies the integrity and correctness of transactions, ensuring the accuracy of settlement processes.
- High Throughput: The system is optimized for high throughput, enabling fast and efficient transaction processing.
- Low Latency: OpEVM minimizes transaction confirmation times, reducing latency and enabling near-instantaneous settlement.
- Decentralized Infrastructure: OpEVM operates on a decentralized infrastructure, ensuring the security and resilience of the settlement system.
- Optimistic Rollup: OpEVM is built as an optimistic rollup solution, leveraging Layer 2 scalability techniques to achieve high transaction throughput while maintaining the security guarantees of the underlying blockchain.


## Components

### Bootstrap Sequencer

The Bootstrap Sequencer component is responsible for bootstrapping the OpEVM. It initializes the necessary parameters, establishes the initial block structure, and sets up the initial state of the system.

### Sequencer

The Sequencer component is the main transaction processor in OpEVM. It receives incoming transactions, orders them, and includes them in the blocks to be added to the blockchain. The Sequencer plays a crucial role in maintaining the integrity and consistency of the OpEVM rollup.

### WatchTower

The WatchTower component is responsible for block validation, fraudproof detection, and transaction verification. It ensures the integrity of incoming blocks and identifies potential fraud or malicious activities.

### Staking

The Staking component handles the staking mechanisms within OpEVM. It manages stakeholder addresses, tracks staked amounts, and facilitates dispute resolution processes.


## Getting Started

### Demo Deployment

For detailed instructions on installation, configuration, and usage, refer to the [DEMO DEPLOYMENT](/docs/demo-deployment.md).

### Devnet/Testnet Deployment

To deploy a devnet or a testnet in AWS using terraform follow the instructions [here](/deployment/readme.md).

## Testing Fraudproof

Testing fraudproof processing is relatively straightforward. Sequencer implementation contains so called fraud server, which provides an HTTP interface which can be used to trigger a one time fraud construction into next produced block. Watchtower will then catch this and produce a fraudproof block, which leads to dispute resolution process.

### To test fraudproof, perform following actions:

1. Run at least one sequencer with fraud server enabled
   - To do this, run sequencer with `--fraud-srv-listen-addr "<address>:<port>"` (e.g.: `op-evm server --fraud-srv-listen-addr ":9990"`)
2. Run at least one watchtower that has staked
3. Optionally `tail` rollup blocks from Avail to easily see the process:
   - Run `op-evm tail --jsonrpc-addr "<sequencer's JSON-RPC address>"`
4. Make an HTTP request to fraud server:
   - e.g. `curl http://localhost:9990/fraud/prime`

The "malicius" sequencer will try to inject a begin dispute resolution transaction into the block, without chain inclusion, and watchtower will catch this and slash the sequencer.

Everytime the "malicious" sequencer has produced one fraudulent block, it will resume normal operation, until the fraud server has been _primed_ again.

## Contributing

We welcome contributions to the OpEVM project. If you find any issues, have suggestions for improvements, or would like to contribute new features, please open a GitHub issue or submit a pull request.

## Contributors

OpEVM was built as a joint project by two Equilibrium Group companies, Equilibrium Labs & Eiger and Avail.

[Avail](https://www.availproject.org/) creates the base layer for future blockchains, enabling developers to build rollups and appchains with scalability, flexibility, and ease.

[Equilibrium Group](https://www.eqg.co/), a blockchain powerhouse founded in 2018, is composed of three entities:
- [Equilibrium Labs](https://equilibrium.co/) designs & builds decentralized infrastructure in collaboration with industry pioneers and as in-house ventures.
- [Eiger](https://www.eiger.co/) provides high-value add engineering services to accelerate web3 mass adoption.
- Membrane Finance is the issuer of [EUROe](https://www.euroe.com/), the first EU-regulated euro-backed stablecoin.


## License

OpEVM is released under the Apache 2.0 License. See the [LICENSE](LICENSE) file for more details.