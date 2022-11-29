#!/bin/sh

BOOTNODE=false

while getopts 'b' OPTION; do
	case "$OPTION" in
		b)
			BOOTNODE=true
			;;
		?)
			echo "usage: configure.sh [-b] <node id>"
			exit 1
	esac
done
shift "$(($OPTIND -1))"

NODE_ID=$1
if [ -z $NODE_ID ]; then
	echo "usage: configure.sh [-b] <node id>"
	exit 1
fi


# Generate secrets
/polygon-edge secrets init --data-dir /data/avail-$NODE_ID


if [ "$BOOTNODE" = true ]; then
	# Capture P2P ID
	P2P_ID=$(/p2pkeytoid /data/avail-$NODE_ID/libp2p/libp2p.key)

	# Append to genesis.json
	jq ".bootnodes += [\"/dns4/$NODE_ID/tcp/1478/p2p/$P2P_ID\"]" /configs/genesis.json > /configs/genesis.json.new && mv /configs/genesis.json.new /configs/genesis.json
fi
