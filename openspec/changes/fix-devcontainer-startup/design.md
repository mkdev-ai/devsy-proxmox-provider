# Design

## Context

See `proposal.md` for motivation and the two blocking errors. The Devsy daemon is
installed on the VM during `devsy up`; its state root must be an absolute path.
The devcontainer build then runs `docker build`, which invokes `docker-credential-devsy`
to authenticate image pulls from `mcr.microsoft.com` (and registries like it). Both
steps are environment/credential-setup concerns, not Proxmox provisioning logic.

## Goals / Non-Goals

**Goals:**
- The Devsy daemon installs with an absolute state root resolved from `$HOME`
  (or an explicit absolute `DEVSSY_DATA_ROOT`), with a clear error otherwise.
- The `docker-credential-devsy` helper is installed where the Docker CLI can find
  it and returns credentials correctly for a devcontainer image pull.
- A newly created Devsy workspace prepares and starts its devcontainer end to end.

**Non-Goals:**
- Changing Proxmox provisioning/lifecycle logic (covered by the migrate/rename
  changes).
- Implementing the `docker-credential-devsy` credential store itself; we assume
  the helper exists and only guarantee it is discoverable and invocable.
- Silently falling back to legacy DevPod-branded resources.

## Decisions

1. **Normalize the daemon state root at resolution time.** Resolve `~` via
   `os.UserHomeDir()` (or an explicit absolute `DEVSSY_DATA_ROOT`) and require the
   result to be absolute. Reject otherwise with an error naming the bad root and
   the expected absolute form. This fixes the `"~/.devsy/data"` case without
   changing the documented default.

2. **Install the credential helper on PATH.** Place `docker-credential-devsy`
   into a directory on `$PATH` (e.g., `~/.docker/cli-plugins/` or a documented
   provider bin dir), matching the Docker CLI's `docker-credential-*` discovery
   convention. Add a contract test that `command -v docker-credential-devsy`
   succeeds and that invoking `docker-credential-devsy get` returns a parseable
   map. Do not depend on a real Docker daemon for this test.

3. **Verify end to end, not only in unit tests.** The acceptance flow creates a
   workspace, installs the Devsy daemon, and runs a devcontainer build that
   exercises the credential helper. Gate a live run on `DEVSSY_PROXMOX_*` so CI
   stays green (emulated/proxied Proxmox or a documented operator-run check).

## Risks / Trade-offs

- If `$HOME` is unset for the daemon installer, expansion still fails; guard with
  an actionable error and document setting `DEVSSY_DATA_ROOT` to an absolute path.
- Installing the helper into a PATH dir is user-specific; prefer a documented,
  reproducible step in the provider setup. Keep the core contract test
  docker-daemon-free so CI stays green.

## Open Questions

- Which `docker-credential-devsy` binary/version should be shipped/used? Confirm
  during implementation against the Devsy CLI release used in CI.
- Is the helper already built in the repo, or does the provider shell out to the
  Devsy CLI to fetch/install it? Decide the install step (download vs
  build-from-source).
