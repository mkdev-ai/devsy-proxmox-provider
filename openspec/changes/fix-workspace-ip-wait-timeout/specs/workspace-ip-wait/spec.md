# Spec Delta

## Purpose

The workspace creation path SHALL wait for a freshly started workspace VM to report a usable IPv4 address with a bounded, actionable timeout, and SHALL verify the QEMU guest agent is alive before polling for the IP, so a slow-but-healthy guest is never treated as a hard timeout.

## ADDED Requirements

### Requirement: Bounded, actionable IP wait

The provider SHALL poll for a non-loopback IPv4 address on a workspace VM with a bounded timeout, and SHALL fail with an error that names the waited duration and the condition when the agent is alive but no IP appears.

#### Scenario: IP appears in time

- **WHEN** the guest agent reports a non-loopback IPv4 address within the bounded wait
- **THEN** `WaitForIP` returns that address without error

#### Scenario: Agent alive but IP never appears

- **WHEN** the guest agent is alive but no non-loopback IPv4 address appears within the bounded wait
- **THEN** `WaitForIP` returns an error naming the waited duration and the condition (agent alive but no IP), instead of a bare "timed out"

#### Scenario: Loopback and IPv6 are ignored

- **WHEN** the only addresses are loopback (127.0.0.1) or IPv6
- **THEN** `WaitForIP` keeps waiting (does not match) until the bounded timeout

### Requirement: Agent-readiness probe before IP poll

The provider SHALL confirm the QEMU guest agent is reachable before waiting for the IP, with its own bounded timeout.

#### Scenario: Agent becomes alive

- **WHEN** the guest agent responds to a query within the readiness timeout
- **THEN** the IP poll proceeds and the provider waits for an IP

#### Scenario: Agent never becomes alive

- **WHEN** the guest agent does not respond within the readiness timeout
- **THEN** the provider fails fast with an actionable error indicating the agent is not ready, instead of waiting indefinitely for an IP that will not come
