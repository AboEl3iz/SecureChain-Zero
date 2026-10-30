# SecureChain-Zero / Project HermeticGate

> **Enterprise DevSecOps Blueprint**: Hermetic Builds, SLSA v1.2 Level 3 Provenance, Keyless Sigstore Attestation, Kyverno Admission Control, and Cilium Tetragon eBPF Runtime Defense.

[![SLSA Level 3](https://img.shields.io/badge/SLSA-Level%203-blue.svg)](https://slsa.dev)
[![NIST SSDF](https://img.shields.io/badge/NIST-SP%20800--218-green.svg)](https://csrc.nist.gov/pubs/sp/800/218/final)
[![EO 14028](https://img.shields.io/badge/Compliance-EO%2014028-orange.svg)](https://www.whitehouse.gov/briefing-room/presidential-actions/2021/05/12/executive-order-on-improving-the-nations-cybersecurity/)
[![Sigstore](https://img.shields.io/badge/Signing-Cosign%20Keyless-purple.svg)](https://sigstore.dev)

---

## 1. Overview & Threat Model

Modern software supply chains are increasingly targeted at **build time**, rather than merely at runtime. SecureChain-Zero implements a zero-trust end-to-end security architecture that systematically defends against canonical supply chain breaches:

| Attack | Vector | SecureChain-Zero Defense |
|---|---|---|
| **SolarWinds (2020)** | Poisoned build server injecting backdoors into artifacts | SLSA L3 isolated builder + OIDC-bound cryptographic provenance |
| **Log4Shell (2021)** | Vulnerable transitive dependencies lurking in images | Automated CycloneDX SBOM + Continuous Dependency-Track scanning |
| **XZ Utils (2024)** | Malicious maintainer compromising upstream builds | Bit-for-bit reproducible apko builds + digest pinning via `apko.lock.json` |
| **Node2Shell (2026)** | Zero-day RCE via native libraries | Shell-less Wolfi base + readOnlyRootFS + Tetragon eBPF SIGKILL |
---

## 2. Architecture & Pipeline Flow

```
┌───────────────────────────┐     ┌────────────────────────────────────────────────────────┐
│   Developer Workstation   │     │        SLSA L3 Reusable Builder (.github/workflows)    │
│  Gitleaks pre-commit hook │────►│  ┌──────────────────────────────────────────────────┐  │
└───────────────────────────┘     │  │          Dagger Engine (BuildKit Sandbox)        │  │
                                  │  │  1. Gitleaks Secret Gate                         │  │
                                  │  │  2. Semgrep SAST Gate                            │  │
                                  │  │  3. Go Multi-Arch Cross-Compile (amd64/arm64)    │  │
                                  │  │  4. apko Wolfi Hermetic Packaging + SBOM         │  │
                                  │  │  5. Trivy SARIF Vulnerability Gate (CRITICAL=1)  │  │
                                  │  │  6. Multi-Arch Index Push (crane)                │  │
                                  │  └──────────────────────┬───────────────────────────┘  │
                                  │                         │                              │
                                  │  7. GHA Runner Cosign Keyless OIDC Signing:            │
                                  │     - Cosign sign container image                      │
                                  │     - Cosign attest CycloneDX SBOM                     │
                                  │     - Cosign attest OpenVEX document                   │
                                  │     - actions/attest-build-provenance (SLSA v1.2)      │
                                  └─────────────────────────┬──────────────────────────────┘
                                                            │ OCI Image + Attestations
                                                            ▼
                                  ┌────────────────────────────────────────────────────────┐
                                  │       Enterprise Kubernetes Cluster (Production)       │
                                  │  ┌──────────────────────────────────────────────────┐  │
                                  │  │         Kyverno Admission Controller Webhook     │  │
                                  │  │   ✓ Verify Cosign Keyless Signature (Fulcio/Rekor│  │
                                  │  │   ✓ Verify SLSA L3 Provenance                    │  │
                                  │  │   ✓ Verify CycloneDX 1.5+ SBOM Attestation       │  │
                                  │  │   ✓ Enforce non-root (UID 65532), readOnlyRootFS │  │
                                  │  │   ✓ Enforce emptyDir memory volume at /tmp       │  │
                                  │  └──────────────────────┬───────────────────────────┘  │
                                  │                         │ Validated Pod Admission      │
                                  │                         ▼                              │
                                  │  ┌──────────────────┐       ┌───────────────────────┐  │
                                  │  │ Hardened Pod     │       │ Cilium Tetragon eBPF  │  │
                                  │  │ - UID 65532      │◄──────│ - sys_execve SIGKILL  │  │
                                  │  │ - readOnlyRootFS │       │ - sys_openat monitor  │  │
                                  │  │ - /tmp tmpfs RAM │       │ - sys_connect monitor │  │
                                  │  └──────────────────┘       └───────────────────────┘  │
                                  └────────────────────────────────────────────────────────┘
```

---

## 3. Directory Layout

```
SecureChain-Zero/
├── README.md                          # Architecture diagram + quick-start
├── SECURITY.md                        # Vulnerability reporting policy
├── .gitignore                         # Build, binary, secret, and core dump filters
│
├── sample-app/                        # Demo Go microservice (HTTP + health/readiness)
│   ├── main.go                        # Entrypoint with graceful termination
│   ├── go.mod                         # Module definition (Go 1.23)
│   └── internal/
│       ├── config/                    # Environment configuration loader
│       └── server/                    # HTTP server, security headers, /tmp check
│
├── build/                             # Hermetic container packaging
│   ├── apko.yaml                      # Wolfi-based declarative image definition
│   ├── apko.lock.json                 # Pinned package digests (reproducibility)
│   └── sbom/                          # Generated SBOM directory (.gitkeep)
│
├── ci/                                # Dagger CI pipeline (Go SDK)
│   ├── main.go                        # Pipeline entrypoint orchestrating stages
│   ├── go.mod                         # Pipeline dependencies
│   ├── dagger.json                    # Dagger module configuration
│   └── stages/
│       ├── secrets.go                 # Gitleaks scanner
│       ├── sast.go                    # Semgrep SAST scanner
│       ├── build.go                   # Multi-arch Go build + apko packaging
│       ├── scan.go                    # Trivy vulnerability scanner
│       ├── push.go                    # Crane multi-arch image index publisher
│       └── sign.go                    # Runner artifact exporter (SLSA L3 boundary)
│
├── .github/
│   ├── settings.yml                   # Branch protection rules & signed commit policy
│   └── workflows/
│       ├── ci.yml                     # Caller workflow (push/PR trigger)
│       └── builder.yml                # SLSA L3 isolated reusable workflow
│
├── policies/
│   ├── kyverno/
│   │   ├── enforce-supply-chain.yaml  # Kyverno verifyImages & attestations
│   │   ├── enforce-pod-security.yaml  # Kyverno non-root, readOnlyRootFS, tmpfs
│   │   └── restrict-image-registry.yaml# Restrict pulls to ghcr.io/myorg/*
│   ├── gitleaks/
│   │   └── .gitleaks.toml             # Gitleaks rule definitions & allowlists
│   ├── semgrep/
│   │   └── .semgrep.yaml              # Semgrep rules (Brotli SC-BROTLI-001)
│   └── trivy/
│       └── .trivyignore               # Documented OpenVEX exceptions
│
├── vex/
│   ├── openvex.json                   # Signed OpenVEX statement
│   └── vex-generator/                # Standalone Go CLI to create/validate VEX
│       ├── go.mod
│       └── main.go
│
├── k8s/
│   ├── namespace.yaml                 # Production & staging namespaces (restricted PSA)
│   ├── deployment.yaml                # Hardened deployment with /tmp emptyDir in RAM
│   ├── network-policy.yaml            # Strict egress isolation (DNS & internal only)
│   ├── kyverno/values.yaml            # Kyverno Helm chart values
│   ├── tetragon/tracing-policy.yaml   # Tetragon eBPF TracingPolicy (SIGKILL on execve)
│   └── trivy-operator/values.yaml     # Trivy Operator continuous cluster scanning
│
├── scripts/
│   ├── bootstrap.sh                   # Toolchain installer (apko, cosign, crane, etc.)
│   ├── local-run.sh                   # Local pipeline runner via Dagger
│   ├── verify.sh                      # Cryptographic trust verification script
│   └── incident-response/
│       ├── virtual-patch.sh           # Ingress-level Brotli header block
│       ├── vex-update.sh              # Update & re-sign OpenVEX statement
│       └── emergency-rebuild.sh       # Emergency bump & lock regeneration
│
└── docs/
    ├── architecture.md                # Comprehensive architectural design
    ├── threat-model.md                # Supply chain threat modeling & mitigations
    ├── compliance-matrix.md           # NIST SSDF, SLSA L3, EO 14028 mapping
    └── runbooks/
        ├── rce-incident-response.md   # Node2Shell zero-day response playbook
        └── vulnerability-triage.md    # Vulnerability triage & OpenVEX workflow
```

---

## 4. Quick Start

### 4.1 Bootstrap Local Toolchain
```bash
./scripts/bootstrap.sh
```

### 4.2 Run Local Hermetic Pipeline
```bash
./scripts/local-run.sh
```

### 4.3 Verify Cryptographic Trust Chain
```bash
./scripts/verify.sh sha256:<IMAGE_DIGEST> ghcr.io/myorg/secure-app
```

### 4.4 Test Kubernetes Policies Locally
```bash
# Test Kyverno policies with Kyverno CLI
kyverno test policies/kyverno/
```
