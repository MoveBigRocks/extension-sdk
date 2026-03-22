#!/usr/bin/env bash
set -euo pipefail

EXTENSION_SOURCE="${1:-.}"

: "${MBR_URL:?Set MBR_URL to your Move Big Rocks instance URL}"
: "${MBR_WORKSPACE_ID:?Set MBR_WORKSPACE_ID to the sandbox workspace ID}"
: "${MBR_LICENSE_TOKEN:?Set MBR_LICENSE_TOKEN to the sandbox extension license token}"

mbr auth whoami --url "${MBR_URL}" >/dev/null
mbr extensions install "${EXTENSION_SOURCE}" --workspace "${MBR_WORKSPACE_ID}" --license-token "${MBR_LICENSE_TOKEN}" --url "${MBR_URL}"
