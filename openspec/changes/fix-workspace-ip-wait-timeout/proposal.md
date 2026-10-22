# Proposal

## Why

`cmd/create.go` (Step 6) waits for a fresh workspace VM to report an IP with a hard **2-minute** timeout via `proxmox.WaitForIP`. `pkg/proxmox/poll.go` implements the poll with `time.After(timeout)` and a 1-second ticker, so it laps a hard "timed out waiting for VM IP after 2m0s" error even when the guest agent is alive but slow. On cold VMs the QEMU guest agent (`qemu-guest-agent`) can take well over 2 minutes to report an IP (network interface appears only after the guest completes DHCP, and the guest can boot slowly).

The result is a deterministic hang: every cold/underloaded VM is deleted and recreated rather than simply waiting. The wait must be increased, and the poll must first confirm the agent is alive before waiting on the IP, so a slow-but-healthy guest is never mistaken for a hard failure.

## What Changes

- **`pkg/proxmox/poll.go`**:
  - Keep polling the QEMU guest agent for a non-loopback IPv4 address, but **first** confirm the agent is reachable via the dedicated `WaitForAgent` readiness probe.
  - Apply a bounded, actionable timeout instead of a bare 2-minute hard cutoff. The timeout is a small multiple of the agent-boot delay, and is surfaced in the error message so a caller can see exactly how long was waited and why it failed.
  - Preserve the existing semantics: return the first non-loopback IPv4 address, ignore loopback and IPv6, and retry on transient agent errors.
- **`cmd/create.go` (Step 6)**:
  - Replace the opaque `2*time.Minute` wait with a bounded, named timeout.
  - Fail with an actionable error when the agent is alive but the IP genuinely never appears, so a hang is debuggable instead of a silent timeout.
- No change to `pkg/proxmox/client.go`, the `Client` interface, or the mock. The readiness probe is an addition, not a behavior change of existing calls.

## Capabilities

- `workspace-ip-wait`: Wait for a workspace VM's IP with an agent-readiness probe and a bounded, actionable timeout.

## Impact

- `pkg/proxmox/poll.go` (new `WaitForAgent` probe + bounded `WaitForIP` timeout).
- `cmd/create.go` (bounded, named timeout + actionable error).
- User-facing docs: `README.md` (connection info + "waiting" messaging may now show a longer, explainable wait).
- OpenSpec: this change owns the new `workspace-ip-wait` capability spec.
