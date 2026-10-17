# Design

## Context

See `proposal.md` for motivation and `specs/proxmox-provider-distribution/spec.md` for the canonical distribution contract. The current checkout's `origin` is already `https://github.com/mkdev-ai/devsy-proxmox-provider`; `go.mod`, Go imports, release linker flags, provider metadata, and README URLs still reference `github.com/iamveen/devpod-proxmox-provider`. The GitHub Actions release workflow already derives binary asset URLs from `github.repository`, but its linker flag is hard-coded.

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

1. **Use the existing canonical repository.** GitHub and this checkout's `origin` already identify `mkdev-ai/devsy-proxmox-provider`; do not rename the repository or alter its remote. The upstream `iamveen/devpod-proxmox-provider` stays independent and unchanged.

2. **Use one canonical Go import path.** Set the module path to `github.com/mkdev-ai/devsy-proxmox-provider`, update all first-party imports, and change the release linker flag that injects `cmd.versionStr`. This makes `go list`, builds, tests, and published version output agree with the renamed repository. Preserve dependency versions and binary asset filenames.

3. **Update only self-referencing distribution URLs.** Point `provider.yaml` metadata and README setup, clone, and install examples at the new canonical repo. Keep release URL generation based on `github.repository`, since it automatically follows the renamed repo; retain the current platform/checksum generation. Do not rewrite references whose purpose is explicitly to identify the separate upstream fork.

4. **Align the module and links before publishing.** Prepare and test module/import/link changes on a branch in the existing canonical repository. Do not change GitHub repository settings or the local `origin`. Do not create a release until module tests, installation from the canonical slug, and generated release URL/checksum checks pass.

## Risks / Trade-offs

- **Module consumers use the old upstream path** → Publish the canonical module path and provide a migration note; do not alter the upstream fork or imply it has moved.
- **Release is published before repository metadata is canonical** → Gate release creation on module tests, manifest URL checks, and a successful provider install from the new slug.
- **Old module/import paths are confused with the upstream fork** → Update maintained imports and docs to the canonical path, and keep explicit references to the upstream fork accurate and unchanged.

## Migration Plan

1. Update Go module/import/linker paths and self-referencing docs/manifest URLs to `github.com/mkdev-ai/devsy-proxmox-provider`; run Go tests and inspect generated provider metadata without publishing.
2. Verify the checkout remains connected to `https://github.com/mkdev-ai/devsy-proxmox-provider`, the upstream fork remote is unchanged, and provider installation plus generated release asset/checksum URLs resolve under the canonical slug.
3. Publish the next tagged release only after all module, install, canonical URL, and checksum checks pass. If validation fails, pause releases and correct the module or references without changing repository settings or moving the upstream fork.