# Design

## Context

The workspace creation flow in `cmd/create.go` has a Step 6 that waits for a freshly cloned, freshly started workspace VM to report a non-loopback IPv4 address. The wait is implemented in `pkg/proxmox/poll.go` (`WaitForIP`) as a loop that polls `GetNetworkInterfaces` on the QEMU guest agent through the `Client` API (already an existing method targeting `/nodes/{node}/qemu/{vmid}/agent/network-get-interfaces`), and a plain `time.After(timeout)` hard deadline of 2 minutes.

The root cause of bug 2 is that the hard cutoff is not "bounded and actionable": a cold VM whose guest agent takes 3–5 minutes to boot and obtain DHCP (the agent can be live and responding to other QMP commands while the network interface is still absent or loopback-only) trips the 2-minute `time.After` and the VM is destroyed and recreated. That is a false failure of a healthy guest, and it makes `devsy workspace up` hang/unreasonably unreliable on slow hosts.

The design separates two concerns so a slow-but-healthy guest is never treated as a hard timeout:

1. **Agent-readiness probe (new)**: `WaitForAgent` polls the guest agent until it is demonstrably alive (the agent responds to a query without a connection/CGU error), with its own bounded timeout. This establishes "the guest is up" before we start waiting on the IP.
2. **Bounded IP wait (revised)**: `WaitForIP` keeps polling for the IP but uses a bounded, actionable timeout (a small multiple of the agent-boot budget, e.g. 10 minutes) and returns an error naming the waited duration and the condition, so a hang is debuggable instead of a bare "timed out".

## Goals / Non-Goals

**Goals:**
- A workspace VM whose guest agent is alive but slow is given enough time to report an IP. The 2-minute hard timeout is removed.
- The IP wait timeout is bounded (no unbounded hang) and the error message states exactly how long was waited and why it failed, making hangs actionable.
- Agent readiness is checked before the IP poll, so a guest that is starting up but not yet able to produce an IP is not mistaken for a dead agent/VM.
- Existing behavior is preserved: return the first non-loopback IPv4 address, ignore loopback and IPv6, retry on transient agent errors; `WaitForIP` still uses the existing `Client.GetNetworkInterfaces` surface.
- Existing tests continue to pass after the change (bounded wait, unchanged mock behavior).

**Non-Goals:**
- Changing the `Client` interface or adding a new mock field for the probe in a way that breaks the existing mock contract. The probe is additive; the existing `GetNetworkInterfaces` path is unchanged.
- Tuning the timeout value to a specific host; the value is a documented, bounded default and is easy to revisit.
- Handling VMs that never get an IP due to missing DHCP/failed guest install (out of scope; that is a different failure mode with different remediation).
- Addressing any of the unrelated OpenSpec changes on the parent branch (`delete-running-vm`, `migrate-proxmox-provider`), which are not part of this change.

## Decisions

1. **Add a dedicated agent-readiness probe before IP polling.** `WaitForAgent` polls `GetNetworkInterfaces` until the agent responds (succeeds) or a small bounded timeout elapses. A successful response means the guest agent is alive; a fresh-failed response is treated as "not ready yet" and retried, mirroring how the IP poll treats transient errors. This gives the caller a clean "agent alive / not alive" signal instead of burying it in the IP-poll error.
2. **Keep the probe on the existing `Client` API surface.** The probe calls the already-existing `GetNetworkInterfaces`, so no interface change or new mock field is needed. The readiness probe is a pure composition of existing calls.
3. **Use a bounded, multiple-of-boot-delay timeout for the IP wait.** Pick a value that is comfortably above a typical cold-boot+DHCP window but small enough to never hang the provider: 10 minutes. The error message names the waited duration and the condition ("agent is alive but no non-loopback IPv4 address reported within N").
4. **Fail with an actionable error when the agent is alive but the IP never appears.** Instead of the opaque "timed out waiting for VM IP after 2m0s", return an error that tells the operator how long was waited, that the agent responded, and what to check (agent installed, DHCP on, guest not frozen).
5. **Preserve the existing IP-poll semantics.** The ticker, loopback/IPv6 skipping, and retry-on-error behavior stays the same; only the deadline and the error text change. `WaitForIP` remains a self-contained function with the same signature so callers (and the existing tests) are unaffected except for the timeout value.
6. **Make the timeout a named, discoverable constant/duration.** Document the chosen duration in the code and in the spec so the value is auditable and can be adjusted per environment without a behavior diff.

## Risks / Trade-offs

- **Longer wait on slow hosts.** If a host routinely takes >10 minutes for a fresh VM to report an IP, the default will still stall. Mitigated by: (a) agent-readiness probe failing fast if the agent never becomes alive, and (b) the actionable error pointing at the specific symptom. The value is configurable by callers if needed.
- **Probe cost.** `WaitForAgent` adds at most a few extra API round-trips on a cold VM. Negligible compared to the value of not destroying a healthy guest.
- **No interface change.** Because we reuse `GetNetworkInterfaces`, we cannot get the *reason* the agent is unresponsive without an additional call. Acceptable: the new probe returns a bounded "agent not ready within N" error, and the caller sees it as a readiness failure distinct from a stale-IP timeout.
- **Test coverage gap.** The existing `poll_test.go` mocks `GetNetworkInterfaces`; this change means the IP-poll path now depends on the probe passing first. We keep the existing mock and add tests for the probe + bounded timeout, without changing the mock surface.
