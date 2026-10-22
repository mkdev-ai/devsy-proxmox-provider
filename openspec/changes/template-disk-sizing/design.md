# Design

## Context

See `proposal.md` for the root-cause analysis. The failing point is the template base disk, not the agent's local data root — the agent's data root is a separate filesystem and is unrelated to the Proxmox template. The user pointed this out and it is correct.

## Goals / Non-Goals

**Goals:**
- The template base disk is created at a configurable size (`--template-disk-size`), so templates built from the Ubuntu cloud image are not stuck at ~3.5 GB.
- The cloned workspace VM is resized to exactly `VM_DISK_SIZE` **before** the VM starts, so the git clone and Docker install that follow always have adequate capacity.
- The change is spec-ified, traceable in `tasks.md`, and verified with the existing test suite.

**Non-Goals:**
- Changing the agent's local data root (a separate filesystem, controlled by `DEVSY_HOME`/`--home`), which is unrelated to the Proxmox template.
- Modifying Proxmox provisioning/lifecycle logic beyond the clone resize.
- Silently falling back to legacy DevPod-branded resources.

## Decisions

1. **Configure template size at template-creation time.** Pass the requested size to `qm importdisk --size` in `setup.sh`. The default stays `3584` MiB (the Ubuntu cloud image default) so rebuilds are backward compatible; set a larger value when you want a bigger template.
2. **Resize the clone to an absolute target before start, not additively afterward.** Proxmox's `qm resize(1)` treats a bare value (no `+`) as an absolute target size. The original code did `DiskSize - templateDiskGB` and applied `+33G`, which is both wrong (it assumed 17 GB) and too late (after the clone). Now the clone is set to exactly `VM_DISK_SIZE` before start.
3. **Document the root-run rebuild step.** The template change only takes effect after re-running `setup.sh` on the Proxmox host; `devsy provider add` / `devsy workspace up` reads the existing template.
4. **Keep the template and the VM disk conceptually separate in docs/UI.** `VM_DISK_SIZE` is the workspace VM disk size; `template-disk-size` is the base. Both are explicit so the user is not misled.

## Risks / Trade-offs

- `qm importdisk` is an additive template build; changing the size requires a full template rebuild on the host. Document this clearly so users do not expect to resize an existing template in place.
- The default `3584` MiB preserves the old behavior for anyone who rebuilds the template without the flag.
- This change does not auto-resize a template that the user already built; they must re-run `setup.sh` with the desired size. The provider's `VM_DISK_SIZE` resize-then-guard removes the risk of an undersized but intended template.

## Open Questions

- Confirm the installed Proxmox `qm importdisk` accepts `--size` in MiB on the target host (derivable from `qm importdisk -h` / `qm(1)`). The change is documented as pending this verification; once confirmed, the task can be marked done in a live context.
