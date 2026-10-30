# Incident Response Playbook: Zero-Day RCE (Node2Shell / Brotli Scenario)

## 1. Trigger & Timeline

This runbook activates upon disclosure of a critical zero-day Remote Code Execution vulnerability in a core package (e.g., CVE-2026-111497: Node2Shell Brotli heap overflow).

```
T+0:00   CVE published on OSV.dev / NVD
T+0:05   Dependency-Track detects match against CycloneDX SBOM → Alerts fired
T+0:10   Incident Commander (IC) declared, war room initialized
T+0:15   STEP 1: Apply Virtual Patch at Ingress (Header Block)
T+0:20   STEP 2: Verify Egress NetworkPolicy lockdown
T+0:30   STEP 3: Inspect Tetragon eBPF telemetry for active exploit attempts
T+1:00   STEP 4: Upstream Wolfi patch verification & emergency rebuild
T+1:30   STEP 5: Canary rollout of attested, signed image
T+2:00   STEP 6: OpenVEX documentation update and sign-off
```

---

## 2. Execution Steps

### Step 1: Immediate Ingress Virtual Patch
Block requests specifying `Content-Encoding: br` at the perimeter before they hit the container workload:
```bash
./scripts/incident-response/virtual-patch.sh production secure-app-ingress
```
*Verify with curl:*
```bash
curl -i -H "Content-Encoding: br" https://app.myorg.com/
# Expected: HTTP 406 Not Acceptable
```

### Step 2: Validate Egress Network Policy
Ensure pods cannot establish outbound connections to attacker Command & Control (C2) servers:
```bash
kubectl get networkpolicy -n production secure-app-strict-egress -o yaml
```

### Step 3: Tetragon Kernel Telemetry Review
Check whether any container processes triggered Tetragon `SIGKILL` or anomalous execve kprobes:
```bash
kubectl logs -n kube-system -l app.kubernetes.io/name=tetragon -c export-stdout \
  | grep "sys_execve" | jq 'select(.process.binary | startswith("/bin/"))'
```

### Step 4: Emergency Rebuild with Pinned Fix
When upstream Wolfi publishes a patched package:
```bash
./scripts/incident-response/emergency-rebuild.sh wolfi-base 1:1.0-r9
```

### Step 5: Update OpenVEX for Non-Affected Services
For services where the vulnerable code path cannot be invoked:
```bash
./scripts/incident-response/vex-update.sh secure-app CVE-2026-111497 <NEW_IMAGE_DIGEST>
```
