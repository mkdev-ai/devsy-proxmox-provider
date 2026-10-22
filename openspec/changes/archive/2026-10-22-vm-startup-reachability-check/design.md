# Design

## Context
See `proposal.md` for motivation and `specs/vm-startup-reachability-check/spec.md` for the
behavior contract. The bug is that `runCreate` goes from `StartVM`/`WaitForTask`
straight to `WaitForIP` with no intermediate reachability check, so a newly
booted VM (guest agent up, but no leased IP yet) sits in the 2-minute IP poll
for the rest of a 10+ minute window.

## Goals
- Insert a reachability check after `StartVM`/`WaitForTask` and before
  `WaitForIP` that polls the qemu status endpoint and the guest agent (~90s).
- Only report success to Devsy once the VM has responded to a status/config
  read, not merely after the start task completed.
- Keep `WaitForIP` as the IP-specific poll; this new check is purely a
  pre-flight reachability gate.
- Add tests that pin the behavior (ready path + unreachable-as-error path) so
  the readiness check cannot silently regress.

## Options Considered
- **Poll only the IP.** Would not fix the gap — the agent can be ready with no
  IP (e.g. dhcp not leased yet), so the 10+ minute window remains.
- **Poll only the hypervisor status (`GetVMStatus`).** Confirms the VM is
  running but not that the guest agent (which supplies the config/agent read)
  is accepting queries.
- **Poll only the guest agent (`GetNetworkInterfaces`).** Confirms the agent is
  up but not that the hypervisor status/read behaves; a mismatch between the
  two is exactly the kind of partial-read that this gate rejects by requiring
  *both*.
- **Poll both (chosen).** Requires the VM to pass a qemu status read AND a
  guest-agent read before the IP wait begins. Cheap, fast on a healthy VM, and
  it directly verifies the reachability described in the bug.

## Decisions
1. **Reachability = both a qemu status read and a guest-agent read succeed.**
   `WaitForVMReady` calls `GetVMStatus` then `GetNetworkInterfaces` each tick;
   either error continues the poll. Both must return no error before the
   function returns.
2. **Timeout is ~90s.** Short enough to not block a healthy VM, long enough to
   absorb a cold boot + guest-agent startup on a slower host. The IP poll is a
   separate, longer in-flight check.
3. **Timeout is an error, not a silent skip.** If the VM does not respond to
   either read within 90s, `WaitForVMReady` returns an error and `runCreate`
   aborts with an actionable message instead of parking in the IP wait.
4. **Keep `WaitForIP` unchanged.** The agent-read path still drives IP discovery;
   this check only pre-flights reachability before that poll starts.
5. **Mock-friendly.** `WaitForVMReady` works through the `Client` interface so
   the existing `MockClient` records `GetVMStatus`+`GetNetworkInterfaces`
   calls; tests can stage readiness by setting `VMStatus`/`Ifaces`/`*Err`.

## Risks
- **Over-blocking a healthy VM.** If `WaitForVMReady` is too slow (a real VM
  that takes >90s to reach a leased IP plus guest-agent readiness), create
  would fail even though the machine is healthy. Mitigated by a generous 90s
  timeout and by it being a pre-flight (nothing downstream changes).
- **Status+agent read ordering ambiguity.** A tick that succeeds on status but
  fails on agent is retried whole; this is intentional because we require both
  in the same tick. No state leaks across ticks.
- **Existing create contract test must be updated.** `TestCreate_DevsyResultAndEnvironmentContract`
  must supply a ready `VMStatus` alongside the ready `Ifaces`, otherwise the
  new gate times out and the test fails.

## Migration Plan
None. The change is additive to the create flow and does not touch existing
machines, provisioning, or the delete lifecycle.

## Verification
- `go test ./...` passes.
- `go test -tags=acceptance -run TestDevsyAcceptance` passes (the env-gated
  emulated suite; readiness is also exercised by the mock unit tests).
- The change is reviewed for the reachability-gate semantics in
  `cmd/create.go`.
