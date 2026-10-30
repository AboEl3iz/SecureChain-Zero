# Compliance & Standards Mapping Matrix

## 1. Regulatory & Industry Standards

This document establishes formal traceability between SecureChain-Zero controls and key cybersecurity standards:
- **NIST SP 800-218**: Secure Software Development Framework (SSDF)
- **SLSA v1.2**: Supply-chain Levels for Software Artifacts (Level 3 Target)
- **Executive Order (EO) 14028**: Improving the Nation's Cybersecurity (§4 Software Supply Chain)
- **OWASP Top 10 CI/CD Security Risks**

---

## 2. Standards Cross-Reference

| Standard | Control ID | Requirement Description | Implementation in SecureChain-Zero | Evidence / Verification Artifact |
|---|---|---|---|---|
| **NIST SSDF** | **PS.1.1** | Protect all code branch access and integrity | Branch protection rules + signed commits via Gitsign | `.github/settings.yml` |
| **NIST SSDF** | **PS.2.1** | Use trusted build platforms with limited access | GitHub Reusable Workflow + Dagger isolation | `.github/workflows/builder.yml` |
| **NIST SSDF** | **PS.3.1** | Protect build process outputs from tampering | Cosign keyless signatures bound to GitHub OIDC token | Rekor log entry & `cosign verify` |
| **NIST SSDF** | **PS.3.2** | Non-bypassable gate preventing unsigned code in prod | Kyverno ClusterPolicy `enforce-supply-chain-trust` | `policies/kyverno/enforce-supply-chain.yaml` |
| **NIST SSDF** | **PW.4.1** | Identify all software components in an SBOM | apko generates CycloneDX 1.5 SBOM automatically | `build/sbom/cyclonedx.json` |
| **NIST SSDF** | **PW.4.4** | Analyze components for known vulnerabilities | Trivy image scan with zero CRITICAL gate in CI | `ci/stages/scan.go` |
| **NIST SSDF** | **RV.1.1** | Gather vulnerability reports from multiple feeds | Dependency-Track continuous ingestion (OSV/NVD/GHSA) | `k8s/trivy-operator/values.yaml` |
| **NIST SSDF** | **RV.1.3** | Analyze and prioritize CVEs by exploitability | OpenVEX document with signed non-affected statements | `vex/openvex.json` |
| **NIST SSDF** | **RV.2.2** | Continuous vulnerability scans of deployed images | In-cluster Trivy Operator scanning running Pods | `k8s/trivy-operator/values.yaml` |
| **SLSA v1.2** | **Build L1** | Build from version-controlled source | Pinned Git commit SHA in build provenance | Attestation in GHCR |
| **SLSA v1.2** | **Build L2** | Hosted build platform with signed provenance | GitHub Actions runner + provenance attestation | `actions/attest-build-provenance` |
| **SLSA v1.2** | **Build L3** | Hardened, isolated build; non-forgeable provenance | Dagger Engine + isolated reusable workflow + OIDC | `actions/attest-build-provenance` |
| **EO 14028** | **§4(e)** | Provide machine-readable signed SBOM | CycloneDX 1.5 SBOM attached via `cosign attest` | OCI SBOM attestation layer |
| **OWASP CI/CD**| **CICD-SEC-1**| Insufficient Flow Control Mechanisms | Strict review gates, linear history, status checks | `.github/settings.yml` |
| **OWASP CI/CD**| **CICD-SEC-2**| Inadequate Identity & Access Management | Ephemeral OIDC tokens only; zero static keys in CI | `.github/workflows/builder.yml` |
| **OWASP CI/CD**| **CICD-SEC-3**| Dependency Chain Abuse | `apko.lock.json` digest pinning; container digest pins | `build/apko.lock.json` |
| **OWASP CI/CD**| **CICD-SEC-6**| Insufficient Credential Hygiene | Gitleaks pre-commit hook and CI gate | `policies/gitleaks/.gitleaks.toml` |
| **OWASP CI/CD**| **CICD-SEC-9**| Improper Artifact Integrity Validation | Cosign signature + SLSA provenance verified by Kyverno| `policies/kyverno/` |
