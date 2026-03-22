#!/usr/bin/env bash
set -euo pipefail

EXTENSION_SOURCE="${1:-.}"

: "${MBR_URL:?Set MBR_URL to your Move Big Rocks instance URL}"
: "${MBR_EXTENSION_ID:?Set MBR_EXTENSION_ID to the installed extension ID}"

mbr auth whoami --url "${MBR_URL}" >/dev/null
mbr extensions upgrade "${EXTENSION_SOURCE}" --id "${MBR_EXTENSION_ID}" --url "${MBR_URL}"
