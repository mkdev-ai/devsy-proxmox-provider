# Spec

## Purpose
Defines the contract for verifying that a just-created workspace VM has
responded to a qemu status and guest-agent read before the provider
starts the IP wait, eliminating the 10+ minute readiness gap in
`devsy provider create`.

## ADDED Requirements

### Requirement: New VM reachable before IP wait
The provider SHALL verify that the just-created VM's qemu status and
guest-agent respond before beginning the IP wait, after `StartVM`/`WaitForTask`
completes, for up to ~90s.

#### Scenario: VM is reachable — proceed to IP wait
- **GIVEN** a VM whose start task completed and that responds to `GetVMStatus`
  and `GetNetworkInterfaces`
- **WHEN** `WaitForVMReady` runs
- **THEN** it returns nil promptly and the create flow continues to
  `WaitForIP`

#### Scenario: VM not reachable — return actionable error
- **GIVEN** a VM whose start task completed but which does not respond to
  `GetVMStatus` or `GetNetworkInterfaces` within ~90s
- **WHEN** `WaitForVMReady` runs
- **THEN** it returns an error and the provider aborts create with an
  actionable message instead of parking in the IP wait

#### Scenario: both status and agent must respond
- **GIVEN** the start task completed, the qemu status read succeeds, but the
  guest-agent read fails
- **WHEN** `WaitForVMReady` runs
- **THEN** it keeps polling until the guest-agent read succeeds, or returns an
  error if the agent never responds; a failing status read also keeps polling

## REMOVED Requirements
None. The only removal is implicit: the previous behavior (no reachability
check between StartVM+`WaitForTask` and `WaitForIP`) is replaced by the
reachability gate above.

## ADDED Non-Functional Requirements
- The IP-specific `WaitForIP` poll SHALL remain unchanged and SHALL still be
  the component that observes the leased address.
- The reachability gate SHALL NOT extend the total create latency for a
  healthy VM (it returns promptly when the VM responds).
