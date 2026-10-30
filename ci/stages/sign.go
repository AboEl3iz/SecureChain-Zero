package stages

import (
	"fmt"
	"os"
)

// ExportRunnerArtifacts exports build outputs (image digest and SBOM path)
// so the GitHub Actions runner can perform Cosign keyless signing and attestation
// with its native OIDC identity token (SLSA L3 boundary separation).
func ExportRunnerArtifacts(digest, sbomPath string) error {
	digestFile := "/tmp/image-digest.txt"
	if err := os.WriteFile(digestFile, []byte(digest), 0600); err != nil {
		return fmt.Errorf("failed to write image digest artifact: %w", err)
	}

	sbomFile := "/tmp/sbom-path.txt"
	if err := os.WriteFile(sbomFile, []byte(sbomPath), 0600); err != nil {
		return fmt.Errorf("failed to write sbom path artifact: %w", err)
	}

	return nil
}
