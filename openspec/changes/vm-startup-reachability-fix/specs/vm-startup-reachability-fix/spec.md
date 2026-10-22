# vm-startup-reachability-fix Specification

## Purpose
Defines the contract for verifying that a just-created workspace VM is
available as soon as the hypervisor reports `qemu status == running`,
without waiting on the slow guest agent. The guest agent remains a bounded
probe inside the IP-wait step, so IP discovery still works reliably.

## ADDED Requirements

### Requirement: VM reported available when qemu status is running
The provider SHALL treat a just-created VM as available as soon as the
hypervisor reports qemu status == "running", which is the authoritative
signal that VM creation/start has completed.

#### Scenario: qemu status is running — report available immediately
- **GIVEN** a VM whose start task completed and whose qemu status is
  "running" (the guest agent may still be slow or unavailable)
- **WHEN** `WaitForVMReady` runs
- **THEN** it returns nil promptly and the create flow continues to the
  IP-wait step (`WaitForIP`)

### Requirement: guest agent probed as a bounded probe in the IP-wait step
The guest agent SHALL be probed separately by `WaitForIP` (Step 7), which
first confirms the agent is alive (`WaitForAgent`, bounded ~30s) and then
polls for a non-loopback IPv4 address with a bounded timeout (~10m).

#### Scenario: agent slow but qemu running — proceed to IP wait
- **GIVEN** a VM whose qemu status is running but whose guest agent is
  slow to become alive
- **WHEN** the IP-wait step runs
- **THEN** the bounded agent probe retries rather than hard-failing, and
  the IP poll waits for a leased IPv4 within a bounded timeout

#### Scenario: no IPv4 leased within bounded timeout — actionable error
- **GIVEN** a VM that is running and whose agent is alive but has no
  non-loopback IPv4 address
- **WHEN** the IP-wait step runs
- **THEN** it SHALL return an actionable error naming the waited duration
  and the alive-agent condition, never a silent hang

## REMOVED Requirements
The previous requirement that `WaitForVMReady` wait for both qemu status
AND guest agent is removed. The guest-agent read is no longer part of the
readiness gate.
