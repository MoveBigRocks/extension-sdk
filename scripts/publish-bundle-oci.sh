#!/usr/bin/env bash
set -euo pipefail

bundle=""
image=""
tag=""
extra_tags=()

while [[ $# -gt 0 ]]; do
  case "$1" in
    --bundle)
      bundle="$2"
      shift 2
      ;;
    --image)
      image="$2"
      shift 2
      ;;
    --tag)
      tag="$2"
      shift 2
      ;;
    --extra-tag)
      extra_tags+=("$2")
      shift 2
      ;;
    *)
      echo "unknown argument: $1" >&2
      exit 1
      ;;
  esac
done

if [[ -z "$bundle" ]]; then
  echo "missing --bundle" >&2
  exit 1
fi
if [[ -z "$image" ]]; then
  echo "missing --image" >&2
  exit 1
fi
if [[ -z "$tag" ]]; then
  echo "missing --tag" >&2
  exit 1
fi

oras push "${image}:${tag}" \
  --disable-path-validation \
  --artifact-type application/vnd.mbr.extension.bundle.v1+json \
  "${bundle}:application/vnd.mbr.extension.bundle.v1+json"

for extra_tag in "${extra_tags[@]}"; do
  oras tag "${image}:${tag}" "${extra_tag}"
done

digest="$(oras manifest fetch "${image}:${tag}" --descriptor | jq -r '.digest')"
echo "Published ${image}:${tag} (${digest})"

if [[ -n "${GITHUB_OUTPUT:-}" ]]; then
  {
    echo "image=${image}"
    echo "tag=${tag}"
    echo "digest=${digest}"
  } >> "${GITHUB_OUTPUT}"
fi
