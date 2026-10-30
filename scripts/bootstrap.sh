#!/usr/bin/env bash
# scripts/bootstrap.sh — DevSecOps Toolchain Bootstrap
# Installs apko, cosign, crane, dagger, trivy, syft, vexctl
set -euo pipefail

BIN_DIR="/usr/local/bin"
if [ ! -w "$BIN_DIR" ]; then
  BIN_DIR="${HOME}/.local/bin"
  mkdir -p "$BIN_DIR"
  export PATH="${BIN_DIR}:${PATH}"
  echo "Notice: Installing to ${BIN_DIR} (non-root mode)"
fi

ARCH=$(uname -m)
case "$ARCH" in
  x86_64)
    ARCH_COSIGN="amd64"
    ARCH_APKO="amd64"
    ARCH_CRANE="x86_64"
    ARCH_TRIVY="64bit"
    ARCH_SYFT="amd64"
    ARCH_VEXCTL="x86_64"
    ;;
  aarch64|arm64)
    ARCH_COSIGN="arm64"
    ARCH_APKO="arm64"
    ARCH_CRANE="arm64"
    ARCH_TRIVY="ARM64"
    ARCH_SYFT="arm64"
    ARCH_VEXCTL="arm64"
    ;;
  *)
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

echo "=== Installing Cosign (Sigstore) ==="
COSIGN_VERSION="v2.4.1"
curl -sSL "https://github.com/sigstore/cosign/releases/download/${COSIGN_VERSION}/cosign-linux-${ARCH_COSIGN}" \
  -o "${BIN_DIR}/cosign" && chmod +x "${BIN_DIR}/cosign"
echo "Cosign installed: $("${BIN_DIR}/cosign" version --json 2>/dev/null || echo "$COSIGN_VERSION")"

echo "=== Installing apko (Chainguard) ==="
APKO_VERSION="v0.19.4"
curl -sSL "https://github.com/chainguard-dev/apko/releases/download/${APKO_VERSION}/apko_${APKO_VERSION#v}_linux_${ARCH_APKO}.tar.gz" \
  | tar xz -C "${BIN_DIR}" apko && chmod +x "${BIN_DIR}/apko"
echo "apko installed: $("${BIN_DIR}/apko" version 2>/dev/null || echo "$APKO_VERSION")"

echo "=== Installing Crane ==="
CRANE_VERSION="v0.20.2"
curl -sSL "https://github.com/google/go-containerregistry/releases/download/${CRANE_VERSION}/go-containerregistry_Linux_${ARCH_CRANE}.tar.gz" \
  | tar xz -C "${BIN_DIR}" crane && chmod +x "${BIN_DIR}/crane"
echo "crane installed: $("${BIN_DIR}/crane" version 2>/dev/null || echo "$CRANE_VERSION")"

echo "=== Installing Dagger ==="
DAGGER_VERSION="0.14.0"
curl -fsSL https://dl.dagger.io/dagger/install.sh | DAGGER_VERSION="${DAGGER_VERSION}" BIN_DIR="${BIN_DIR}" sh
echo "dagger installed: $("${BIN_DIR}/dagger" version 2>/dev/null || echo "$DAGGER_VERSION")"

echo "=== Installing Trivy ==="
TRIVY_VERSION="0.57.1"
curl -sSL "https://github.com/aquasecurity/trivy/releases/download/v${TRIVY_VERSION}/trivy_${TRIVY_VERSION}_Linux-${ARCH_TRIVY}.tar.gz" \
  | tar xz -C "${BIN_DIR}" trivy && chmod +x "${BIN_DIR}/trivy"
echo "trivy installed: $("${BIN_DIR}/trivy" --version 2>/dev/null || echo "$TRIVY_VERSION")"

echo "=== Installing Syft ==="
SYFT_VERSION="v1.11.1"
curl -sSL "https://github.com/anchore/syft/releases/download/${SYFT_VERSION}/syft_${SYFT_VERSION#v}_linux_${ARCH_SYFT}.tar.gz" \
  | tar xz -C "${BIN_DIR}" syft && chmod +x "${BIN_DIR}/syft"
echo "syft installed: $("${BIN_DIR}/syft" version 2>/dev/null || echo "$SYFT_VERSION")"

echo "=== Installing vexctl ==="
VEXCTL_VERSION="v0.3.0"
curl -sSL "https://github.com/openvex/vexctl/releases/download/${VEXCTL_VERSION}/vexctl_${VEXCTL_VERSION#v}_linux_${ARCH_VEXCTL}.tar.gz" \
  | tar xz -C "${BIN_DIR}" vexctl 2>/dev/null || true
if [ -f "${BIN_DIR}/vexctl" ]; then
  chmod +x "${BIN_DIR}/vexctl"
  echo "vexctl installed"
fi

echo ""
echo "✅ Toolchain bootstrap complete. Binaries available in: ${BIN_DIR}"
