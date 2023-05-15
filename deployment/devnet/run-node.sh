#!/usr/bin/env bash

./workspace/accounts -balance 5 -avail-addr "ws://${1}:9944/v1/json-rpc" -path "./workspace/account-mnemonic"
sudo chown ubuntu:ubuntu ./workspace/data
sudo mv ./workspace/node.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl start node.service
sudo systemctl enable node.service
