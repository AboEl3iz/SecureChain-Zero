#!/usr/bin/env bash
# scripts/incident-response/virtual-patch.sh
# Applies an emergency virtual patch at the Ingress layer to block malicious payloads (CVE-2026-111497)
set -euo pipefail

NAMESPACE="${1:-production}"
INGRESS_NAME="${2:-secure-app-ingress}"

echo "=== Applying Emergency Ingress Virtual Patch ==="
echo "Target: ${INGRESS_NAME} in namespace ${NAMESPACE}"

cat <<EOF | kubectl apply -n "${NAMESPACE}" -f -
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ${INGRESS_NAME}
  annotations:
    nginx.ingress.kubernetes.io/server-snippet: |
      # Virtual patch: block malicious Brotli payloads (CVE-2026-111497)
      if (\$http_content_encoding ~* "br") {
          return 406 "Content-Encoding: br is not accepted";
      }
spec:
  ingressClassName: nginx
  rules:
    - http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: secure-app-service
                port:
                  number: 8080
EOF

echo "✅ Virtual patch applied successfully. Inbound Brotli decompression payloads are blocked."
