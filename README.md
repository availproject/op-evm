# Avail Settlement Layer PoC

Avail SL is a blockchain settlement system designed for efficient and secure transaction processing. It provides a decentralized infrastructure for settlement and enables high-throughput, low-latency transaction processing on the blockchain. Avail SL is built on top of the Polygon network, extending Polygon Edge and offers advanced features for block validation, fraudproof detection, and transaction verification.

## Features

- Block Validation: Avail SL ensures that incoming blocks conform to the specified structure and contain valid transaction data.
- Fraudproof Detection: The system is equipped with fraudproof detection mechanisms to identify and handle malicious blocks.
- Transaction Verification: Avail SL verifies the integrity and correctness of transactions, ensuring the accuracy of settlement processes.
- High Throughput: The system is optimized for high throughput, enabling fast and efficient transaction processing.
- Low Latency: Avail SL minimizes transaction confirmation times, reducing latency and enabling near-instantaneous settlement.
- Decentralized Infrastructure: Avail SL operates on a decentralized infrastructure, ensuring the security and resilience of the settlement system.
- Optimistic Rollup: Avail SL is built as an optimistic rollup solution, leveraging Layer 2 scalability techniques to achieve high transaction throughput while maintaining the security guarantees of the underlying blockchain.


## Components

### Bootstrap Sequencer

The Bootstrap Sequencer component is responsible for bootstrapping the Avail settlement system. It initializes the necessary parameters, establishes the initial block structure, and sets up the initial state of the system.

### Sequencer

The Sequencer component is the main transaction processor in Avail. It receives incoming transactions, orders them, and includes them in the blocks to be added to the blockchain. The Sequencer plays a crucial role in maintaining the integrity and consistency of the Avail settlement system.

### WatchTower

The WatchTower component is responsible for block validation, fraudproof detection, and transaction verification. It ensures the integrity of incoming blocks and identifies potential fraud or malicious activities.

### Blockchain

The Blockchain component provides the underlying blockchain infrastructure for Avail. It stores and manages the blockchain data, including blocks, transactions, and state information. The consensus mechanism is based on the Avail database, ensuring secure and reliable settlement operations.

### Staking

The Staking component handles the staking mechanisms within Avail. It manages stakeholder addresses, tracks staked amounts, and facilitates dispute resolution processes.

## Settlement Layer

The settlement layer in Avail operates on an optimistic model, enabling efficient and rapid transaction processing. It leverages advanced techniques and algorithms to ensure high-throughput settlement while maintaining data integrity and security.

## Getting Started

To get started with Avail, follow these steps:

1. Install the required dependencies and libraries.
2. Set up the Avail environment by configuring the blockchain connection and network parameters.
3. Deploy and initialize the Bootstrap Sequencer and Sequencer components.
4. Connect to the blockchain network and start processing incoming transactions.
5. Monitor the system for fraudproof detections and handle dispute resolution processes when necessary.

For detailed instructions on installation, configuration, and usage, refer to the TODO.

## Testing Fraudproof

Testing fraudproof processing is relatively straightforward. Sequencer implementation contains so called fraud server, which provides an HTTP interface which can be used to trigger a one time fraud construction into next produced block. Watchtower will then catch this and produce a fraudproof block, which leads to dispute resolution process.

### To test fraudproof, perform following actions:

1. Run at least one sequencer with fraud server enabled
   - To do this, run sequencer with `--fraud-srv-listen-addr "<address>:<port>"` (e.g.: `avail-settlement server --fraud-srv-listen-addr ":9990"`)
2. Run at least one watchtower that has staked
3. Optionally `tail` Settlement Layer blocks from Avail to easily see the process:
   - Run `avail-settlement tail --jsonrpc-addr "<sequencer's JSON-RPC address>"`
4. Make an HTTP request to fraud server:
   - e.g. `curl http://localhost:9990/fraud/prime`

The "malicius" sequencer will try to inject a begin dispute resolution transaction into the block, without chain inclusion, and watchtower will catch this and slash the sequencer.

Everytime the "malicious" sequencer has produced one fraudulent block, it will resume normal operation, until the fraud server has been _primed_ again.

## Contributing

We welcome contributions to the Avail project. If you find any issues, have suggestions for improvements, or would like to contribute new features, please open a GitHub issue or submit a pull request.

## License

TODO

## Contact

For any inquiries or questions, please contact the Avail SL development team at TODO
