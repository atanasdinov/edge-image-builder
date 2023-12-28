#!/bin/bash
set -euo pipefail

# Mounting /usr/local is required by the RKE2 installer script
mount /usr/local || true

{{- if .Type }}
export INSTALL_RKE2_TYPE={{ .Type }}
{{- end }}
{{- if .Channel }}
export INSTALL_RKE2_CHANNEL={{ .Channel }}
{{- end }}
{{- if .Version }}
export INSTALL_RKE2_VERSION={{ .Version }}
{{- end }}
{{- if .Token }}
export RKE2_TOKEN={{ .Token }}
{{- end }}
{{- if .SELinux }}
export RKE2_SELINUX=true
export INSTALL_RKE2_METHOD=rpm
{{- end }}

./rke2_installer.sh

# Update $PATH to include RKE2 binary
echo 'export KUBECONFIG=/etc/rancher/rke2/rke2.yaml' >> ~/.bashrc
echo 'export PATH=${PATH}:/var/lib/rancher/rke2/bin' >> ~/.bashrc
source ~/.bashrc

systemctl enable rke2-{{ .Type }}.service
