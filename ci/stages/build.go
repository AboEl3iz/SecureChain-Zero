package stages

import (
	"context"
	"fmt"

	"dagger.io/dagger"
)

// ArchTarget defines cross-compilation mappings between Go and Wolfi/apko.
type ArchTarget struct {
	GoArch  string // Go environment value (amd64, arm64)
	ApkArch string // apko/Wolfi arch name (x86_64, aarch64)
}

// BuiltArch contains build artifacts for a specific architecture.
type BuiltArch struct {
	Target ArchTarget
	Binary *dagger.File
	Image  *dagger.File
	SBOM   *dagger.File
}

// BuildMultiArchBinaries cross-compiles static Go binaries with hardened flags.
func BuildMultiArchBinaries(
	client *dagger.Client,
	src *dagger.Directory,
	targets []ArchTarget,
) map[string]*dagger.File {
	binaries := make(map[string]*dagger.File)

	goBuilder := client.Container().
		From("golang:1.23.4-alpine").
		WithMountedDirectory("/src", src).
		WithWorkdir("/src/sample-app").
		WithEnvVariable("CGO_ENABLED", "0").
		WithEnvVariable("GOOS", "linux")

	for _, t := range targets {
		outPath := fmt.Sprintf("/tmp/server-%s", t.GoArch)
		bin := goBuilder.
			WithEnvVariable("GOARCH", t.GoArch).
			WithExec([]string{
				"go", "build",
				"-trimpath",
				"-ldflags=-s -w -extldflags=-static",
				"-o", outPath,
				".",
			}).
			File(outPath)

		binaries[t.GoArch] = bin
	}

	return binaries
}

// BuildApkoImage builds an isolated OCI image and generates CycloneDX/SPDX SBOMs.
func BuildApkoImage(
	ctx context.Context,
	client *dagger.Client,
	src *dagger.Directory,
	binary *dagger.File,
	apkArch string,
	goArch string,
) (*dagger.File, *dagger.File, error) {
	outTar := fmt.Sprintf("/tmp/base-%s.tar", goArch)

	// 1. Build hermetic Wolfi base image using modern Chainguard apko
	apkoContainer := client.Container().
		From("cgr.dev/chainguard/apko:latest").
		WithMountedDirectory("/work", src).
		WithWorkdir("/work").
		WithExec([]string{
			"apko", "build",
			"--log-level=info",
			"--arch", apkArch,
			"build/apko.yaml",
			"ghcr.io/myorg/secure-app:build-" + goArch,
			outTar,
		})

	baseTar := apkoContainer.File(outTar)

	// 2. Inject compiled multi-arch Go binary into /app/server (non-root UID 65532)
	finalContainer := client.Container(dagger.ContainerOpts{Platform: dagger.Platform("linux/" + goArch)}).
		Import(baseTar).
		WithFile("/app/server", binary, dagger.ContainerWithFileOpts{
			Permissions: 0755,
			Owner:       "65532:65532",
		})

	imageTar := finalContainer.AsTarball()

	// 3. Generate CycloneDX SBOM via Syft for NIST & EO 14028 compliance
	syftContainer := client.Container().
		From("anchore/syft:v1.11.1").
		WithMountedFile("/image.tar", imageTar).
		WithExec([]string{
			"syft", "scan",
			"/image.tar",
			"-o", "cyclonedx-json=/tmp/cyclonedx.json",
		})

	sbomFile := syftContainer.File("/tmp/cyclonedx.json")

	return imageTar, sbomFile, nil
}
