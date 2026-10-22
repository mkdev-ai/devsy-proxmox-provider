# Tasks

## 1. Rewrite WaitForVMReady to gate ONLY on qemu status
- [x] 1.1 Replace the existing gate (both qemu status + guest agent must
       respond) with a status-only gate:
       poll `GetVMStatus`; return nil as soon as `status == "running"`.
       Remove the guest-agent read from this function.
- [x] 1.2 Update the doc comment on `WaitForVMReady` to describe the new
       status-only semantics.

## 2. Update tests to match the new gate semantics
- [x] 2.1 Update `TestWaitForVMReady_ReturnsNil_WhenBothReadsSucceed` to
       `TestWaitForVMReady_ReturnsNil_WhenRunning_Only`: supply only
       `VMStatus` (no `Ifaces`), and assert `GetVMStatus` was called.
- [x] 2.2 Rename and update `TestWaitForVMReady_ReturnsError_WhenAgentUnavailable`
       to `TestWaitForVMReady_ReturnsError_WhenAgentUnavailableNoLongerFails`:
       verify the gate no longer hard-fails when the guest agent is slow
       (status=running wins; agent probed separately by WaitForIP).
- [x] 2.3 Update `TestWaitForVMReady_ReturnsError_WhenStatusUnavailable` to
       not supply `Ifaces`.
- [x] 2.4 Update `TestCreate_DevsyResultAndEnvironmentContract` in
       `cmd/cmd_test.go`: supply a ready `VMStatus`; remove the assertion
       that `GetNetworkInterfaces` was called.

## 3. Verify
- [x] 3.1 Run `go build ./...` (passes).
- [x] 3.2 Run `go test ./...` (passes; all WaitForVMReady + create
       contract tests pass).
- [x] 3.3 Run `gofmt -l .` (prints nothing).

## 4. OpenSpec deliverables
- [x] 4.1 Create this change under `openspec/changes/vm-startup-reachability-fix/`
       (proposal.md, design.md, tasks.md, specs/).
