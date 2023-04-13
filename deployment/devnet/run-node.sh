#!/usr/bin/env bash

./accounts -balance 1000 -avail-addr "ws://${1}:9944/v1/json-rpc" -path "./account-mnemonic"
sudo chown ubuntu:ubuntu ./data
sudo mv node.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl start node.service
sudo systemctl enable node.service
