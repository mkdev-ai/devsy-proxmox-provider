# Spec Delta

## Purpose

Defines the distribution contract for the Devsy Proxmox provider so that the
`provider.yaml` manifest advertises every platform the release workflow builds
and references the corresponding upstream GitHub Release assets with
checksum-verified binary URLs.

## CHANGED REQUIREMENTS

### Requirement: Manifest advertises all four built platforms

The provider manifest SHALL list a binary location for every platform the
release workflow builds: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64,
and windows/amd64, windows/arm64. Each entry SHALL use the upstream release
asset naming `proxmox-provider-{os}-{arch}`.

#### Scenario: Windows binary is advertised on amd64

- **WHEN** the manifest is read for a Windows amd64 host
- **THEN** a binary location for `os: windows, arch: amd64` is present with
  `path: https://github.com/mkdev-ai/devsy-proxmox-provider/releases/download/v<version>/proxmox-provider-windows-amd64`

#### Scenario: Windows binary is advertised on arm64

- **WHEN** the manifest is read for a Windows arm64 host
- **THEN** a binary location for `os: windows, arch: arm64` is present with
  `path: https://github.com/mkdev-ai/devsy-proxmox-provider/releases/download/v<version>/proxmox-provider-windows-arm64`

### Requirement: Manifest asset URLs match upstream release assets

Every binary entry's `path` SHALL reference the exact GitHub Release asset name
and SHA-256 checksum that the release workflow publishes for that platform.

#### Scenario: Download resolves to the published asset

- **WHEN** a user installs the provider on a supported platform
- **THEN** the manifest URL for that platform resolves to the published release
  asset and the declared checksum matches the uploaded asset's SHA-256.

### Requirement: Provider binary is cross-platform

The provider binary SHALL compile and run on linux/amd64, linux/arm64,
darwin/amd64, darwin/arm64, windows/amd64, and windows/arm64.

#### Scenario: Windows binaries build and run

- **WHEN** the provider is built for windows/amd64 and windows/arm64
- **THEN** the resulting binaries are executable and expose a version/help
  command (builds with `GOOS=windows` succeed, and the binary runs without a
  DOS-box dependency requiring CGO).
