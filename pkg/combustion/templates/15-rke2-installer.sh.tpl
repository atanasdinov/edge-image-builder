#!/bin/bash
set -euo pipefail

{{ if .SELinux -}}
export RKE2_SELINUX=true
# SELinux is only supported via the RPM installation method
export INSTALL_RKE2_METHOD=rpm
{{- else -}}
export INSTALL_RKE2_METHOD=tar
# The default value (/usr/local) is not accessible at this point and mounting
# it would simply redirect the RKE2 installer script to use /opt/rke2 anyway
export INSTALL_RKE2_TAR_PREFIX=/opt/rke2
{{- end }}

{{- if .Type }}
export INSTALL_RKE2_TYPE={{ .Type }}
{{- end }}

{{- if .Channel }}
export INSTALL_RKE2_CHANNEL={{ .Channel }}
{{- end }}

{{- if .Version }}
export INSTALL_RKE2_VERSION={{ .Version }}
{{- end }}

{{- if .ConfigFile }}
mkdir -p /etc/rancher/rke2/
cp {{ .ConfigFile }} /etc/rancher/rke2/config.yaml
{{- end }}

./rke2_installer.sh

echo "export KUBECONFIG=/etc/rancher/rke2/rke2.yaml" >> ~/.bashrc
echo "export PATH=${PATH}:/var/lib/rancher/rke2/bin" >> ~/.bashrc

systemctl enable rke2-{{ or .Type "server" }}.service
