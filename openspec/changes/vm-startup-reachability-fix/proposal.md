# Proposal: Gate VM readiness on qemu status == running (not the slow guest agent)

## Why
After a workspace VM is started, `WaitForVMReady` previously required BOTH a
running qemu status AND a responsive guest agent before the IP wait began.
On real cold boots the QEMU guest agent can take minutes to become alive, so
a perfectly healthy VM (qemu status == running) was hard-failed with
`verifying VM readiness: timed out waiting for VM readiness after 1m30s`
before the workspace was ever reported to Devsy.

The fix treats `qemu status == "running"` as the authoritative signal that
VM creation has completed. The guest agent is kept as a bounded probe inside
`WaitForIP` (Step 7), which retries on error and uses a longer bounded
timeout (~10m), so IP discovery never silently hangs either.

As a result:
- The VM is reported available right after the complete creation process
  (qemu status running == creation complete).
- IP discovery still discovers the IP reliably afterwards, never silently
  hanging (bounded retry + bounded timeout).

## What Changes
- `pkg/proxmox/poll.go`: `WaitForVMReady` now polls ONLY the qemu status
  endpoint and returns nil as soon as `status == "running"`. The guest-agent
  probe was removed from this gate.
- `pkg/proxmox/poll_test.go`: the test fixtures are updated to match the
  new semantics:
  - `TestWaitForVMReady_ReturnsNil_WhenRunning_Only` verifies readiness
    returns nil when qemu status is running (guest agent NOT required).
  - `TestWaitForVMReady_ReturnsError_WhenAgentUnavailableNoLongerFails`
    verifies the gate no longer hard-fails when the guest agent is slow
    (status=running wins; agent probed separately by WaitForIP).
  - `TestWaitForVMReady_ReturnsError_WhenStatusUnavailable` verifies that
    when qemu status is not running, the gate keeps polling and returns an
    error after the timeout.
- `cmd/create.go`: the Step 6 comment is updated to describe that the gate
  polls qemu status only; the guest agent is probed in the IP-wait step.
- `cmd/cmd_test.go`: `TestCreate_DevsyResultAndEnvironmentContract` now
  supplies a ready `VMStatus` (the only requirement of the new gate). The
  assertion that `GetNetworkInterfaces` was called is removed because the
  gate no longer requires it.

## Capabilities
`devsy-proxmox-provider`: provision and manage Proxmox workspace VMs while
reporting the VM as available immediately when qemu status reports running,
without waiting on the slow guest agent.

## Impact
- `pkg/proxmox/poll.go` (WaitForVMReady).
- `pkg/proxmox/poll_test.go` + `cmd/cmd_test.go` (test fixtures).
- `cmd/create.go` (comment only).
