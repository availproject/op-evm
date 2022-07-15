# Polygon Avail Settlement

Polygon Avail Settlement provides a settlement layer for Polygon Avail.


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