#!/usr/bin/env bash

apt-get update && apt-get install -y awscli unzip

GITHUB_TOKEN=$(aws ssm get-parameter --name "${github_token_ssm_parameter_path}" --region ${region} --query "Parameter.Value" --output text --with-decryption)

curl -H "Authorization: token $GITHUB_TOKEN" -H "Accept:application/octet-stream" -L -o avail_settlement_artifact.zip "${avail_settlement_artifact_url}"
unzip -o avail_settlement_artifact.zip -d "${workspace}"

echo "${config_yaml_base64}" | base64 -d > "${workspace}/config.yaml"
echo "${avail_settlement_service_base64}" | base64 -d > "/etc/systemd/system/avail-settlement.service"
echo "${genesis_json_base64}" | base64 -d > "${workspace}/genesis.json"

# Init node keystore
${workspace}/avail-settlement secrets init --insecure --data-dir ${workspace}/data

# Deposit some tokens in the avail blockchain
${workspace}/avail-settlement availaccount --balance=6 --avail-addr="ws://${avail_addr}/v1/json-rpc" --path="${workspace}/account-mnemonic" --retry

sudo chown -R ${user}. "${workspace}"
sudo systemctl start avail-settlement
sudo systemctl enable avail-settlement
