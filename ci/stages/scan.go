package stages

import (
	"context"
	"fmt"

	"dagger.io/dagger"
)

// RunVulnScan executes Trivy container image scanning with SARIF output and a CRITICAL gate.
func RunVulnScan(
	ctx context.Context,
	client *dagger.Client,
	src *dagger.Directory,
	imageTar *dagger.File,
) error {
	_, err := client.Container().
		From("aquasec/trivy:0.57.1").
		WithMountedDirectory("/src", src).
		WithMountedFile("/image.tar", imageTar).
		WithExec([]string{
			"trivy", "image",
			"--input", "/image.tar",
			"--severity", "CRITICAL",
			"--exit-code", "1", // Fail pipeline on CRITICAL CVE
			"--format", "table",
			"--ignorefile", "/src/policies/trivy/.trivyignore",
		}).
		Sync(ctx)

	if err != nil {
		return fmt.Errorf("VULNERABILITY GATE FAILED — CRITICAL CVEs discovered in container image: %w", err)
	}
	return nil
}
