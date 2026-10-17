# Proposal

## Why

The canonical GitHub repository is already `mkdev-ai/devsy-proxmox-provider`, but the Go module path, provider metadata, and installation links still point to the upstream fork `iamveen/devpod-proxmox-provider`. Align the module and self-referencing distribution links so users install the maintained Devsy provider from the canonical repository.

## What Changes

- Use the existing canonical repository `mkdev-ai/devsy-proxmox-provider`; keep the upstream fork `iamveen/devpod-proxmox-provider` unchanged.
- Change the Go module path, all Go imports, and release linker references to `github.com/mkdev-ai/devsy-proxmox-provider`.
- Point provider metadata, README clone/install/setup links, and other self-referencing URLs at the renamed canonical repository.
- Keep release asset names, supported platforms, checksums, and versioning behavior stable; verify release generation uses the existing canonical repository.

## Capabilities

### New Capabilities

- `proxmox-provider-distribution`: Install and retrieve checksum-verified provider releases from the canonical Devsy Proxmox provider repository.

### Modified Capabilities

- None.

## Impact

Affected areas include `go.mod`, Go imports, `.github/workflows/release.yml`, `provider.yaml.tmpl`, and README/setup distribution links. The GitHub repository and local `origin` are already canonical; this change must not rename them or modify the upstream fork. This change is separate from the Devsy provider runtime migration.