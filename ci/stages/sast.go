package stages

import (
	"context"
	"fmt"

	"dagger.io/dagger"
)

// RunSASTScan executes Semgrep static analysis with security rule enforcement (OWASP CICD-SEC-3).
func RunSASTScan(ctx context.Context, client *dagger.Client, src *dagger.Directory) error {
	output, err := client.Container().
		From("returntocorp/semgrep:1.90.0").
		WithMountedDirectory("/src", src).
		WithWorkdir("/src").
		WithExec([]string{
			"semgrep",
			"--config=/src/policies/semgrep/.semgrep.yaml",
			"--config=auto",
			"--sarif-output=/tmp/semgrep.sarif",
			"--severity=ERROR",
			"--error", // Exit non-zero on findings
		}).
		Stdout(ctx)

	if err != nil {
		return fmt.Errorf("SAST GATE FAILED — security vulnerabilities detected:\n%s (err: %w)", output, err)
	}
	return nil
}
