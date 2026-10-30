# Security Policy

## Reporting a Vulnerability

The SecureChain-Zero team takes software supply chain security and runtime integrity seriously. If you discover a vulnerability, misconfiguration, or security loophole within this repository, please follow this disclosure process.

### Supported Versions

Only the latest commit on the `main` branch is actively supported and evaluated for security vulnerabilities.

| Version / Branch | Supported          |
| ---------------- | ------------------ |
| `main`           | :white_check_mark: |
| Feature Branches | :x:                |

---

## Reporting Process

1. **Private Vulnerability Disclosure**:
   - Please report vulnerabilities directly via GitHub's **Private Vulnerability Reporting** feature in this repository.
   - Alternatively, email `security@myorg.com` with the subject: `[VULNERABILITY] SecureChain-Zero - <Component>`.
2. **Include Required Details**:
   - Description of the vulnerability and attack vector (e.g., Build-time Poisoning, OIDC impersonation, Kyverno gate bypass, Container breakout).
   - Reproduction steps or Proof of Concept (PoC).
   - Impact assessment according to CVSS v3.1 / v4.0.
3. **Response Timeline**:
   - **T+24 hours**: Initial triage acknowledgement and war room assignment.
   - **T+72 hours**: Formal exploitability assessment and remediation plan.
   - **T+7 days**: Coordinated patch release, signed OpenVEX statement, and public disclosure.

---

## Supply Chain Security Principles

All artifacts built in this repository adhere to:
- **SLSA v1.2 Level 3**: Hermetic builds, isolated build environments, cryptographically verifiable provenance.
- **NIST SP 800-218 (SSDF)**: Provenance attestation, continuous SBOM generation (CycloneDX 1.5), and admission gate enforcement.
- **Keyless Sigstore Architecture**: Ephemeral OIDC-bound signing certs via Fulcio CA with Rekor transparency logging.
- **Runtime Defense**: In-kernel eBPF syscall monitoring via Cilium Tetragon with immediate SIGKILL on unauthorized process execution.
