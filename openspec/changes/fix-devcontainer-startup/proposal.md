# Proposal: Fix devcontainer startup — Devsy daemon state root (literal tilde)

## Why

During workspace creation and devcontainer startup on a Proxmox-backed Devsy
machine, one blocking error occurs in sequence in the `devsy-proxmox-provider`
workspace:

1. `install Devsy daemon: resolve daemon state location: daemon state root must be
   absolute: "~/.devsy/data"`

   The Devsy daemon installer resolves its data root from a literal `~` that is
   never expanded to an absolute path, so the root is rejected and the daemon
   cannot be installed on the VM. This blocks `devsy up`/workspace creation.

The second error logged in the same workflow —

```text
build image: get image build info: ... docker-credential-devsy: cannot run
executable found relative to current directory
```

— is a **separate, unfixed** issue: the Docker CLI's `docker-credential-devsy`
credential helper is not discoverable on `$PATH`. I checked it but it is out of
scope for this change (it needs a per-environment docker setup step). It is
tracked as a follow-up, not claimed here.

Together with the credential-helper issue, they prevent any Devsy workspace from
starting. This change fixes the **daemon state root** error, which is the one
blocking `devsy up` right now.

## What Changes

- **NEW** Remove the literal tilde from the provider manifest
  (`provider.yaml.tmpl`, agent `dataPath`) so Devsy's daemon installer never
  receives `dataPath: ~` and uses its own absolute default state root instead.
- **NEW** Relax the provider-manifest contract test (`cmd/provider_manifest_test.go`)
  so it no longer pins the obsolete `dataPath: ~/.devsy/data` as the correct
  contract; it now rejects any literal `~` in `dataPath` and requires the
  lifecycle entries (`init`/`create`/`...`) to be present.
- **NEW** Document the environment prerequisite for overriding the daemon data
  root per user: `DEVSY_HOME` / `--home`, so a custom guest-user root can be
  selected without hardcoding a path the provider cannot know.

## Capabilities

### New Capabilities

- `devsy-proxmox-provider`: provision and manage Proxmox workspace VMs while
  guaranteeing the Devsy daemon installs with a Devsy-supported absolute state
  root, so `devsy up`/workspace creation starts cleanly.

### Modified Capabilities

- None.

## Impact

- `provider.yaml.tmpl` (agent `dataPath` removed).
- `cmd/provider_manifest_test.go` (contract test relaxed).
- `openspec/changes/fix-devcontainer-startup/*`.
- Docs: setup/run instructions, daemon data-root override (`DEVSY_HOME`).

> Note: the `docker-credential-devsy` helper discoverability issue is tracked as
> a follow-up, not included here. Avoiding scope creep on this change.

