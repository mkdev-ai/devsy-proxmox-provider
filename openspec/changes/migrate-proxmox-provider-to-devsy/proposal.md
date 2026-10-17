# Proposal

## Why

The provider currently provisions Proxmox VMs but still implements DevPod's provider and SSH-agent contract. When run by Devsy, the DevPod helper command is passed back as the SSH injection handshake's start line, so workspace startup fails before the agent can connect. Migrating to Devsy's maintained provider contract and validating the complete install-to-SSH flow will make the existing Proxmox integration usable again.

## What Changes

- **BREAKING** Migrate the provider manifest, commands, configuration syntax, and user-facing product references to the current Devsy machine-provider and agent-injection contract; DevPod CLI compatibility is not a goal.
- Create new Proxmox accounts, tokens, templates, tags, and workspace VMs with Devsy-branded names. Do not discover, adopt, resume, rename, or delete resources created with legacy DevPod names; they remain untouched and outside this provider's supported scope.
- Make provider initialization validate its configured options, Proxmox API access, and template through Devsy's supported initialization flow after options are supplied.
- Update all user-facing installation, configuration, setup/uninstall, and troubleshooting documentation and messages to use Devsy commands and terminology. Clearly state that pre-migration DevPod-named resources are not adopted by the Devsy provider. Preserve the separate root-run Proxmox setup script that creates the service account, API token/ACLs, and cloud-init template; Devsy provider installation must not run privileged setup implicitly. The configured network bridge and storage pool remain pre-existing prerequisites.
- Align the Go module path and self-referencing provider/docs URLs with the existing canonical repository in a separate change; leave the upstream fork repository untouched and do not rename the current GitHub repository.
- Add automated contract coverage and an end-to-end acceptance check that creates a workspace, establishes agent/SSH connectivity, resumes it, and tears it down using Devsy.

## Capabilities

### New Capabilities

- `proxmox-devsy-provider`: Configure the Proxmox machine provider in Devsy, provision and manage workspace VMs, and establish agent/SSH access reliably.

### Modified Capabilities

- None.

## Impact

Affected areas include `provider.yaml.tmpl`, provider commands and options in `cmd/` and `pkg/options/`, `scripts/setup.sh`, `scripts/uninstall.sh`, unit and integration tests, release/install instructions, and all user-facing documentation. The workflow depends on the current Devsy provider/agent contract, a root-run Proxmox setup step, an existing network bridge and storage pool, and access to a Proxmox environment for end-to-end validation. The canonical repository is `mkdev-ai/devsy-proxmox-provider`; Go module/import and self-referencing URL alignment are handled separately, and the upstream fork is left untouched.