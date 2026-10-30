#!/usr/bin/env bash
# scripts/local-run.sh — Run Dagger pipeline locally
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

echo "=== SecureChain-Zero Local Pipeline Execution ==="
echo "Working directory: ${ROOT_DIR}"

if ! command -v dagger >/dev/null 2>&1; then
  echo "Dagger is not installed in PATH. Running bootstrap..."
  "${SCRIPT_DIR}/bootstrap.sh"
fi

export REGISTRY="${REGISTRY:-ghcr.io/myorg/secure-app}"
export IMAGE_TAG="${IMAGE_TAG:-local-$(date +%s)}"
export REGISTRY_TOKEN="${REGISTRY_TOKEN:-dummy-local-token}"

cd "${ROOT_DIR}/ci"
echo "Starting Dagger engine and executing hermetic pipeline..."
go run main.go

echo "=== Local execution finished successfully ==="
