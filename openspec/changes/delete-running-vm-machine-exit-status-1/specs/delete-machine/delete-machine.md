# Spec Delta

## Purpose
Defines the contract for reliably deleting a workspace machine on Proxmox VE
without surfacing `exit status 1` to the user. Covers the correct HTTP API for
machine destruction, the delete/lifecycle sequence, and the tests that verify
deletion succeeds for a running machine.

## ADDED Requirements

### Requirement: Delete machine returns success
The provider SHALL delete the machine's Proxmox VM (and purge its disks) and
return success to Devsy when the operation completes.

#### Scenario: Delete a running machine
- **GIVEN** a devsy-named workspace VM is running
- **WHEN** `devsy provider delete` (or `devsy up` cleanup) runs
- **THEN** the provider shuts the VM down, waits for the shutdown task to
  complete, destroys the VM with `purge=1` + `destroy-unreferenced-disks=1`,
  and exits with status 0.

#### Scenario: Delete a stopped machine
- **GIVEN** a devsy-named workspace VM is stopped
- **WHEN** delete runs
- **THEN** the provider destroys the VM with purge and exits with status 0.

#### Scenario: Delete reports the real error
- **GIVEN** the Proxmox API returns an error on the destroy request
- **THEN** the provider returns an actionable error that includes the Proxmox
  HTTP code and message, instead of a bare `exit status 1`.

## REMOVED Requirements
- `DeleteVM` MUST NOT POST to the non-existent `/nodes/{node}/qemu/{vmid}/destroy`
  endpoint, which Proxmox returns HTTP 501 "not implemented" for.
