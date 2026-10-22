# Tasks

## 1. Add windows binaries to the template

- [ ] 1.1 Add `windows/amd64` and `windows/arm64` entries to
  `provider.yaml.tmpl` under `binaries.PROXMOX_PROVIDER`, using the upstream
  release asset naming `proxmox-provider-{os}-{arch}` and placeholder checksums.
- [ ] 1.2 Verify the template now has 4 platform entries.

## 2. Align asset names/checksums with the manifest contract

- [ ] 2.1 Confirm the release workflow build matrix covers exactly the 4
  platforms the manifest advertises (linux amd64/arm64, darwin amd64/arm64).
- [ ] 2.2 Confirm the release workflow emits `dist/proxmox-provider-{os}-{arch}`
  and `.sha256` for each, and uploads them to the GitHub Release.

## 3. Verify cross-platform install behavior (local)

- [ ] 3.1 Build the provider for windows/amd64 and windows/arm64 locally and
  confirm they run (e.g. `GOOS=windows GOARCH=amd64 go build` plus a smoke
  `./proxmox-provider version`/help invocation).
- [ ] 3.2 Confirm the manifest `path` URLs resolve against the upstream release
  asset naming convention `proxmox-provider-{os}-{arch}` for all 4 platforms.

## 4. OpenSpec artifacts

- [ ] 4.1 Create the change with `openspec new change windows-binary-distribution`
  (done in this worktree).
- [ ] 4.2 Write `proposal.md`, `design.md`, `tasks.md`, and a spec under
  `specs/proxmox-provider-distribution/`.
- [ ] 4.3 Run `openspec validate change` and `openspec status --change` in this
  worktree.

## 5. Implementation + integration (PR to main)

- [ ] 5.1 Commit the manifest changes on
  `feature/windows-binary-distribution`.
- [ ] 5.2 Push the branch and open a PR to `main`.
- [ ] 5.3 After merge, a tagged release regenerates `provider.yaml` with the
  windows entries and the correct checksums.
