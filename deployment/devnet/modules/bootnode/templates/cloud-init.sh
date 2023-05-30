#!/usr/bin/env bash

apt-get update && apt-get install -y awscli unzip

GITHUB_TOKEN=$(aws ssm get-parameter --name "${github_token_ssm_parameter_path}" --region ${region} --query "Parameter.Value" --output text --with-decryption)

curl -H "Authorization: token $GITHUB_TOKEN" -H "Accept:application/octet-stream" -L -o avail_settlement_artifact.zip "${avail_settlement_artifact_url}"
unzip -o avail_settlement_artifact.zip -d "${workspace}"

curl -H "Authorization: token $GITHUB_TOKEN" -H "Accept:application/octet-stream" -L -o accounts_artifact.zip "${accounts_artifact_url}"
unzip -o accounts_artifact.zip -d "${workspace}"

echo "${config_yaml_base64}" | base64 -d > "${workspace}/config.yaml"
echo "${secrets_config_json_base64}" | base64 -d > "${workspace}/secrets-config.json"
echo "${avail_settlement_service_base64}" | base64 -d > "/etc/systemd/system/avail-settlement.service"

aws s3 cp "s3://${s3_bucket_name}/genesis.json" "${workspace}"

# Deposit some tokens in the avail blockchain
${workspace}/accounts -balance 6 -avail-addr "ws://${avail_addr}:9944/v1/json-rpc" -path "${workspace}/account-mnemonic" --retry

sudo chown -R ${user}. "${workspace}"
sudo systemctl start avail-settlement
sudo systemctl enable avail-settlement
