#!/usr/bin/env bash
set -euo pipefail

: "${MBR_URL:?Set MBR_URL to your Move Big Rocks instance URL}"

if [[ -n "${MBR_EXTENSION_SOURCE_DIR:-}" ]]; then
  mbr auth whoami --url "${MBR_URL}" >/dev/null
  verify_args=(
    "${MBR_EXTENSION_SOURCE_DIR}"
    --url "${MBR_URL}"
  )
  if [[ -n "${MBR_WORKSPACE_ID:-}" ]]; then
    verify_args+=(--workspace "${MBR_WORKSPACE_ID}")
  fi
  if [[ -n "${MBR_LICENSE_TOKEN:-}" ]]; then
    verify_args+=(--license-token "${MBR_LICENSE_TOKEN}")
  fi
  mbr extensions verify "${verify_args[@]}" --json
  exit 0
fi

: "${MBR_EXTENSION_ID:?Set MBR_EXTENSION_ID to the installed extension ID or set MBR_EXTENSION_SOURCE_DIR}"

mbr auth whoami --url "${MBR_URL}" >/dev/null
mbr extensions validate --id "${MBR_EXTENSION_ID}" --url "${MBR_URL}"
mbr extensions activate --id "${MBR_EXTENSION_ID}" --url "${MBR_URL}"
mbr extensions show --id "${MBR_EXTENSION_ID}" --url "${MBR_URL}" --json
mbr extensions monitor --id "${MBR_EXTENSION_ID}" --url "${MBR_URL}" --json

if [[ -n "${MBR_WORKSPACE_ID:-}" ]]; then
  mbr extensions nav --workspace "${MBR_WORKSPACE_ID}" --url "${MBR_URL}" --json
  mbr extensions widgets --workspace "${MBR_WORKSPACE_ID}" --url "${MBR_URL}" --json
fi
