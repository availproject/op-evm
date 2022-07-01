# Polygon Avail Settlement

Polygon Avail Settlement provides a settlement layer for Polygon Avail.


# Node Secrets

```
/home/cortex/go/bin/polygon-edge secrets init --data-dir ./data/bootnode-1

[SECRETS INIT]
Public key (address) = 0xaDE7EE1e3b98A1Eb8E4D8E6c52fd073dB1a1304B
Node ID              = 16Uiu2HAmUUNRnZLKRitXN9waugxMeqLYZ6PnwA8iPoiLMqRVZwQf

/home/cortex/go/bin/polygon-edge secrets init --data-dir ./data/node-1

[SECRETS INIT]
Public key (address) = 0x8A96444859cE76d3F13D044F3962cb25795F32D6
Node ID              = 16Uiu2HAmSg1YmxfdEBDgGBwmHW3xDnndo5ooPzEarbCtJmMhmiJZ

/home/cortex/go/bin/polygon-edge secrets init --data-dir ./data/node-2

[SECRETS INIT]
Public key (address) = 0x7201c9c9b1b56b09AB5e1013251270E8F5D023C5
Node ID              = 16Uiu2HAmUCgwmzn9WiLYvbZSD7zRdMKuhu2BAQnjF9T8H3WbqKM1

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