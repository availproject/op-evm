#!/usr/bin/env bash

# Make sure the script is always executed from the directory where the script is located
current_dir=$(pwd)
cd "$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )" || exit
function cleanup {
  rm -rf configs
  cd "$current_dir" || exit
}
trap cleanup EXIT
#TODO make sure the needed binaries exist before running this script!

function remote_copy {
  scp -oStrictHostKeyChecking=no -i ./configs/id_rsa -r "$2" "ubuntu@$1:$3"
}

function remote_exec {
  ssh -oStrictHostKeyChecking=no -i ./configs/id_rsa "ubuntu@$1" "$2"
}

function generate_config {
    grpc_port=$1
    jsonrpc_port=$2
    p2p_port=$3
    node_type=$4
    nat_addr=$5

    cat <<EOF
chain_config: /home/ubuntu/genesis.json
secrets_config: ""
data_dir: /home/ubuntu/data
block_gas_target: "0x0"
grpc_addr: :${grpc_port}
jsonrpc_addr: :${jsonrpc_port}
telemetry:
    prometheus_addr: ""
network:
    no_discover: false
    libp2p_addr: 0.0.0.0:${p2p_port}
    nat_addr: ${nat_addr}
    dns_addr: ""
    max_peers: 40
    max_outbound_peers: 8
    max_inbound_peers: 32
seal: true
tx_pool:
    price_limit: 0
    max_slots: 4096
    max_account_enqueued: 128
log_level: ERROR
restore_file: ""
block_time_s: 2
headers:
    access_control_allow_origins:
        - '*'
log_to: ""
json_rpc_batch_request_limit: 20
json_rpc_block_range_limit: 1000
node_type: ${node_type}
json_log_format: false

EOF
}

function generate_service {
    avail_addr=$1
    node_type=$2
    cat <<EOF
[Unit]
Description=
After=network.target

[Service]
Type=simple
ExecStart=/home/ubuntu/avail-settlement $( [ "$node_type" = "bootstrap-sequencer" ] && echo '-bootstrap' ) -config-file=/home/ubuntu/config.yaml -avail-addr ws://${avail_addr}:9944/v1/json-rpc -account-config-file="/home/ubuntu/account-mnemonic"
User=ubuntu
Group=ubuntu

[Install]
WantedBy=multi-user.target

EOF
}
mkdir configs
terraform output --raw ssh_pk > ./configs/id_rsa
chmod 400 ./configs/id_rsa

all_instances=$(terraform output --json all_instances)

avail_addr=
nodes_addr=()
while read -r i; do
  node_type=$(echo "$i" | jq -r .tags.NodeType)
  node_addr=$(echo "$i" | jq -r .public_dns)
  if [ "$node_type" = "avail" ]; then
        avail_addr=$node_addr
  else
    nodes_addr+=("$node_addr")
  fi
done < <(echo "$all_instances" | jq -c '.[]')

bootnodes=()
addresses=()
while read -r i; do
    node_type=$(echo "$i" | jq -r .tags.NodeType)
    public_ip=$(echo "$i" | jq -r .public_ip)
    node_addr=$(echo "$i" | jq -r .public_dns)
    p2p_port=$(echo "$i" | jq -r .tags_all.P2PPort)
    grpc_port=$(echo "$i" | jq -r .tags_all.GRPCPort)
    jsonrpc_port=$(echo "$i" | jq -r .tags_all.JsonRPCPort)

    if [ "$node_type" = "avail" ]; then
      continue
    fi

    mkdir -p "configs/${node_addr}"
    secrets_json=$(polygon-edge secrets init --data-dir "./configs/${node_addr}/data" --json --insecure)

    addresses+=("$(echo "$secrets_json" | jq -r .[0].address)")
    if [ "$node_type" = "bootstrap-sequencer"  ] || [ "$node_type" = "sequencer"  ]; then
      bootnodes+=("/ip4/$public_ip/tcp/$p2p_port/p2p/$(echo "$secrets_json" | jq -r .[0].node_id)")
    fi

    generate_config "$grpc_port" "$jsonrpc_port" "$p2p_port" "$node_type" "$public_ip" > "./configs/$node_addr/config.yaml"
    generate_service "$avail_addr" "$node_type" > "./configs/$node_addr/node.service"

done < <(echo "$all_instances" | jq -c '.[]')

# Modify the bootnodes field in genesis.json
genesis=$(jq ".bootnodes = []" < ../../configs/genesis.json)
for i in "${!bootnodes[@]}"
do
  genesis=$(echo "$genesis" | jq ".bootnodes[$i] = \"${bootnodes[$i]}\"")
done
for addr in "${addresses[@]}"
do
  genesis=$(echo "$genesis" | jq ".genesis.alloc[\"${addr}\"] = {\"balance\": \"0x3635c9adc5dea00000\"}")
done
echo "$genesis" > ./configs/genesis.json

remote_exec "$avail_addr" "sudo systemctl stop avail"
remote_exec "$avail_addr" "rm -rf ./data"
remote_copy "$avail_addr" "./run-avail.sh" "/home/ubuntu/"
remote_copy "$avail_addr" "./avail.service" "/home/ubuntu/"
remote_exec "$avail_addr" "./run-avail.sh"

sleep 10

for node_addr in "${nodes_addr[@]}"
do
  remote_exec "$node_addr" "sudo systemctl stop node"
  remote_exec "$node_addr" "rm -rf ./data"
  remote_copy "$node_addr" "./configs/$node_addr/." "/home/ubuntu/"
  remote_copy "$node_addr" "./configs/genesis.json" "/home/ubuntu/"
  remote_copy "$node_addr" "../../avail-settlement" "/home/ubuntu/"
  remote_copy "$node_addr" "../../tools/accounts/accounts" "/home/ubuntu/"
  remote_copy "$node_addr" "./run-node.sh" "/home/ubuntu/"
  remote_exec "$node_addr" "./run-node.sh $avail_addr"
done
