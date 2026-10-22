# Proposal: Verify VM reachability before the IP wait in the create flow

## Why
After a workspace VM is started, the create flow currently:

1. `StartVM` POSTs `/status/start`, then `WaitForTask` (5 min) waits for the
   start task to reach `stopped`/`OK`. That only proves the hypervisor accepted
   the start request and finished bookkeeping.
2. The flow jumps straight to `WaitForIP` (2 min), which polls
   `GET .../agent/network-get-interfaces` until a non-loopback IPv4 appears.

Between those two steps the provider does **no reachability check at all**. On
a real Proxmox host the VM boots, qemu-guest-agent starts, and the hypervisor
reports success, but the provider never verifies the VM responded to a status
or config read. What happens next is also what the bug describes: the provider
spins through the full 2-minute IP wait, and on a cold boot / first-IP-dhcp the
VM can surface its first usable address well after the IP poll begins, so
`devsy up` / workspace creation reports success to Devsy only after a very
long delay (10+ minutes).

The fix inserts a short reachability/health check between `WaitForTask` and
`WaitForIP`: poll the qemu status endpoint and the guest agent until both
respond, for roughly 90s. A VM that is already reachable returns immediately,
so the IP wait starts from a verified-running machine instead of from a fresh
cold boot. If the VM does not respond at all within ~90s the provider aborts
with an actionable error instead of silently parking in the IP wait for the
remainder of the 10+ minute window.

Note this does not remove or replace `WaitForIP` — it precedes it. The IP wait
is still required because the guest agent reports the IP, and dhcp may not have
leased an address by the time the agent is ready.

## What Changes
- `pkg/proxmox/poll.go`: new `WaitForVMReady` that polls `GetVMStatus` and
  `GetNetworkInterfaces` (qemu status + guest agent) until both respond or a
  timeout (default ~90s).
- `cmd/create.go`: call `WaitForVMReady` right after `StartVM`/`WaitForTask`,
  before the `Waiting for VM to get an IP address...` / `WaitForIP` block.

## Capabilities
- `devsy-proxmox-provider`: provision and manage Proxmox workspace VMs while
  verifying the newly started VM is reachable (qemu status + guest agent)
  before waiting for its IP, removing the 10+ minute readiness gap.

## Impact
- `pkg/proxmox/poll.go` (new function).
- `cmd/create.go` (one extra call in the create flow).
- `pkg/proxmox/poll_test.go` + `cmd/cmd_test.go` (new tests; existing create
  contract test must supply a ready VMStatus so the new check passes).

> Note: this change is purely additive — no provider-visible behavior changes
> beyond eliminating the long delayed-IP case. It does not alter `WaitForIP`.
