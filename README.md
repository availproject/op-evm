# Polygon Avail Settlement

Polygon Avail Settlement provides a settlement layer for Polygon Avail.


# Node Secrets

```
/home/cortex/go/bin/polygon-edge secrets init --data-dir ./data/bootnode-1

[SECRETS INIT]
Public key (address) = 0x256C74Fe7639aA3dAdC6269cECdfC8d86C19a149
Node ID              = 16Uiu2HAmLRftAwcbtdhkVHkP9N81vhKUHR5X7yigEhdvB87whpMX

/home/cortex/go/bin/polygon-edge secrets init --data-dir ./data/node-1

[SECRETS INIT]
Public key (address) = 0x96791061710ECe4BAD6a0a0356144Cb5F406D710
Node ID              = 16Uiu2HAm8jiypDWerBbCruwfrA1SjvLRDVPwjrAnLLh5jbK3yxn3

/home/cortex/go/bin/polygon-edge secrets init --data-dir ./data/node-2

[SECRETS INIT]
Public key (address) = 0xFAED89Be49a8885F413DC73201DD560bBD9D8215
Node ID              = 16Uiu2HAmHDV7vW76xc8D9xgLgVjEP8Kc7yndVATBf2g5BhqfEJUT

```

## Bootnode

```
/ip4/127.0.0.1/tcp/10001/p2p/16Uiu2HAmLRftAwcbtdhkVHkP9N81vhKUHR5X7yigEhdvB87whpMX
```


# Server startup (polygon-edge)

## Bootnode (seal)
```
polygon-edge server --data-dir ./data/bootnode-1 --chain ./configs/genesis.json --grpc-address :10000 --libp2p :10001 --json
rpc :10002 --seal
```

```
go build && ./avail-settlement -config-file="./configs/bootnode.yaml"
```

## Node 1 (non seal)
```
polygon-edge server --data-dir ./data/node-1 --chain ./configs/genesis.json --grpc-address :20000 --libp2p :20001 --json
rpc :20002 --seal
```

```
go build && ./avail-settlement -config-file="./configs/node-1.yaml"
```
