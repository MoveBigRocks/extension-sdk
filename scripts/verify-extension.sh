#!/usr/bin/env bash
set -euo pipefail

: "${MBR_URL:?Set MBR_URL to your Move Big Rocks instance URL}"
: "${MBR_EXTENSION_ID:?Set MBR_EXTENSION_ID to the installed extension ID}"

mbr auth whoami --url "${MBR_URL}" >/dev/null
mbr extensions validate "${MBR_EXTENSION_ID}" --url "${MBR_URL}"
mbr extensions activate "${MBR_EXTENSION_ID}" --url "${MBR_URL}"
mbr extensions monitor --id "${MBR_EXTENSION_ID}" --url "${MBR_URL}" --json
