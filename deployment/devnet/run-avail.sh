#!/usr/bin/env bash

wget https://github.com/maticnetwork/avail/releases/download/v1.3.0-rc3/data-avail-linux-aarch64.tar.gz
tar -xzvf data-avail-linux-aarch64.tar.gz -C ./workspace
sudo mv ./workspace/avail.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl start avail.service
sudo systemctl enable avail.service
