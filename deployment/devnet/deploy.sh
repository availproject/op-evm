#!/usr/bin/env bash

# Make sure the script is always executed from the directory where the script is located
pushd "$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )" || exit
function cleanup {
  rm -rf configs
  popd || exit
}
trap cleanup EXIT
dns_name=$(terraform output --raw dns_name)
avail_addr=$(terraform output --raw avail_addr)

#TODO make sure the needed binaries exist before running this script!

function remote_copy {
  scp -oStrictHostKeyChecking=no -oProxyCommand="sh -c \"aws ssm start-session --target %h --document-name AWS-StartSSHSession --parameters 'portNumber=%p'\"" -i ./configs/id_rsa -r "$2" "ubuntu@$1:$3"
}

function remote_exec {
  ssh -oStrictHostKeyChecking=no -oProxyCommand="sh -c \"aws ssm start-session --target %h --document-name AWS-StartSSHSession --parameters 'portNumber=%p'\"" -i ./configs/id_rsa "ubuntu@$1" "$2"
}

function generate_config {
    p2p_port=$3
    node_type=$4

    cat <<EOF
chain_config: /home/ubuntu/workspace/genesis.json
secrets_config: ""
data_dir: /home/ubuntu/workspace/data
block_gas_target: "0x0"
grpc_addr: :20001
jsonrpc_addr: :20002
telemetry:
    prometheus_addr: ""
network:
    no_discover: false
    libp2p_addr: 0.0.0.0:${p2p_port}
    nat_addr: ""
    dns_addr: "dns/${dns_name}"
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
    node_type=$2
    cat <<EOF
[Unit]
Description=
After=network.target

[Service]
Type=simple
ExecStart=/home/ubuntu/workspace/avail-settlement $( [ "$node_type" = "bootstrap-sequencer" ] && echo '-bootstrap' ) -config-file=/home/ubuntu/workspace/config.yaml -avail-addr ws://${avail_addr}:9944/v1/json-rpc -account-config-file="/home/ubuntu/workspace/account-mnemonic"
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

instance_ids=()
while read -r i; do
  node_type=$(echo "$i" | jq -r .tags.NodeType)
  instance_id=$(echo "$i" | jq -r .id)
  if [ "$node_type" != "avail" ]; then
    instance_ids+=("$instance_id")
  fi
done < <(echo "$all_instances" | jq -c '.[]')

bootnodes=()
addresses=()
while read -r i; do
    node_type=$(echo "$i" | jq -r .tags.NodeType)
    instance_id=$(echo "$i" | jq -r .id)
    p2p_port=$(echo "$i" | jq -r .tags_all.P2PPort)

    if [ "$node_type" = "avail" ]; then
      continue
    fi

    mkdir -p "configs/${instance_id}"
    secrets_json=$(polygon-edge secrets init --data-dir "./configs/${instance_id}/data" --json --insecure)

    addresses+=("$(echo "$secrets_json" | jq -r .[0].address)")
    if [ "$node_type" = "bootstrap-sequencer"  ] || [ "$node_type" = "sequencer"  ]; then
      bootnodes+=("/dns/$dns_name/tcp/$p2p_port/p2p/$(echo "$secrets_json" | jq -r .[0].node_id)")
    fi

    generate_config "$p2p_port" "$node_type" > "./configs/$instance_id/config.yaml"
    generate_service "$node_type" > "./configs/$instance_id/node.service"

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

for instance_id in "${instance_ids[@]}"
do
  remote_exec "$instance_id" "sudo systemctl stop node"
  remote_exec "$instance_id" "rm -rf /home/ubuntu/workspace/data"
  remote_copy "$instance_id" "./configs/$instance_id/." "/home/ubuntu/workspace"
  remote_copy "$instance_id" "./configs/genesis.json" "/home/ubuntu/workspace"
  remote_copy "$instance_id" "../../avail-settlement" "/home/ubuntu/workspace"
  remote_copy "$instance_id" "../../tools/accounts/accounts" "/home/ubuntu/workspace"
  remote_copy "$instance_id" "./run-node.sh" "/home/ubuntu/workspace"
  remote_exec "$instance_id" "./workspace/run-node.sh $avail_addr"
done
