# Design

## Context
See `proposal.md` for motivation and `specs/delete-machine/delete-machine.md` for
the behavior contract. The live logs prove the current delete path calls
`POST /nodes/sojki/qemu/3006/destroy`, which Proxmox does not implement and
returns HTTP 501. The API that actually works is `DELETE /nodes/{node}/qemu/{vmid}`
with `purge=1` (and optionally `destroy-unreferenced-disks=1`), matching
`qm destroy --purge`.

## Goals / Non-Goals

**Goals:**
- `devsy provider delete` (and `devsy up` cleanup) exits 0 when the machine and
  its VM are removed on a real Proxmox VE host.
- `DeleteVM` calls the correct Proxmox API: `DELETE /nodes/{node}/qemu/{vmid}`,
  with `purge=1` and `destroy-unreferenced-disks=1`.
- The shutdown task is actually waited for before destroy (no transient 500s).
- Transient Proxmox 500s on destroy are retried with backoff; the final error is
  actionable.

**Non-Goals:**
- Changing how Devsy names the provider binary or hosts it.
- Multi-cluster / cross-node deletion.
- Handling a VM stuck in a crashed state forever (timed out, not fixed).

## Decisions
1. **Use `DELETE /nodes/{node}/qemu/{vmid}` with purge, not POST
   `/destroy`.** The endpoint Proxmox implements is the standard `qm destroy`
   API. The provider must call it with `purge=1` and
   `destroy-unreferenced-disks=1`; method POST on `/destroy` yields HTTP 501.
2. **Wait for shutdown before destroy.** Replace `time.Sleep(2 * time.Second)`
   + ignored `StopVM` with a real `WaitForTask` on the shutdown UPID. Only call
   `DeleteVM(..., true)` after the task reports `stopped` + `OK`.
3. **Retry transient destroy errors.** Add fixed-backoff (1s/2s/4s, max 3)
   retries for HTTP 5xx / API errors on the destroy request. Errors are wrapped
   with the Proxmox HTTP code and message so the user sees what actually failed.
4. **Emergency cleanup on transient failure.** After a destroy failure, attempt
   to stop (ignore 404) then destroy without purge, so a stuck VM is not left
   behind.
5. **Verify with the existing env-gated acceptance suite.** Extend
   `TestDevsyAcceptance` to start a devsy-named VM, then delete it, asserting
   the machine is gone with no leftover VM and no orphaned finalizer.

## Risks / Trade-offs
- **Wrong HTTP method forgotten** → The contract test `TestDeleteVM_RunningVM`
  asserts the path is NOT `/destroy` and that `DELETE /nodes/{node}/qemu/{vmid}`
  is called with `purge=1` + `destroy-unreferenced-disks=1`, so a regression is
  caught immediately.
- **Destroy races on a running VM** → The shutdown-wait decision means we never
  destroy a running VM, avoiding the transient 500 that the old code produced.
- **Emergency cleanup is a best-effort safety net** → It is logged but does not
  override the primary failed path; the caller still sees the root cause.

## Migration Plan
None needed for existing machines: this change fixes delete on machines the
provider already manages. It does not touch named VMs, finalizers, or machine
objects that the provider already manages.
