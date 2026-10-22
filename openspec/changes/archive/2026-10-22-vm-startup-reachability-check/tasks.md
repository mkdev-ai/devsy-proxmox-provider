# Tasks

## 1. Add a qemu-status + guest-agent reachability gate
- [x] 1.1 Add `WaitForVMReady` to `pkg/proxmox/poll.go`: poll `GetVMStatus` and
       `GetNetworkInterfaces` until both respond or timeout (~90s). Any single
       read error restarts the tick; both must clear in one tick to return.

## 2. Wire the gate into the create flow
- [x] 2.1 In `cmd/create.go`, after `StartVM` + `WaitForTask` and before the
       `Waiting for VM to get an IP address...` / `WaitForIP` block, insert a
       call to `WaitForVMReady(ctx, client, opts.Node, vmid, 90*time.Second)`.
- [x] 2.2 Propagate the result as an actionable `verify VM readiness` error.

## 3. Pin behavior in tests
- [x] 3.1 Add `TestWaitForVMReady` cases in `pkg/proxmox/poll_test.go`:
       - ready status + ready interfaces → nil
       - ready status but agent error → error after timeout
       - status error but ready interfaces → error after timeout
- [x] 3.2 Update `cmd/cmd_test.go` so the create contract test supplies a ready
       `VMStatus` alongside the ready `Ifaces` (the new gate requires both).
- [x] 3.3 Ensure the mock still records `GetVMStatus` + `GetNetworkInterfaces`
       calls during create.

## 4. Verify
- [x] 4.1 Run `go test ./...` (unit suite).
- [x] 4.2 Run the env-gated acceptance suite if env credentials are available
       (`go test -tags=acceptance -run TestDevsyAcceptance`); otherwise rely on
       the unit suite plus a real-Host smoke noting the fixed gate.

## 5. OpenSpec deliverables
- [x] 5.1 Create this change under `openspec/changes/vm-startup-reachability-check/`
       (proposal, design, spec, tasks).
- [x] 5.2 Validate the change (`openspec validate`), commit, and open a PR.
- [x] 5.3 On merge, archive the change into `openspec/changes/archive/`.
