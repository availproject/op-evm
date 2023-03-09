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
bootnodes=()
nodes_addr=()
while read -r i; do
    node_type=$(echo "$i" | jq -r .tags.NodeType)
    public_ip=$(echo "$i" | jq -r .public_ip)
    node_addr=$(echo "$i" | jq -r .public_dns)
    p2p_port=$(echo "$i" | jq -r .tags_all.P2PPort)
    grpc_port=$(echo "$i" | jq -r .tags_all.GRPCPort)
    jsonrpc_port=$(echo "$i" | jq -r .tags_all.JsonRPCPort)

    if [ "$node_type" = "avail" ]; then
      avail_addr=$node_addr
      continue
    else
      nodes_addr+=("$node_addr")
    fi

    mkdir -p "configs/${node_addr}"
    node_id=$(../../third_party/polygon-edge/polygon-edge secrets init --data-dir "./configs/${node_addr}/data" --json | jq -r .[0].node_id)

    if [ "$node_type" = "bootstrap-sequencer"  ] || [ "$node_type" = "sequencer"  ]; then
      bootnodes+=("/ip4/$public_ip/tcp/$p2p_port/p2p/$node_id")
    fi
    echo "generating config for node: $node_type"
    generate_config "$grpc_port" "$jsonrpc_port" "$p2p_port" "$node_type" "$public_ip" > "./configs/$node_addr/config.yaml"
    generate_service "$avail_addr" > "./configs/$node_addr/node.service"

done < <(echo "$all_instances" | jq -c '.[]')

../../third_party/polygon-edge/polygon-edge genesis \
  --dir "./configs/genesis.json" \
  --name "polygon-avail-settlement" \
  --premine "0x064A4a5053F3de5eacF5E72A2E97D5F9CF55f031:1000000000000000000000" \
  "${bootnodes[@]/#/--bootnode=}" \
  --consensus "ibft" \
  --ibft-validator "0x1bC763b9c36Bb679B17Fc9ed01Ec5e27AF145864" \
  --ibft-validator-type "ecdsa"

../../third_party/polygon-edge/polygon-edge genesis predeploy \
  --chain "./configs/genesis.json" \
  --predeploy-address "0x0110000000000000000000000000000000000001" \
  --artifacts-path "../../third_party/avail-settlement-contracts/staking/artifacts/contracts/Staking.sol/Staking.json" \
  --constructor-args "1" \
  --constructor-args "10"

# cannot use -i flag because macos version of sed doesn't allow this flag
sed 's/\"balance\": \"0x0\"/\"balance\": \"0x3635c9adc5dea00000\"/g' < "./configs/genesis.json" > "./configs/genesis2.json"
mv "./configs/genesis2.json" "./configs/genesis.json"

scp -oStrictHostKeyChecking=no -i ./configs/id_rsa "./avail.service" "ubuntu@$avail_addr:/home/ubuntu/"
ssh -n -i ./configs/id_rsa "ubuntu@$avail_addr" 'wget https://github.com/maticnetwork/avail/releases/download/v1.4.0-rc3/data-avail-linux-aarch64.tar.gz && tar -xzvf data-avail-linux-aarch64.tar.gz && sudo mv avail.service /etc/systemd/system/ && sudo systemctl daemon-reload && sudo systemctl start avail.service && sudo systemctl enable avail.service'

for node_addr in "${nodes_addr[@]}"
do
  scp -oStrictHostKeyChecking=no -i ./configs/id_rsa -r "./configs/$node_addr/." "ubuntu@$node_addr:/home/ubuntu/"
  scp -oStrictHostKeyChecking=no -i ./configs/id_rsa -r "./configs/genesis.json" "ubuntu@$node_addr:/home/ubuntu/"
  scp -oStrictHostKeyChecking=no -i ./configs/id_rsa "../../server/server" "ubuntu@$node_addr:/home/ubuntu/"
  ssh -n -i ./configs/id_rsa "ubuntu@$node_addr" 'sudo mv node.service /etc/systemd/system/ && sudo systemctl daemon-reload && sudo systemctl start node.service && sudo systemctl enable node.service'
done
