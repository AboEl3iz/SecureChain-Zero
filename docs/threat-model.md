# Threat Model & Attack Vector Analysis

## 1. Scope & Objective

This threat model outlines supply chain and runtime threats confronting modern microservices, mapping concrete attack scenarios to SecureChain-Zero mitigations.

---

## 2. Threat Actor Profiles

| Threat Actor ID | Profile | Motivations | Capabilities |
|---|---|---|---|
| **OCI-1** | Malicious Dependency Contributor | Upstream compromise (e.g., XZ Utils) | Inject obfuscated malicious code or trojanized release tars |
| **OCI-2** | Compromised Developer Machine | Push tainted image directly to registry | Stolen registry credentials, local build tampering |
| **OCI-3** | Poisoned Pipeline Attacker (PPE) | Exfiltrate secrets or inject build logic | Malicious Pull Request modifying CI workflow scripts |
| **OCI-4** | Zero-Day Runtime Exploiter | RCE via unpatched memory or parser bug | Exploit Brotli decompression or HTTP buffer overflow |
| **OCI-5** | Host Breakout Attacker | Container breakout to Kubernetes node | Exploit excess Linux capabilities or root privileges |

---

## 3. Threat Mitigation Matrix

### OCI-1: Upstream Compromise
- **Mitigation**:
  - `build/apko.lock.json` pins exact SHA256 package digests.
  - Multi-source vulnerability correlation (OSV, NVD, GitHub Advisory) via Dependency-Track.
  - Static analysis with Semgrep blocking dangerous library calls.

### OCI-2: Direct Registry Poisoning
- **Mitigation**:
  - Registry allows image pushes only from authenticated CI runners.
  - Kyverno ClusterPolicy `enforce-supply-chain-trust` blocks admission of images unless signed by the specific GitHub Actions reusable workflow identity URL.
  - Developer workstation signatures are rejected by admission control.

### OCI-3: Poisoned Pipeline Execution (PPE)
- **Mitigation**:
  - Reusable workflow isolation (`builder.yml`).
  - Caller workflow (`ci.yml`) runs with read-only permissions and triggers the builder only on protected `main` branch merges.
  - Pipeline logic is compiled Go code executed inside Dagger, not easily tampered inline shell steps.

### OCI-4: Zero-Day RCE (Node2Shell)
- **Mitigation**:
  - Distroless Wolfi image lacks `/bin/sh`, `/bin/bash`, Python, curl, and wget.
  - `readOnlyRootFilesystem: true` prevents dropping malware binaries.
  - Tetragon eBPF kills any `execve` invoking system binaries with `Sigkill`.

### OCI-5: Container Escape
- **Mitigation**:
  - Container runs strictly as UID 65532 (`nonroot`).
  - All Linux capabilities dropped (`drop: ["ALL"]`).
  - Privilege escalation disabled (`allowPrivilegeEscalation: false`).
  - Host namespaces forbidden (`hostPID`, `hostIPC`, `hostNetwork`).
