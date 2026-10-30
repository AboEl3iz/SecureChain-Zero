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
		outPath := fmt.Sprintf("/out/server-%s", t.GoArch)
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
	apkoContainer := client.Container().
		From("ghcr.io/chainguard-dev/apko:v0.19.4").
		WithMountedDirectory("/work", src).
		WithMountedFile("/work/bin/server", binary).
		WithWorkdir("/work").
		WithExec([]string{
			"apko", "build",
			"--log-level=info",
			"--arch", apkArch,
			"--sbom-formats=cyclonedx,spdx",
			"build/apko.yaml",
			"ghcr.io/myorg/secure-app:build-" + goArch,
			"/out/image-" + goArch + ".tar",
			"--sbom-path=/out/sbom-" + goArch,
		})

	imageFile := apkoContainer.File("/out/image-" + goArch + ".tar")
	sbomFile := apkoContainer.File("/out/sbom-" + goArch + "/cyclonedx.json")

	return imageFile, sbomFile, nil
}
