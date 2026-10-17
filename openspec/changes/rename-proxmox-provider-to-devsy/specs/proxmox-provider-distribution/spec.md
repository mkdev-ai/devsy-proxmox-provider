# Spec Delta

## Purpose

Defines the canonical repository identity and release availability users rely on to install and update the Devsy Proxmox provider from the maintained project.

## ADDED Requirements

### Requirement: Canonical provider repository

The provider's canonical source, metadata, and documented installation instructions SHALL identify `mkdev-ai/devsy-proxmox-provider`; the upstream fork `iamveen/devpod-proxmox-provider` SHALL remain a separate, unchanged repository.

#### Scenario: Install the provider from its canonical source
- **WHEN** a user follows the provider installation instructions
- **THEN** the Devsy CLI retrieves the provider from `github.com/mkdev-ai/devsy-proxmox-provider`

#### Scenario: Resolve the canonical repository
- **WHEN** a user opens the canonical repository URL
- **THEN** it resolves to this provider repository, while the upstream fork URL continues to resolve to the upstream fork

### Requirement: Canonical checksum-verified releases

Each platform advertised by the provider manifest SHALL have a binary release asset and matching SHA-256 checksum hosted under the canonical repository, and the generated provider manifest SHALL reference those exact assets and checksums.

#### Scenario: Install a release on a supported platform
- **WHEN** Devsy installs a published provider version on a platform advertised by the manifest
- **THEN** it downloads that platform's binary from the canonical repository and can verify it with the matching SHA-256 checksum

#### Scenario: Generate release metadata
- **WHEN** the release workflow builds and publishes a tagged version
- **THEN** the provider manifest uses the canonical repository for all supported platform asset URLs and includes each generated checksum