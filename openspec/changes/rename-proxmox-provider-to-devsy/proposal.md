# Proposal

## Why

This checkout's canonical GitHub repository is `mkdev-ai/devspy-proxmox-provider`, but its Go module path, provider metadata, and installation links still point to the upstream fork `iamveen/devpod-proxmox-provider`. Rename the current repository to `mkdev-ai/devsy-proxmox-provider` and align its module and self-referencing distribution links so users install the maintained Devsy provider from the correct source.

## What Changes

- **BREAKING** Rename the current GitHub repository from `mkdev-ai/devspy-proxmox-provider` to `mkdev-ai/devsy-proxmox-provider`; keep the upstream fork `iamveen/devpod-proxmox-provider` unchanged.
- Change the Go module path, all Go imports, and release linker references to `github.com/mkdev-ai/devsy-proxmox-provider`.
- Point provider metadata, README clone/install/setup links, and other self-referencing URLs at the renamed canonical repository.
- Keep release asset names, supported platforms, checksums, and versioning behavior stable; verify release generation uses the canonical repository after the rename.

## Capabilities

### New Capabilities

- `proxmox-provider-distribution`: Install and retrieve checksum-verified provider releases from the canonical Devsy Proxmox provider repository.

### Modified Capabilities

- None.

## Impact

Affected areas include GitHub repository settings and the local `origin` URL, `go.mod`, Go imports, `.github/workflows/release.yml`, `provider.yaml.tmpl`, and README/setup distribution links. The rename must preserve the upstream fork and coordinate its redirect behavior with existing clones and module consumers. This change is separate from the Devsy provider runtime migration.