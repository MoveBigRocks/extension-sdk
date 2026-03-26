#!/usr/bin/env bash
set -euo pipefail

EXTENSION_SOURCE="${1:-.}"

: "${MBR_URL:?Set MBR_URL to your Move Big Rocks instance URL}"
: "${MBR_WORKSPACE_ID:?Set MBR_WORKSPACE_ID to the sandbox workspace ID}"

mbr auth whoami --url "${MBR_URL}" >/dev/null
install_args=(
  "${EXTENSION_SOURCE}"
  --workspace "${MBR_WORKSPACE_ID}"
  --url "${MBR_URL}"
)

if [[ -n "${MBR_LICENSE_TOKEN:-}" ]]; then
  install_args+=(--license-token "${MBR_LICENSE_TOKEN}")
fi

mbr extensions install "${install_args[@]}"
