# Design: Gate VM readiness on qemu status == running

## Context
Bug: after a cold boot, the QEMU guest agent takes well over 90s to answer.
The previous `WaitForVMReady` gate required both a running qemu status AND
a responsive guest agent. Because the gate waited for the guest agent, a
perfectly healthy VM (qemu status == running) was hard-failed with
`verifying VM readiness: timed out waiting for VM readiness after 1m30s`
before the workspace was ever reported to Devsy.

## Goals
- The VM is reported available right after the complete VM creation process,
  where "creation complete" is defined by qemu status == running (the
  authoritative signal that the hypervisor has finished creating/starting
  the VM).
- IP discovery still discovers the IP reliably afterwards, and never
  silently hangs: `WaitForIP` (Step 7) still probes the guest agent with a
  bounded retry and a bounded timeout (~10m).

## Non-goals
- Modifying the guest-agent probe itself (it stays as-is in WaitForIP).
- Changing the total latency of create for a healthy VM.

## Decisions
1. **Reachability = qemu status == "running".** The qemu status read is the
   authoritative signal that the hypervisor has finished creating/starting
   the VM. `WaitForVMReady` polls only `GetVMStatus` and returns nil as soon
   as the status is running. No guest-agent read is required here.
2. **Guest agent stays as a bounded probe inside WaitForIP.** `WaitForIP`
   (Step 7) retains its `WaitForAgent` first (30s) and then the IP poll with
   a bounded timeout (10m). This guarantees IP discovery is still reliable
   and never silently hangs.
3. **Timeout is an error, not a silent skip.** `WaitForVMReady` returns an
   error if the VM does not reach running within the timeout.
4. **Test semantics.** The mock-based unit tests are updated so the gate is
   verified to return nil on status=running even when the agent is slow
   (`TestWaitForVMReady_ReturnsError_WhenAgentUnavailableNoLongerFails`),
   and to keep polling and erroring when status is not running
   (`TestWaitForVMReady_ReturnsError_WhenStatusUnavailable`).
5. **Create-contract test.** `TestCreate_DevsyResultAndEnvironmentContract`
   supplies a ready `VMStatus` (the only requirement of the new gate); the
   assertion that `GetNetworkInterfaces` was recorded is removed because the
   gate no longer requires it.

## Risks
- **VM reported available before the guest agent is ready.** This is
   intentional: the downstream IP wait still requires the guest agent and a
   leased IPv4, so a VM that is "available" but has no IP will still be
   transparently retried by the IP wait (bounded to ~10m).

## Verification
- `go build ./...` passes.
- `go test ./...` passes; the existing.mock contract test
  `TestWaitForVMReady_ReturnsNil_WhenRunning_Only` (mocks VMStatus running +
  no Ifaces) still passes.
- `TestCreate_DevsyResultAndEnvironmentContract` still produces the Devsy
  result JSON with `ipAddress=192.0.2.10`.
