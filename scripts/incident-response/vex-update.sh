#!/usr/bin/env bash
# scripts/incident-response/vex-update.sh
# Generates and cryptographically signs an updated OpenVEX statement
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/../.." && pwd)"

SERVICE="${1:-secure-app}"
CVE="${2:-CVE-2026-111497}"
IMAGE_DIGEST="${3:-}"
ORG="${ORG:-myorg}"

if [ -z "$IMAGE_DIGEST" ]; then
  if [ -f "/tmp/image-digest.txt" ]; then
    IMAGE_DIGEST="$(cat /tmp/image-digest.txt)"
  else
    echo "Usage: $0 <service_name> <cve_id> <sha256:digest>"
    exit 1
  fi
fi

PRODUCT_PURL="pkg:oci/${ORG}/${SERVICE}@${IMAGE_DIGEST}"
OUTPUT_VEX="${ROOT_DIR}/vex/openvex.json"

echo "=== Generating Updated OpenVEX Statement ==="
if command -v vexctl >/dev/null 2>&1; then
  vexctl create \
    --author="security-bot@myorg.com" \
    --product="${PRODUCT_PURL}" \
    --vuln="${CVE}" \
    --status=not_affected \
    --justification=vulnerable_code_not_in_execute_path \
    --impact="Brotli decompression is disabled; HTTP Accept-Encoding limited to identity." \
    > "${OUTPUT_VEX}"
else
  echo "Using built-in vex-generator..."
  cd "${ROOT_DIR}/vex/vex-generator"
  go run main.go \
    -product="${PRODUCT_PURL}" \
    -vuln="${CVE}" \
    -status="not_affected" \
    -justification="vulnerable_code_not_in_execute_path" \
    -impact="Brotli decompression is disabled; HTTP Accept-Encoding limited to identity." \
    -out="${OUTPUT_VEX}"
fi

echo "=== Attesting Updated VEX with Cosign Keyless ==="
cosign attest \
  --yes \
  --predicate "${OUTPUT_VEX}" \
  --type openvex \
  "ghcr.io/${ORG}/${SERVICE}@${IMAGE_DIGEST}"

echo "✅ VEX updated, attested, and signed for ${SERVICE}"
