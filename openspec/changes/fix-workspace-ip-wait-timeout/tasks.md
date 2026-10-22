# Tasks

## 1. Add bounded, actionable IP wait + agent-readiness probe

- [x] 1.1 Add `WaitForAgent` to `pkg/proxmox/poll.go`: poll `client.GetNetworkInterfaces` until the guest agent responds (alive), with a bounded timeout (e.g. 30s), ignoring transient agent errors.
- [x] 1.2 Replace the bare 2-minute `time.After` deadline in `WaitForIP` with a bounded, named timeout (e.g. 10m) and an actionable error message that names the waited duration and the condition when the agent is alive but no IP appears.
- [x] 1.3 Preserve existing `WaitForIP` semantics: return the first non-loopback IPv4, ignore loopback and IPv6, retry on transient agent errors.

## 2. Wire the probe into Step 6 of workspace creation

- [x] 2.1 In `cmd/create.go`, call `WaitForAgent` before `WaitForIP` in Step 6.
- [x] 2.2 On agent-not-ready, fail fast with an actionable error instead of stalling on the IP poll; on agent-alive-but-no-IP, fail with the bounded/actionable IP-wait error.
- [x] 2.3 Ensure the IP-wait error message is actionable (names duration + symptom) for operators.

## 3. Verify

- [x] 3.1 `gofmt -l pkg/proxmox/ cmd/` reports no_files.
- [x] 3.2 `go build ./...` succeeds.
- [x] 3.3 `go test ./pkg/proxmox/... ./cmd/...` passes (existing + new probe/timeout tests).
- [x] 3.4 Archive this change with `openspec archive` (no spec update in this repo; `skip-specs` if needed).
