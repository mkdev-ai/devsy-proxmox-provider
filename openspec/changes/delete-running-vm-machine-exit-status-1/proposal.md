# Proposal

## Why
The canonical GitHub repository is already `mkdev-ai/devsy-proxmox-provider`, but
running `devsy up` (or an explicit provider delete) on a live Proxmox VE machine
ends with the user seeing:

```
errorfailed to delete machine: machineId=python-af390, err=run provider command
"${PROXMOX_PROVIDER} delete" (): exit status 1
```

Live logs from the failing host show the actual failure:

```
Error: deleting VM: API error: HTTP 501 — {"message":"Method 'POST
/nodes/sojki/qemu/3006/destroy' not implemented","data":null}
```

The provider's `delete` command (`cmd/lifecycle.go` `runDelete`) returns a
non-nil error, which cobra turns into a non-zero process exit, so Devsy reports
`exit status 1`.

## What Changes
- `pkg/proxmox/client.go` `DeleteVM`:
  - Change the request from `POST /nodes/{node}/qemu/{vmid}/destroy` to
    `DELETE /nodes/{node}/qemu/{vmid}` (the endpoint Proxmox actually implements).
  - Add query params `purge=1` and `destroy-unreferenced-disks=1` so the VM and
    its disks are removed.
  - Add a transient-error retry (fixed backoff 1s / 2s / 4s, max 3 attempts) for
    HTTP 5xx / API errors on the destroy request.
  - Surface the Proxmox API error detail in the wrapped error.
- `cmd/lifecycle.go` `runDelete`:
  - Remove the `time.Sleep(2 * time.Second)` hack and the ignored `StopVM` call.
  - Wait for the shutdown task with `WaitForTask` (capped at ~30s) before
    destroying.
  - Only call `DeleteVM(..., true)` after the shutdown task reports `stopped` +
    `OK`.
  - Keep the emergency-cleanup step (stop, then destroy without purge) after a
    transient destroy failure.
- Tests: contract cases for a running VM, a transient 500, and an end-to-end
  `TestRunDelete`; extend the env-gated acceptance suite.
- `README.md`: document that delete works for running machines and that
  emergency cleanup is available on failure.

## Capabilities

### New Capabilities
- `devsy-proxmox-provider-delete-machine`: Delete a devsy workspace machine on a
  real Proxmox VE host, returning success (exit 0) when the VM and disks are
  removed.

### Modified Capabilities
- None.

## Impact
Affected files: `pkg/proxmox/client.go`, `cmd/lifecycle.go`, `pkg/proxmox`
test helpers, `cmd/cmd_test.go`, `README.md`. This change is entirely additive
and safe: a failing delete still leaves the machine intact (with an
emergency-cleanup fallback that first stops then destroys without purge), so no
running machine is ever dropped silently.
