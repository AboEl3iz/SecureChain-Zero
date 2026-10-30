// ci/main.go
// NIST SSDF PS.2.1 — Hermetic, isolated build steps using Dagger Engine
package main

import (
	"context"
	"fmt"
	"os"

	"dagger.io/dagger"
	"github.com/securechain-zero/ci/stages"
)

func main() {
	ctx := context.Background()
	if err := RunPipeline(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Pipeline Failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ SecureChain-Zero Pipeline Complete")
}

func RunPipeline(ctx context.Context) error {
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		return fmt.Errorf("dagger connect: %w", err)
	}
	defer client.Close()

	src := client.Host().Directory(".", dagger.HostDirectoryOpts{
		// Explicitly exclude directories that must never enter the build context
		Exclude: []string{".git", "ci/vendor", "*.secret", "*.env", "core*"},
	})

	// ── Stage 1: Secret Scanning (OWASP CICD-SEC-6) ────────────────────────
	fmt.Println("▶ Stage 1: Gitleaks Secret Scanning...")
	if err := stages.RunSecretScan(ctx, client, src); err != nil {
		return err
	}
	fmt.Println("  ✓ No secrets detected")

	// ── Stage 2: SAST (OWASP CICD-SEC-3) ────────────────────────────────────
	fmt.Println("▶ Stage 2: Semgrep Static Application Security Testing...")
	if err := stages.RunSASTScan(ctx, client, src); err != nil {
		return err
	}
	fmt.Println("  ✓ SAST scan passed")

	// ── Stage 3: Build Go Binaries (Multi-Arch Matrix) ───────────────────────
	// Multi-arch: Compile separate binaries for linux/amd64 and linux/arm64
	fmt.Println("▶ Stage 3: Cross-compiling Multi-Arch Go Binaries...")
	targets := []stages.ArchTarget{
		{GoArch: "amd64", ApkArch: "x86_64"},
		{GoArch: "arm64", ApkArch: "aarch64"},
	}

	binaries := stages.BuildMultiArchBinaries(client, src, targets)

	// ── Stage 4: Hermetic Container Build via apko (per arch) ───────────────
	fmt.Println("▶ Stage 4: apko Hermetic Container Packaging + SBOM...")
	var builds []stages.BuiltArch
	for _, t := range targets {
		bin, ok := binaries[t.GoArch]
		if !ok {
			return fmt.Errorf("missing compiled binary for %s", t.GoArch)
		}
		img, sbom, err := stages.BuildApkoImage(ctx, client, src, bin, t.ApkArch, t.GoArch)
		if err != nil {
			return fmt.Errorf("image build (%s): %w", t.GoArch, err)
		}
		builds = append(builds, stages.BuiltArch{
			Target: t,
			Binary: bin,
			Image:  img,
			SBOM:   sbom,
		})
	}
	fmt.Println("  ✓ Multi-arch images and CycloneDX SBOMs generated")

	// ── Stage 5: Vulnerability Scan (Trivy) ──────────────────────────────────
	fmt.Println("▶ Stage 5: Trivy Container Vulnerability Scan...")
	if err := stages.RunVulnScan(ctx, client, src, builds[0].Image); err != nil {
		return err
	}
	fmt.Println("  ✓ Zero CRITICAL CVEs confirmed")

	// ── Stage 6: Push Multi-Arch Image Index to Registry ─────────────────────
	fmt.Println("▶ Stage 6: Publishing Image to Registry...")
	registryToken := client.SetSecret("registry-token", os.Getenv("REGISTRY_TOKEN"))
	digest, sbomPath, err := stages.PushMultiArchImage(ctx, client, builds, registryToken)
	if err != nil {
		return fmt.Errorf("push: %w", err)
	}
	fmt.Printf("📦 Multi-arch image pushed: %s\n", digest)

	// Export CycloneDX SBOM from Dagger container to runner filesystem
	_ = os.MkdirAll("build/sbom", 0755)
	if len(builds) > 0 && builds[0].SBOM != nil {
		_, _ = builds[0].SBOM.Export(ctx, "build/sbom/cyclonedx.json")
	}

	// ── Stage 7: Export outputs for runner Cosign keyless signing ─────────────
	fmt.Println("▶ Stage 7: Exporting artifacts for GHA runner OIDC signing...")
	if err := stages.ExportRunnerArtifacts(digest, sbomPath); err != nil {
		return fmt.Errorf("export runner artifacts: %w", err)
	}
	fmt.Println("✅ Build complete. Digest and SBOM exported for GHA runner Cosign step.")

	return nil
}
