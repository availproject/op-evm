#!/usr/bin/env bash

USER="ubuntu"
GROUP="ubuntu"
WORKSPACE="/home/$USER/workspace"
mkdir "$WORKSPACE" -p

wget https://github.com/maticnetwork/avail/releases/download/v1.3.0-rc3/data-avail-linux-aarch64.tar.gz
tar -xzvf data-avail-linux-aarch64.tar.gz -C "$WORKSPACE"

cat > /etc/systemd/system/avail.service <<EOF
[Unit]
Description=Avail DA Node
After=network.target

[Service]
Type=simple
Restart=always
RestartSec=10
ExecStart=/home/ubuntu/workspace/data-avail-linux-aarch64 --dev --ws-external --rpc-external --rpc-cors all

[Install]
WantedBy=multi-user.target
EOF

# Change ownership for the workspace
chown -R $USER:$GROUP "$WORKSPACE"

# Enable and start the service
systemctl daemon-reload
systemctl start avail
systemctl enable avail
