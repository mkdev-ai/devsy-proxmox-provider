# Design

## Context

See `proposal.md` for motivation and `specs/proxmox-provider-distribution/spec.md` for the canonical distribution contract. The current checkout's `origin` is `https://github.com/mkdev-ai/devspy-proxmox-provider`; `go.mod`, Go imports, release linker flags, provider metadata, and README URLs still reference `github.com/iamveen/devpod-proxmox-provider`. The GitHub Actions release workflow already derives binary asset URLs from `github.repository`, but its linker flag is hard-coded.

## Goals / Non-Goals

**Goals:**
- Make `mkdev-ai/devsy-proxmox-provider` the repository, Go module, and user-facing distribution identity.
- Preserve release asset names, supported OS/architectures, and SHA-256 verification behavior.
- Leave `iamveen/devpod-proxmox-provider` unchanged as the upstream fork.

**Non-Goals:**
- Change Proxmox resource behavior or Devsy provider runtime behavior; those belong to `migrate-proxmox-provider-to-devsy`.
- Rename or modify the upstream fork.
- Publish a release from the old slug after the repository rename.

## Decisions

1. **Rename the current repository, not the upstream fork.** Change the existing `mkdev-ai/devspy-proxmox-provider` repository name to `mkdev-ai/devsy-proxmox-provider`. GitHub's redirect from the former current slug may preserve existing links, but the new slug is canonical and all maintained references must use it. The upstream `iamveen/devpod-proxmox-provider` stays independent and unchanged.

2. **Use one canonical Go import path.** Set the module path to `github.com/mkdev-ai/devsy-proxmox-provider`, update all first-party imports, and change the release linker flag that injects `cmd.versionStr`. This makes `go list`, builds, tests, and published version output agree with the renamed repository. Preserve dependency versions and binary asset filenames.

3. **Update only self-referencing distribution URLs.** Point `provider.yaml` metadata and README setup, clone, and install examples at the new canonical repo. Keep release URL generation based on `github.repository`, since it automatically follows the renamed repo; retain the current platform/checksum generation. Do not rewrite references whose purpose is explicitly to identify the separate upstream fork.

4. **Sequence the rename before publishing.** Prepare and test code/link changes on a branch, merge them while repository access is available, perform the GitHub rename, then update the local `origin` URL and validate installation/release metadata under the new path. Do not create a release until the rename and canonical URL checks pass. GitHub admin access is required for the repository rename; the plan must identify that as an operator action, not simulate it by changing the local remote.

## Risks / Trade-offs

- **GitHub rename redirects are unavailable or conflict with an existing repository** → Check destination availability before renaming and verify both the new canonical path and old-current-slug redirect immediately after the operation.
- **Module consumers use the old upstream path** → Publish the canonical module path and provide a migration note; do not alter the upstream fork or imply it has moved.
- **Release is published before repository metadata is canonical** → Gate release creation on module tests, manifest URL checks, and a successful provider install from the new slug.
- **Local clones retain the old `origin`** → Update this checkout's `origin` only after confirming the GitHub rename succeeded; preserve other remotes, especially any upstream remote.

## Migration Plan

1. Update Go module/import/linker paths and self-referencing docs/manifest URLs to `github.com/mkdev-ai/devsy-proxmox-provider`; run Go tests and inspect generated provider metadata without publishing.
2. Confirm admin access and destination slug availability, merge the prepared changes, and rename the current GitHub repository from `devspy-proxmox-provider` to `devsy-proxmox-provider`.
3. Update this clone's `origin` to the new URL without touching the upstream fork remote. Verify `go test ./...`, provider install from the canonical slug, release asset URLs, and the redirect from the prior current slug.
4. Publish the next tagged release only after all canonical URL and checksum checks pass. If the rename or validation fails, pause releases and restore access using GitHub's rename controls; do not move or rename the upstream fork as a rollback.