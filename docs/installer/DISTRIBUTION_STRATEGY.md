# Distribution Strategy

## 1. Distribution model

VULNEX is distributed as:
1. Platform-specific CLI binary/package.
2. Versioned OCI application images in an official registry.
3. Release manifest, checksums, signatures, SBOM, and release notes.
4. Documentation and optional offline deployment bundle in a later phase.

Users do not need Git or source code to install a published release.

## 2. Initial artifacts

| Platform | Artifact |
|---|---|
| Linux amd64 | `vulnex-linux-amd64.tar.gz` |
| Linux arm64 | `vulnex-linux-arm64.tar.gz` |
| macOS arm64 | `vulnex-darwin-arm64.tar.gz` |
| macOS amd64 | `vulnex-darwin-amd64.tar.gz` |
| Windows amd64 | `vulnex-windows-amd64.zip` |
| Debian/Ubuntu | `.deb` package |
| RHEL-compatible | `.rpm` package |

Linux server is the first supported production target. macOS and Windows artifacts follow after runtime-specific integration tests.

## 3. OCI images

Separate images are recommended:
- `vulnex-api`
- `vulnex-web`
- `vulnex-worker`

Use immutable digests in generated deployment manifests. Human-readable version tags are convenience labels, not the trust anchor.

## 4. Build and release

- GoReleaser for CLI binaries and packages.
- GitHub Actions for CI and release orchestration.
- Build frontend and backend images in reproducible pipelines.
- Generate SBOM for CLI and container images.
- Sign artifacts and images.
- Publish SHA-256 checksums.
- Pin dependencies and scanner versions.
- Scan images and dependencies before release.

## 5. Verification

Installer must verify the release manifest signature and artifact digest before deployment. Failure to verify is a hard failure; do not offer an insecure bypass in normal production workflows.

## 6. Offline deployment

Post-MVP capability:
- Export a bundle containing signed images, manifest, compose specification, and CLI.
- Import and verify bundle on an offline server.
- Clearly document that offline bundle freshness depends on the date it was created.

## 7. Licensing

Project license target is Apache-2.0. Maintain third-party notices and review the licenses of scanner tools, templates, libraries, and container base images. Distribution rights must be verified before bundling or mirroring third-party tools.
