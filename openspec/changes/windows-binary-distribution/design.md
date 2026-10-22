# Design

## Context

`provider.yaml` defines the distribution manifest Devsy installs under the
`PROXMOX_PROVIDER` binary name. Its `binaries` section maps
`{os, arch, path, checksum}`. The release workflow builds:

```bash
PLATFORMS="linux/amd64 linux/arm64 darwin/amd64 darwin/arm64"
```

and emits `dist/proxmox-provider-{os}-{arch}` plus `dist/proxmox-provider-{os}-{arch}.sha256`
for each, then uploads those files (plus `provider.yaml`) to the GitHub Release
for the tag.

The current manifest advertises only linux/macOS. Windows users see
`cannot download provider binary PROXMOX_PROVIDER, because no binary location
matched OS windows and ARCH amd64`.

## Upstream release asset inventory (v0.2.0)

Verified via the GitHub Releases API:

| asset name |
|---|
| provider.yaml |
| proxmox-provider-darwin-amd64 |
| proxmox-provider-darwin-arm64 |
| proxmox-provider-linux-amd64 |
| proxmox-provider-linux-arm64 |

There is no `proxmox-provider-windows-amd64` or
`proxmox-provider-windows-arm64` asset yet, but the build matrix supports
windows and the naming convention for each platform is
`proxmox-provider-{os}-{arch}` (as used by linux and darwin). So the Windows
assets to expect at publish time are:

- `proxmox-provider-windows-amd64`
- `proxmox-provider-windows-arm64`

## Root cause (short)

- Manifest is missing windows entries (both amd64 and arm64).
- The linux/darwin asset URLs in the manifest do not match the actual upstream
  release asset names. The manifest uses bare download names
  (`linux-amd64`, `darwin-amd64`), while the real release assets are
  `proxmox-provider-linux-amd64` / `proxmox-provider-darwin-amd64`. A user or
  the CLI following these URLs would 404.

## Design decisions

1. **Advertise all four built platforms in the manifest.** Add `windows/amd64`
   and `windows/arm64` entries with the real upstream release asset naming and
   checksums, so the manifest is the source of truth for what the release
   publishes.

2. **Use the actual upstream release asset convention.**
   `proxmox-provider-{os}-{arch}` (e.g. `proxmox-provider-linux-amd64`,
   `proxmox-provider-darwin-arm64`). Do not reuse the old
   `linux-amd64`-style bare names.

3. **Keep the build matrix at 4 platforms** in
   `.github/workflows/release.yml`. No change needed there; the manifest must
   simply reflect what the workflow already produces.

4. **Checksum alignment.** Each manifest entry carries a SHA-256 checksum. The
   release workflow computes `sha256sum dist/proxmox-provider-{os}-{arch}` per
   platform and embeds it via `sed` from the `.sha256` file. The manifest must
   use the same checksum values the workflow writes, so the CLI can verify
   downloads. For the two windows entries, the checksums are generated at
   publish time from the built binaries; this change pins the expected names
   and the workflow derives the values deterministically.

5. **Source of truth.** `provider.yaml.tmpl` is the template the workflow
   `sed`s into `provider.yaml`. Add the windows entries to the template; the
   generated `provider.yaml` follows. Do not hand-edit `provider.yaml`.

## Risks / Trade-offs

- **Checksum drift until a tagged release runs.** The windows checksums are
  produced by the release workflow at publish time, so a newly tagged release
  must be generated before consumers can verify Windows downloads. This change
  is safe: the manifest names the correct assets and the workflow fills the
  values; a consumer installing an un-tagged build gets a download error, not a
  silent verification failure.
- **CLI may not yet accept windows.** The upstream provider binary is a
  CGO_ENABLED=0 Go binary and the naming matches existing linux/darwin assets,
  so adoption is purely a distribution-manifest change. No code change in this
  repo is required to make Windows installable.

## Open Questions

- Confirm the windows builds the release workflow produces are byte-stable
  (`CGO_ENABLED=0`, same Go toolchain) so the generated checksums are
  deterministic across runners. (Already the case for linux and darwin in the
  same workflow.)
