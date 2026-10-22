# Spec Delta

## Purpose

Defines the observable contract for using the Proxmox machine provider with Devsy, from provider initialization through reliable SSH agent access and VM lifecycle management.

## ADDED Requirements

### Requirement: Devsy machine-provider integration

The provider SHALL use Devsy's supported machine-provider and agent configuration contract and SHALL NOT require DevPod-only variables or helper commands.

#### Scenario: Devsy installs and selects the provider
- **WHEN** a user installs the provider with Devsy and supplies all required Proxmox options
- **THEN** Devsy recognizes it as a machine provider and can invoke its lifecycle and agent operations using the current Devsy binary

### Requirement: Devsy user-facing documentation and syntax

Provider metadata, setup output, and user-facing documentation SHALL use Devsy product terminology and CLI syntax; any mention of DevPod-era resources SHALL state that they are unsupported and are not adopted by the Devsy provider.

#### Scenario: Install and configure from provider documentation
- **WHEN** a user follows the provider's documented installation and workspace workflow
- **THEN** all CLI commands use Devsy syntax and all setup instructions describe the Devsy provider flow

#### Scenario: Explain unsupported legacy Proxmox resources
- **WHEN** documentation mentions an existing resource created by an older provider version
- **THEN** the DevPod-era identifier is explicitly labeled unsupported and documentation does not suggest importing or falling back to it

### Requirement: Devsy-branded names for new Proxmox resources

The setup and provider workflows SHALL use Devsy-branded defaults for newly created Proxmox accounts, API tokens, cloud-init templates, guest users, tags, and workspace VM names.

#### Scenario: Prepare a new Proxmox installation
- **WHEN** an administrator runs setup with its default names on a new Proxmox host
- **THEN** newly created provider and template resources use Devsy-branded names

#### Scenario: Create a new workspace VM
- **WHEN** Devsy creates a workspace using the migrated provider
- **THEN** the new VM uses the Devsy workspace naming prefix and tag

### Requirement: SSH agent command relay

The provider SHALL execute each remote command supplied by Devsy on the target VM and relay its standard input, output, and error streams without emitting an unevaluated command template into the protocol stream.

#### Scenario: Devsy injects its agent over SSH
- **WHEN** Devsy requests an agent command on a running VM
- **THEN** the command runs on the VM and the injection handshake receives the expected protocol response rather than the provider's command template

### Requirement: Provider initialization validates readiness

The provider SHALL report initialization success only when required options are valid and the configured Proxmox API, node, and VM template are usable.

#### Scenario: Valid provider configuration
- **WHEN** Devsy initializes the provider with valid credentials and an available template
- **THEN** initialization succeeds and the provider is ready to create a workspace VM

#### Scenario: Invalid or unreachable Proxmox configuration
- **WHEN** initialization cannot validate a required option, API connection, node, or template
- **THEN** it exits unsuccessfully with an actionable error and does not report the provider as ready

### Requirement: Proxmox machine lifecycle

The provider SHALL create, start, stop, report the status of, and delete the Proxmox VM associated with a Devsy machine while returning the connection information Devsy needs to reach a newly created VM.

#### Scenario: Create and connect to a workspace machine
- **WHEN** Devsy creates a machine from a configured Proxmox cloud-init template
- **THEN** the provider starts the VM, waits for its reachable IP address, and returns the machine connection details

#### Scenario: Resume and delete a machine
- **WHEN** Devsy starts, stops, checks, or deletes an existing machine
- **THEN** the provider applies that operation to the associated VM and reports the resulting status using Devsy's expected values

### Requirement: No fallback to legacy Proxmox resources

The provider SHALL discover and manage machines only by Devsy-branded Proxmox identifiers and SHALL NOT adopt, resume, modify, rename, or delete resources created with legacy DevPod-era names.

#### Scenario: Lifecycle lookup finds only a legacy VM
- **WHEN** a Devsy lifecycle operation has no matching Devsy-named VM but a VM with the same machine identity has a legacy DevPod-era name
- **THEN** the provider reports the Devsy machine as not found and does not fall back to or alter the legacy VM

#### Scenario: Create a machine whose identity has a legacy VM
- **WHEN** Devsy creates a new machine with an identity that exists only on a legacy DevPod-named VM
- **THEN** the provider creates a new VM using the Devsy naming scheme and leaves the legacy VM untouched

### Requirement: Automatic provider readiness after configuration

The documented Devsy installation and configuration flow SHALL initialize and validate the provider once the required Proxmox options have been supplied, without requiring an undocumented manual step.

#### Scenario: Install and configure a ready provider
- **WHEN** a user installs the provider and supplies valid required options through the documented Devsy flow
- **THEN** the provider becomes ready and the user can start a workspace without manually invoking provider internals

#### Scenario: Configure before Proxmox prerequisites are ready
- **WHEN** the required Proxmox host preparation or provider options are missing
- **THEN** initialization reports the unmet prerequisite clearly and does not claim that workspace creation is ready

### Requirement: Explicit Proxmox host bootstrap

The setup workflow SHALL retain a separate script that an administrator runs as root on the Proxmox host to create or verify the provider service account, API token and ACLs, and cloud-init template; Devsy provider installation SHALL NOT invoke this privileged script implicitly.

#### Scenario: Prepare a Proxmox host
- **WHEN** an administrator runs the documented setup script as root with the desired node and template settings
- **THEN** the script creates or verifies the provider account, access grants, and cloud-init template and reports the API token needed to configure the Devsy provider

#### Scenario: Repeat Proxmox host setup
- **WHEN** an administrator reruns the setup script for resources it has already prepared
- **THEN** the script recognizes the existing resources and completes without creating duplicate accounts, tokens, or templates

### Requirement: Template disk sizing
The setup workflow SHALL create the cloud-init template with a configurable
root disk size so the template base is not artificially small. The provider
SHALL resize the cloned workspace VM to `VM_DISK_SIZE` immediately after
cloning and **before** the VM starts, so the workspace install (git clone /
docker pull) always has the requested capacity even if the template base
was small.

#### Scenario: Template created at the desired size
- **WHEN** an administrator runs `scripts/setup.sh --template-disk-size
  <size-mib>`
- **THEN** the template's root disk (virtio0) is created at that exact
  capacity, not the Ubuntu cloud image default of ~3.5 GB.

#### Scenario: Workspace VM sized before install
- **WHEN** a workspace is created with `VM_DISK_SIZE` greater than the
  template base
- **THEN** the provider resizes the clone to exactly `VM_DISK_SIZE` before
  starting the VM, and the git clone and Docker pull that follow see
  adequate disk space.
