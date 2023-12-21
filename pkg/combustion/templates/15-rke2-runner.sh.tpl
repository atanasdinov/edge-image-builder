#!/bin/bash
set -euo pipefail

# Create a systemd unit that runs the RKE2 service
cat <<- EOF > /etc/systemd/system/rke2_runner.service
[Unit]
Description=Enable RKE2
Wants=network-online.target
After=network.target network-online.target
ConditionPathExists=|/opt/rke2/bin/rke2
ConditionPathExists=|/usr/bin/rke2
ConditionPathExists=|/usr/local/bin/rke2

[Service]
User=root
Type=oneshot
TimeoutStartSec=600
ExecStart=/bin/sh -c "systemctl enable --now rke2-{{ .Type }}.service"

RemainAfterExit=yes
KillMode=process

# Disable & cleanup
ExecStartPost=/bin/sh -c "systemctl disable rke2_runner"
ExecStartPost=rm -f /etc/systemd/system/rke2_runner.service

[Install]
WantedBy=multi-user.target
EOF

systemctl enable rke2_runner.service
