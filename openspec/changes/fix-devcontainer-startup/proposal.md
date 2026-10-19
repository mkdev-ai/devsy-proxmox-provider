# Proposal: Fix devcontainer startup (Devsy daemon root + docker-credential-devsy)

## Why

During workspace creation and devcontainer startup on a Proxmox-backed Devsy
machine, two blocking errors occur in sequence in the `devsy-proxmox-provider`
workspace:

1. `install Devsy daemon: resolve daemon state location: daemon state root must be
   absolute: "~/.devsy/data"`

   The Devsy daemon installer resolves its data root from a literal `~` that is
   never expanded to an absolute path, so the root is rejected and the daemon
   cannot be installed on the VM. This blocks `devsy up`/workspace creation.

2. `build image: get image build info: get image config remotely: retrieve image
   mcr.microsoft.com/devcontainers/go: error getting credentials - err: exec:
   "docker-credential-devsy": cannot run executable found relative to current
   directory`

   The Docker CLI invokes the `docker-credential-devsy` credential helper to
   authenticate registry access for devcontainer images. The helper binary is not
   discoverable on `$PATH`, so auth fails and the devcontainer cannot be
   prepared or started.

Together these two errors prevent any Devsy workspace from starting. The fix is
environment/credential-setup focused: make the daemon state root absolute and
make `docker-credential-devsy` discoverable and invocable by Docker.

## What Changes

- **NEW** Make the devcontainer start reliably by removing the `dataPath`
  tilde from the provider manifest (`provider.yaml.tmpl`) so Devsy's daemon
  installer no longer rejects it as a non-absolute root, and relax the
  provider-manifest contract test to no longer pin the obsolete tilde.
- **NEW** Document the environment prerequisite (`DEVSY_HOME`) for overriding
  the daemon data root per user.

## Capabilities

### New Capabilities

- `devsy-proxmox-provider`: provision and manage Proxmox workspace VMs while
  guaranteeing the Devsy daemon installs with a Devsy-supported absolute state
  root, so `devsy up`/workspace creation starts cleanly.

### Modified Capabilities

- None.

## Impact

- `provider.yaml.tmpl` (agent.dataPath) and `cmd/provider_manifest_test.go`
  (contract test).
- `openspec/changes/fix-devcontainer-startup/*`.
- Docs: setup/run instructions, environment prerequisites.
