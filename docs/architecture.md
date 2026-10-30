# System Architecture — SecureChain-Zero (Project HermeticGate)

## 1. Architectural Philosophy

The SecureChain-Zero architecture operates on five non-negotiable security principles:

1. **Hermetic by Construction**: Every build is executed inside an isolated container sandbox where network inputs are strictly pinned by cryptographic digest. Re-running the build yields bit-for-bit identical outputs.
2. **Keyless Sigstore Cryptography**: No private signing keys are generated or stored on disk or in CI secrets. Ephemeral certificates are minted by Sigstore Fulcio via OpenID Connect (OIDC) identity tokens and anchored into the Rekor transparency log.
3. **Cryptographically Attested Artifacts**: Binaries and container images do not travel alone; they are accompanied by in-toto attestations covering SLSA provenance, CycloneDX SBOMs, and OpenVEX exploitability statements.
4. **Admission Control as Non-Bypassable Gate**: Untrusted or unverified images cannot execute inside Kubernetes. The Kyverno Admission Controller verifies signatures and attestations before any Pod is scheduled.
5. **Active In-Kernel eBPF Containment**: Runtime security cannot depend on userspace agents. Cilium Tetragon intercepts system calls at the Linux kernel boundary and delivers instantaneous `SIGKILL` to unauthorized processes.

---

## 2. Core Subsystems

### 2.1 Packaging: apko & Wolfi
- **Wolfi Linux**: Designed specifically for containers and cloud-native applications, providing daily security rebuilds and rapid patch turnaround for OS packages.
- **apko**: Generates distroless OCI images directly from package repositories without Docker daemon privileges or intermediate layers.
- **Reproducibility**: Locked via `build/apko.lock.json`, which pins package names, architectures, versions, and SHA256 hashes.

### 2.2 Pipeline Engine: Dagger (Go SDK)
- Pipelines are written in typed Go, eliminating brittle Bash scripts and vulnerable YAML configurations.
- Execution graphs are cached intelligently and run consistently on developer laptops and CI runners.
- Build/Sign boundary separation: Dagger executes build, SAST, secret detection, and image packaging; Cosign keyless signing runs directly on the runner with access to the GitHub Actions OIDC token.

### 2.3 Cryptographic Trust: Sigstore
- **Fulcio CA**: Mints 10-minute X.509 certificates tied to GitHub Actions workflow claims (`builder.yml`).
- **Rekor**: Publicly auditable transparency log ensuring non-repudiation.
- **Cosign**: Signs image indexes and attaches CycloneDX 1.5 SBOMs, SLSA v1.2 Provenance, and OpenVEX JSON documents as OCI layers.

### 2.4 Admission Gate: Kyverno
- Evaluates `verifyImages` rules against the private registry (`ghcr.io/myorg/*`).
- Replaces image tags with immutable SHA256 digests (`mutateDigest: true`).
- Validates that Pod security requirements are met (UID 65532 nonroot, readOnlyRootFilesystem, and memory-backed emptyDir at `/tmp`).

### 2.5 Runtime Defense: Cilium Tetragon
- Attaches eBPF kprobes to `sys_execve`, `sys_openat`, and `sys_connect`.
- Kills any process attempting to execute shells (`/bin/sh`, `/bin/bash`, `/usr/bin/*`).
- Reports anomalous kernel activity in real time.
