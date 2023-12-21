#!/bin/bash
set -euo pipefail

mkdir -p /etc/rancher/rke2/

cat << EOF > /etc/rancher/rke2/config.yaml
cni:
  - {{ or .CNI "canal" }}
{{- if .EnableMultus }}
  - multus
{{- end }}

{{- if .ServerURL }}
server: {{ .ServerURL }}
{{- end }}

{{- if .Token }}
token: {{ .Token }}
{{- end }}

{{- if .ConfigMode }}
write-kubeconfig-mode: {{ .ConfigMode }}
{{- end }}

{{- if .SAN }}
tls-san:
{{- range .SAN }}
  - {{ . }}
{{- end }}
{{- end }}

{{- if .Taints }}
node-taint:
{{- range .Taints }}
  - {{ . }}
{{- end }}
{{- end }}

{{- if .Labels }}
node-label:
{{- range .Labels }}
  - {{ . }}
{{- end }}
{{- end }}

{{- if .Debug }}
debug: {{ .Debug }}
{{- end }}

{{- if .SELinux }}
selinux: {{ .SELinux }}
{{- end }}

{{- if .Disable }}
disable:
{{- range .Disable }}
  - {{ . }}
{{- end }}
{{- end }}
EOF
