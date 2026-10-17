# Tasks

## 1. Canonical Go Module Identity

- [ ] 1.1 Confirm `mkdev-ai/devsy-proxmox-provider` is available for rename and the maintainer has GitHub repository-admin access; verify the upstream fork remains a separate remote/repository before changing repository settings.
- [ ] 1.2 Change `go.mod` to `github.com/mkdev-ai/devsy-proxmox-provider`, update all first-party imports and the release `-ldflags` symbol path; verify `go list -m` reports the canonical module and `go test ./...` passes.
- [ ] 1.3 After tasks 1.1–1.2 pass, commit only the Go module/import/linker updates as `build: use canonical Devsy provider module path`; verify the commit includes the module, import, and build metadata changes only.

## 2. Provider Metadata and Release References

- [ ] 2.1 Update `provider.yaml.tmpl` home and README clone/install/setup links to `mkdev-ai/devsy-proxmox-provider`; verify no maintained self-reference points at the upstream fork, while any explicit upstream-fork reference remains unchanged.
- [ ] 2.2 Generate a release manifest from the workflow inputs and verify all Linux/macOS amd64/arm64 asset URLs and SHA-256 checksums point at the canonical repository; verify asset names and version behavior remain unchanged.
- [ ] 2.3 After tasks 2.1–2.2 pass, commit only provider metadata, documentation links, and release changes as `docs(release): use canonical Devsy repository URLs`; verify no unrelated files are included.

## 3. Repository Rename and Publication

- [ ] 3.1 Merge the prepared module/metadata changes, rename the current GitHub repository from `mkdev-ai/devspy-proxmox-provider` to `mkdev-ai/devsy-proxmox-provider`, then update this clone's `origin` only after the new canonical URL resolves; verify the former current slug redirects and the upstream fork URL is unchanged.
- [ ] 3.2 Install the provider from `github.com/mkdev-ai/devsy-proxmox-provider` and verify its manifest and binary checksum resolve successfully; do not publish a tagged release until this passes.
- [ ] 3.3 Publish a tagged release from the renamed repository and verify the release assets and generated provider manifest are served from the canonical URL without changing supported platforms or asset names.
- [ ] 3.4 If post-rename verification requires source fixes, commit them separately with a message naming the exact failing URL/import or release behavior; verify each commit contains only the fix for that failure.