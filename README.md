# Polygon Avail Settlement

Polygon Avail Settlement provides a settlement layer for Polygon Avail.


# Run instructions

0.) Requirements

- Avail node needs to be present and running locally.
- Need to clone and execute following make file commands from the root branch directory.

1.) Start sequencer and validator nodes. 

Have two tabs, and run them one by another, fast. 

```
make start-sequencer
make start-validator
```

2.) Open third tab and run the E2E test

```
make start-e2e
```

----

If all is produced well you will see following:

```
cortex@rij01-data01:~/eq/settlement$ make start-e2e
cd tools/e2e && go build -o e2e
./tools/e2e/e2e
2022/08/04 12:24:47 Current Headers -> Sequencer: 3269 | Validator: 3269 | Synced: true 
2022/08/04 12:24:47 Genesis Account Hex: 0x064A4a5053F3de5eacF5E72A2E97D5F9CF55f031 | Test Account Hex: 0x65F0bDe66C970F391bd648B7ea22e1c193221c65
2022/08/04 12:24:47 Sequencer Balances -> Genesis: 989 | Test: 11 
2022/08/04 12:24:47 Validator Balances -> Genesis: 989 | Test: 11 
2022/08/04 12:24:47 Initial balances are matching between sequencer and validator nodes!
2022/08/04 12:24:47 Sequencer -> initiating transfer of 1 ETH from genesis to test account...
2022/08/04 12:24:48 Sequencer -> genesis to test account 1 ETH transfer success! Tx hash: 0xeae656cc94e5e7bef5b27cd90cb8fbd7b7fb421daf9938f1386d717a5cfda187 
2022/08/04 12:24:48 Sequencer -> Awaiting for balance confirmation. Target -> Genesis: 988 | Test: 12
2022/08/04 12:24:53 Sequencer -> Ticker balance check -> Genesis Account: 989 | Test Account: 11 
2022/08/04 12:24:53 Sequencer -> Balances not matching yet... Waiting 5 seconds and rechecking...
2022/08/04 12:24:58 Sequencer -> Ticker balance check -> Genesis Account: 989 | Test Account: 11 
2022/08/04 12:24:58 Sequencer -> Balances not matching yet... Waiting 5 seconds and rechecking...
2022/08/04 12:25:03 Sequencer -> Ticker balance check -> Genesis Account: 989 | Test Account: 11 
2022/08/04 12:25:03 Sequencer -> Balances not matching yet... Waiting 5 seconds and rechecking...
2022/08/04 12:25:08 Sequencer -> Ticker balance check -> Genesis Account: 989 | Test Account: 11 
2022/08/04 12:25:08 Sequencer -> Balances not matching yet... Waiting 5 seconds and rechecking...
2022/08/04 12:25:13 Sequencer -> Ticker balance check -> Genesis Account: 989 | Test Account: 11 
2022/08/04 12:25:13 Sequencer -> Balances not matching yet... Waiting 5 seconds and rechecking...
2022/08/04 12:25:18 Sequencer -> Ticker balance check -> Genesis Account: 989 | Test Account: 11 
2022/08/04 12:25:18 Sequencer -> Balances not matching yet... Waiting 5 seconds and rechecking...
2022/08/04 12:25:23 Sequencer -> Ticker balance check -> Genesis Account: 988 | Test Account: 12 
2022/08/04 12:25:23 Sequencer -> Balance transfer confirmation successful! Time took: 35.69459452s 
2022/08/04 12:25:23 Validator -> Starting transfer confirmation check...
2022/08/04 12:25:28 Validator -> Ticker balance check -> Genesis Account: 988 | Test Account: 12 
2022/08/04 12:25:28 Validator -> Balance transfer confirmation successful! Total time took: 40.695656088s 
2022/08/04 12:25:28 E2E BALANCE TEST SUCCESSFUL!
```

Each of these methods will run go build and produce sufficient binaries.
Precompiled wallets as well as the datasets necessary for nodes to run are already in place.


# Node Secrets

```
/home/cortex/go/bin/polygon-edge secrets init --data-dir ./data/avail-bootnode-1

[SECRETS INIT]
Public key (address) = 0x1bC763b9c36Bb679B17Fc9ed01Ec5e27AF145864
Node ID              = 16Uiu2HAmMNxPzdzkNmtV97e9Y7kvHWahpGysW2Mq7GdDCDFdAcZa

/home/cortex/go/bin/polygon-edge secrets init --data-dir ./data/avail-node-1

[SECRETS INIT]
Public key (address) = 0x00D916EFbEeDb102A4D235a1EB525Fa147E5588e
Node ID              = 16Uiu2HAkwyY1aXwC7o7nrUofsBXwUkxYwsj21LtE9jmhmTuei5mw

/home/cortex/go/bin/polygon-edge secrets init --data-dir ./data/avail-node-2

[SECRETS INIT]
Public key (address) = 0x2734E3c95E2dBD08363f5298247b30a186c48b82
Node ID              = 16Uiu2HAm1vufDoGrukYQaTCDDE642B45HFLnUo5J2QcPpTosuiYp

```

## Bootnode

```
/ip4/127.0.0.1/tcp/10001/p2p/16Uiu2HAmUUNRnZLKRitXN9waugxMeqLYZ6PnwA8iPoiLMqRVZwQf
```


# Server startup (polygon-edge)

## Bootnode (seal)
```
polygon-edge server --data-dir ./data/bootnode-1 --chain ./configs/genesis.json --grpc-address :10000 --libp2p :10001 --json
rpc :10002 --seal
```

```
make build && ./server/server -config-file="./configs/bootnode.yaml"
```

## Node (non seal)
```
polygon-edge server --data-dir ./data/node-1 --chain ./configs/genesis.json --grpc-address :20000 --libp2p :20001 --json
rpc :20002
```

```
make build && ./server/server -config-file="./configs/node-1.yaml"
```

# Client

```
cortex@rij01-data01:~/eq/settlement$ make build-client && ./client/client 
cd client && go build -o client
2022/06/29 15:09:54 client: &ethclient.Client{c:(*rpc.Client)(0xc0001be100)}
2022/06/29 15:09:54 Got the header number: 0
```