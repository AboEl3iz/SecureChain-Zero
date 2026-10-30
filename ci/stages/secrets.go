package stages

import (
	"context"
	"fmt"

	"dagger.io/dagger"
)

// RunSecretScan executes Gitleaks against the codebase (OWASP CICD-SEC-6).
func RunSecretScan(ctx context.Context, client *dagger.Client, src *dagger.Directory) error {
	output, err := client.Container().
		From("ghcr.io/gitleaks/gitleaks:v8.21.2").
		WithMountedDirectory("/src", src).
		WithWorkdir("/src").
		WithExec([]string{
			"gitleaks", "dir",
			"--no-git",
			"--exit-code=1",
			"--config=/src/policies/gitleaks/.gitleaks.toml",
			"--report-format=sarif",
			"--report-path=/tmp/gitleaks.sarif",
			"--verbose",
		}).
		Stdout(ctx)

	if err != nil {
		return fmt.Errorf("SECRET GATE FAILED — credentials/tokens detected in source:\n%s (err: %w)", output, err)
	}
	return nil
}
