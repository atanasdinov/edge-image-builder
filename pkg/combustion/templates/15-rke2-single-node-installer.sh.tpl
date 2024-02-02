#!/bin/bash
set -euo pipefail

mount /var

mkdir -p /var/lib/rancher/rke2/agent/images/
cp {{ .imagesPath }}/* /var/lib/rancher/rke2/agent/images/

{{- if .vipManifest }}
mkdir -p /var/lib/rancher/rke2/server/manifests/
cp {{ .vipManifest }} /var/lib/rancher/rke2/server/manifests/{{ .vipManifest }}
{{- end }}

{{- if .manifestsPath }}
mkdir -p /var/lib/rancher/rke2/server/manifests/
cp {{ .manifestsPath }}/* /var/lib/rancher/rke2/server/manifests/
{{- end }}

umount /var

{{- if and .apiVIP .apiHost }}
echo "{{ .apiVIP }} {{ .apiHost }}" >> /etc/hosts
{{- end }}

mkdir -p /etc/rancher/rke2/
cp {{ .configFile }} /etc/rancher/rke2/config.yaml

{{- if .manifestsPath }}
cp {{ .registryMirrors }} /etc/rancher/rke2/registries.yaml
{{- end }}

export INSTALL_RKE2_ARTIFACT_PATH={{ .installPath }}

mount /usr/local
./rke2_installer.sh
umount /usr/local

systemctl enable rke2-server.service
