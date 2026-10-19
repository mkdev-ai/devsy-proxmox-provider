# Tasks: resolve machine-delete exit status 1

> Implementation state: root cause identified and confirmed from live logs
> (HTTP 501 "Method POST .../destroy not implemented"). The Go implementation +
> tests are pending.

## 1. Fix the destroy endpoint in `pkg/proxmox/client.go`
- [ ] 1.1 Change `DeleteVM` to use `DELETE /nodes/{node}/qemu/{vmid}` (remove the
      trailing `/destroy` path segment).
- [ ] 1.2 Add `purge=1` and `destroy-unreferenced-disks=1` query params.
- [ ] 1.3 Switch the HTTP method to `http.MethodDelete`.
- [ ] 1.4 Add a transient-error retry (1s/2s/4s, max 3 attempts) for HTTP 5xx /
      API errors on the destroy request.
- [ ] 1.5 Surface the underlying Proxmox API error detail in the wrapped error.

## 2. Fix the shutdown + destroy lifecycle in `cmd/lifecycle.go`
- [ ] 2.1 Remove the `time.Sleep(2 * time.Second)` hack and the ignored `StopVM`
      call.
- [ ] 2.2 Wait for the shutdown task with `WaitForTask` (capped at ~30s) and only
      proceed to destroy if the task reports `stopped` + `OK`.
- [ ] 2.3 Keep the emergency-cleanup step (stop, destroy without purge) after a
      transient destroy failure.

## 3. Contract tests
- [ ] 3.1 Add `TestDeleteVM_RunningVM`: mock shutdown returns OK, then destroy
      with purge succeeds; assert the request uses `DELETE /nodes/{node}/qemu/{vmid}`
      with `purge=1` + `destroy-unreferenced-disks=1`, and that `/destroy` is NOT
      called.
- [ ] 3.2 Add `TestDeleteVM_Transient500`: destroy fails 2x with 500, succeeds
      on the 3rd; assert retry happened and the final error is wrapped.
- [ ] 3.3 Add `TestRunDelete` end-to-end against a fake HTTP client that
      exercises shutdown -> wait -> purge destroy -> wait.

## 4. End-to-end acceptance (env-gated)
- [ ] 4.1 Extend `cmd/cmd_test.go` so the acceptance suite: creates a devsy-named
      VM, starts it, deletes it, and asserts the machine is gone with no leftover
      VM and no orphaned finalizer.
- [ ] 4.2 Run `go test -tags=acceptance -v -run TestDevsyAcceptance` with
      `DEVSY_PROXMOX_*` set and verify exit 0 on a live Proxmox node.

## 5. Documentation
- [ ] 5.1 Update `README.md` delete/cli docs: mention that delete works for
      running machines and that emergency cleanup is available on failure.
