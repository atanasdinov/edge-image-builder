#!/bin/bash
set -euo pipefail

# Mount /usr/local to store the RKE2 script
mount /usr/local || true

curl -L --output rke2_installer.sh https://get.rke2.io && install -m755 rke2_installer.sh /usr/local/bin/

# Create a systemd unit that installs RKE2
cat <<- EOF > /etc/systemd/system/rke2_installer.service
[Unit]
Description=Install RKE2
Wants=network-online.target
After=network.target network-online.target
ConditionPathExists=/usr/local/bin/rke2_installer.sh
ConditionPathExists=!/opt/rke2/bin/rke2

[Service]
User=root
Type=forking
TimeoutStartSec=600

{{- if .Type }}
Environment="INSTALL_RKE2_TYPE={{ .Type }}"
{{- end }}
{{- if .Channel }}
Environment="INSTALL_RKE2_CHANNEL={{ .Channel }}"
{{- end }}
{{- if .Version }}
Environment="INSTALL_RKE2_VERSION={{ .Version }}"
{{- end }}
{{- if .Token }}
Environment="RKE2_TOKEN={{ .Token }}"
{{- end }}
{{- if .SELinux }}
Environment="RKE2_SELINUX=true"
Environment="INSTALL_RKE2_METHOD=rpm"
{{- end }}

ExecStart=/usr/local/bin/rke2_installer.sh

RemainAfterExit=yes
KillMode=process

# Update $PATH to include RKE2 binary
ExecStartPost=/bin/sh -c "echo 'export KUBECONFIG=/etc/rancher/rke2/rke2.yaml' >> ~/.bashrc; echo 'export PATH=${PATH}:/var/lib/rancher/rke2/bin' >> ~/.bashrc; source ~/.bashrc"

# Disable & cleanup
ExecStartPost=rm -f /usr/local/bin/rke2_installer.sh
ExecStartPost=/bin/sh -c "systemctl disable rke2_installer"
ExecStartPost=rm -f /etc/systemd/system/rke2_installer.service

# Reboot to enable RKE2 binaries and service files
ExecStartPost=/bin/sh -c "reboot"

[Install]
WantedBy=multi-user.target
EOF

systemctl enable rke2_installer.service
