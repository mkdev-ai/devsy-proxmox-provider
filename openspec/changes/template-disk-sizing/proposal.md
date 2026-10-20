# Proposal

## Why

The template disk the workspace is built from is far too small. `scripts/setup.sh` creates the cloud-init template from the Ubuntu cloud image with no size argument, so the template lands at the image's default — 3584 M (~3.5 GB), as the Proxmox web UI screenshot of `devsy-ubuntu-24-04` (ID 9000) confirms. The provider's `VM_DISK_SIZE` option only resizes the *clone* after cloning, and it did so against a hard-coded "expect ~17 GB template" assumption, so it could never produce a 50 GB workspace. The result is that the git clone / docker pull that follows the VM start fills the small disk and fails with `fatal: write error: No space left on device`.

The workspace should be created with the correct capacity **before** anything is fetched, so the git clone and Docker install always have room.

## What Changes

- Add a `--template-disk-size` (MiB) option to `scripts/setup.sh` so the template base disk is created at the requested capacity instead of the 3.5 GB Ubuntu cloud image default.
- In `cmd/create.go`, resize the clone to exactly `VM_DISK_SIZE` computed from the real template base, **before** the VM starts (not after, and not a brittle `+increase` calc from a guessed 17 GB).
- Record the behavior contract in the OpenSpec spec (`specs/devsy-proxmox-provider/spec.md`) and a task slice in `tasks.md`.

## Capabilities

### New Capabilities

- `template-disk-sizing`: Create the Proxmox template at a configurable disk size and resize the cloned workspace VM to `VM_DISK_SIZE` before install.

## Impact

Affected areas include `scripts/setup.sh`, `cmd/create.go`, the OpenSpec spec and tasks, and user-facing docs. Requires root on the Proxmox host to rebuild the template (the `qm importdisk --size` flag is supported by Proxmox; verify against the installed `qm` man page on the target host).

