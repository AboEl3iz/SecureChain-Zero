#!/usr/bin/env bash
# scripts/verify.sh — End-to-end supply chain trust verification
set -euo pipefail

IMAGE_DIGEST="${1:-}"
if [ -z "$IMAGE_DIGEST" ]; then
  if [ -f "/tmp/image-digest.txt" ]; then
    IMAGE_DIGEST="$(cat /tmp/image-digest.txt)"
  else
    echo "Usage: $0 <sha256:digest> [image_repo]"
    echo "Example: $0 sha256:abc123... ghcr.io/myorg/secure-app"
    exit 1
  fi
fi

IMAGE_REPO="${2:-ghcr.io/myorg/secure-app}"
IMAGE="${IMAGE_REPO}@${IMAGE_DIGEST}"
WORKFLOW="https://github.com/myorg/SecureChain-Zero/.github/workflows/builder.yml@refs/heads/main"
OIDC_ISSUER="https://token.actions.githubusercontent.com"

echo "=========================================================="
echo "Verifying Supply Chain Trust for: ${IMAGE}"
echo "=========================================================="

echo "=== 1. Verifying Cosign Keyless Image Signature ==="
cosign verify \
  --certificate-identity="${WORKFLOW}" \
  --certificate-oidc-issuer="${OIDC_ISSUER}" \
  "${IMAGE}" | jq .

echo "=== 2. Verifying SLSA v1.2 Provenance Attestation ==="
cosign verify-attestation \
  --type slsaprovenance \
  --certificate-identity="${WORKFLOW}" \
  --certificate-oidc-issuer="${OIDC_ISSUER}" \
  "${IMAGE}" | jq '.payload | @base64d | fromjson | .predicate'

echo "=== 3. Verifying CycloneDX SBOM Attestation (EO 14028 §4) ==="
cosign verify-attestation \
  --type cyclonedx \
  --certificate-identity="${WORKFLOW}" \
  --certificate-oidc-issuer="${OIDC_ISSUER}" \
  "${IMAGE}" | jq '.payload | @base64d | fromjson | .predicate.components | length'

echo "=== 4. Verifying OpenVEX Attestation (NIST SSDF RV.1.3) ==="
cosign verify-attestation \
  --type openvex \
  --certificate-identity="${WORKFLOW}" \
  --certificate-oidc-issuer="${OIDC_ISSUER}" \
  "${IMAGE}" | jq '.payload | @base64d | fromjson | .predicate'

echo ""
echo "✅ Full attestation chain verified: Signature, SLSA L3 Provenance, CycloneDX SBOM, OpenVEX."
