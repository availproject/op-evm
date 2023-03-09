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

function generate_config {
    grpc_port=$1
    jsonrpc_port=$2
    p2p_port=$3
    node_type=$4

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
    libp2p_addr: :${p2p_port}
    nat_addr: ""
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
    cat <<EOF
[Unit]
Description=
After=network.target

[Service]
Type=simple
ExecStart=/home/ubuntu/server -config-file=/home/ubuntu/config.yaml -avail-addr ws://${avail_addr}:9944/v1/json-rpc

[Install]
WantedBy=multi-user.target

EOF
}
mkdir configs
terraform output --raw ssh_pk > ./configs/id_rsa
chmod 400 ./configs/id_rsa

all_instances=$(terraform output --json all_instances)

avail_addr=

while read -r i; do
    function node {
      echo "$i" | jq -r "$1"
    }
    if [ "$(node .tags_all.NodeType)" = "avail" ]; then
      avail_addr=$(node .public_dns)
      continue
    fi
    id=$(node .id)
    mkdir -p "configs/${id}"
    ../../third_party/polygon-edge/polygon-edge secrets init \
      --data-dir "./configs/${id}/data" \
      --json > "./configs/${id}/secrets.json"

done < <(echo "$all_instances" | jq -c '.[]')

while read -r i; do
    function node {
      echo "$i" | jq -r "$1"
    }

    if [ "$(node .tags_all.NodeType)" = "avail" ]; then
      # connect to the avail node and startup the avail service in dev mode
      scp -oStrictHostKeyChecking=no -i ./configs/id_rsa "./avail.service" "ubuntu@$(node .public_dns):/home/ubuntu/"
      ssh -n -i ./configs/id_rsa "ubuntu@$(node .public_dns)" 'wget https://github.com/maticnetwork/avail/releases/download/v1.4.0-rc3/data-avail-linux-aarch64.tar.gz && tar -xzvf data-avail-linux-aarch64.tar.gz && sudo mv avail.service /etc/systemd/system/ && sudo systemctl daemon-reload && sudo systemctl start avail.service && sudo systemctl enable avail.service'
      continue
    fi

    id=$(node .id)
    bootnode_json=""
    while read -r j; do
      function node2 {
        echo "$j" | jq -r "$1"
      }

      if [ "$id" = "$(node2 .id)" ]; then
        continue
      fi

      if [ "$(node2 .tags.NodeType)" = "bootstrap-sequencer"  ] || [ "$(node2 .tags.NodeType)" = "sequencer"  ]; then
        bootnode_json=$j
        break
      fi
    done < <(echo "$all_instances" | jq -c '.[]')

    if [ -z "$bootnode_json" ]; then
      echo "Boot node not found for instance '$id'; exiting!"
      exit 1
      continue
    fi
    function bootnode() {
        echo "$bootnode_json" | jq -r "$1"
    }

    ../../third_party/polygon-edge/polygon-edge genesis \
      --dir "./configs/$id/genesis.json" \
      --name "polygon-avail-settlement" \
      --premine "0x064A4a5053F3de5eacF5E72A2E97D5F9CF55f031:1000000000000000000000" \
      --consensus "ibft" \
      --bootnode "/ip4/$(bootnode .public_ip)/tcp/$(bootnode .tags_all.P2PPort)/p2p/$(cat "./configs/$(bootnode .id)/secrets.json" | jq -r .[0].node_id)" \
      --ibft-validator "0x1bC763b9c36Bb679B17Fc9ed01Ec5e27AF145864" \
      --ibft-validator-type "ecdsa"

    ../../third_party/polygon-edge/polygon-edge genesis predeploy \
      --chain "./configs/$id/genesis.json" \
      --predeploy-address "0x0110000000000000000000000000000000000001" \
      --artifacts-path "../../third_party/avail-settlement-contracts/staking/artifacts/contracts/Staking.sol/Staking.json" \
      --constructor-args "1" \
      --constructor-args "10"

    # cannot use -i flag because of macos machines
    sed 's/\"balance\": \"0x0\"/\"balance\": \"0x3635c9adc5dea00000\"/g' < "./configs/$id/genesis.json" > "./configs/$id/genesis2.json"
    mv "./configs/$id/genesis2.json" "./configs/$id/genesis.json"

    generate_config "$(node .tags_all.GRPCPort)" "$(node .tags_all.JsonRPCPort)" "$(node .tags_all.P2PPort)" "$(node .tags_all.NodeType)" > "./configs/$id/config.yaml"
    generate_service "$avail_addr" > "./configs/$id/node.service"

    scp -oStrictHostKeyChecking=no -i ./configs/id_rsa -r "./configs/$id/." "ubuntu@$(node .public_dns):/home/ubuntu/"
    scp -oStrictHostKeyChecking=no -i ./configs/id_rsa "../../server/server" "ubuntu@$(node .public_dns):/home/ubuntu/"

    ssh -n -i ./configs/id_rsa "ubuntu@$(node .public_dns)" 'sudo mv node.service /etc/systemd/system/ && sudo systemctl daemon-reload && sudo systemctl start node.service && sudo systemctl enable node.service'

done < <(echo "$all_instances" | jq -c '.[]')

# copy and startup avail