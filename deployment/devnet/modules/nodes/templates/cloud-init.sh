#!/usr/bin/env bash

apt-get update && apt-get install -y awscli unzip

GITHUB_TOKEN=$(aws ssm get-parameter --name "${github_token_ssm_parameter_path}" --region ${region} --query "Parameter.Value" --output text --with-decryption)

curl -H "Authorization: token $GITHUB_TOKEN" -H "Accept:application/octet-stream" -L -o op_evm_artifact.zip "${op_evm_artifact_url}"
unzip -o op_evm_artifact.zip -d "${workspace}"

echo "${config_yaml_base64}" | base64 -d > "${workspace}/config.yaml"
echo "${op_evm_service_base64}" | base64 -d > "/etc/systemd/system/op-evm.service"
echo "${genesis_json_base64}" | base64 -d > "${workspace}/genesis.json"

# Init node keystore
${workspace}/op-evm secrets init --insecure --data-dir ${workspace}/data

# Deposit some tokens in the avail blockchain
${workspace}/op-evm availaccount --balance=6 --avail-addr="ws://${avail_addr}/v1/json-rpc" --path="${workspace}/account-mnemonic" --retry

sudo chown -R ${user}. "${workspace}"
sudo systemctl start op-evm
sudo systemctl enable op-evm
