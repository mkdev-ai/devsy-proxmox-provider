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

- **NEW** Validate and normalize the Devsy daemon state root to an absolute path
  derived from `$HOME`; reject unexpanded/relative roots with an actionable error
  naming the bad root and the required absolute form.
- **NEW** Ensure the `docker-credential-devsy` helper is installed into a
  PATH-accessible directory and verify the Docker CLI can invoke it when resolving
  registry credentials.
- **NEW** Add contract tests for daemon-root normalization and for
  `docker-credential-devsy` discoverability/invocability.
- **NEW** Document the required environment prerequisites (absolute absolute
  daemon state root or set `DEVSSY_DATA_ROOT`, and the PATH entry for
  `docker-credential-devsy`).

## Capabilities

### New Capabilities

- `devsy-proxmox-provider`: provision and manage Proxmox workspace VMs while
  guaranteeing the Devsy daemon has an absolute state root and the Docker CLI can
  authenticate to registries via `docker-credential-devsy` so a devcontainer can be
  prepared and started.

### Modified Capabilities

- None.

## Impact

Affected files: provider daemon-install logic (wherever the daemon state root is
resolved), the Docker credential-helper setup/install step, unit and integration
tests, and user-facing setup/run documentation. Prereqs: a working Devsy CLI,
an installed `docker-credential-devsy` binary, and a valid absolute daemon data
root. This change is env/credential-setup focused; it does not amend Proxmox
provisioning or lifecycle logic (those are covered by the `migrate-*`/`rename-*`
changes).
