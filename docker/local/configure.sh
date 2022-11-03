#!/bin/sh

if [ -z $1 ]; then
	echo "usage: configure.sh <node id>"
	exit 1
fi

NODE_ID=$1

# Generate secrets
/polygon-edge secrets init --data-dir /data/avail-$NODE_ID

# Capture P2P ID
P2P_ID=$(/p2pkeytoid /data/avail-$NODE_ID/libp2p/libp2p.key)

# Append to genesis.json
jq ".bootnodes += [\"/dns4/$NODE_ID/tcp/10001/p2p/$P2P_ID\"]" /configs/genesis.json > /configs/genesis.json.new && mv /configs/genesis.json.new /configs/genesis.json

