# Tasks

## 1. Reproduce and isolate both blocking errors
- [x] 1.1 Reproduce `install Devsy daemon: daemon state root must be absolute` in
       a fresh workspace and capture the environment (`$HOME` value; where the
       state root string is constructed).
- [x] 1.2 Reproduce `docker-credential-devsy: cannot run executable found
       relative to current directory` and confirm the Docker CLI lookup order.
- [x] 1.3 Add focused regression tests (binary-level, no Docker needed) that pin
       both root causes.

## 2. Normalize the Devsy daemon state root
- [x] 2.1 Add a `resolveStateRoot()` helper that expands `~` to `os.UserHomeDir()`
       and validates the result is absolute.
- [x] 2.2 Wire it into the daemon installer and fail with an actionable error when
       the root cannot be resolved to an absolute path.
- [x] 2.3 Add unit checks: a literal `~/.devsy/data` is rejected; a valid
       `$HOME/.devsy/data` is accepted as absolute; the default is absolute.

## 3. Install and verify docker-credential-devsy
- [ ] 3.1 Add a task that installs `docker-credential-devsy` into a
       PATH-accessible directory (or documents where the Devsy CLI places it).
       No install step exists yet — the helper is expected to be present on the
       VM when the devcontainer runs. Deferred until environment setup.
- [x] 3.2 Add a contract test that `command -v docker-credential-devsy` finds it
       and that invoking `docker-credential-devsy get` returns a parseable map.
- [ ] 3.3 (DEFERRED) Environment-gated E2E test creating a workspace, preparing
       a devcontainer that pulls an authenticated image, and asserting the
       helper resolved and the build succeeded. Skipped for now — the live Proxmox
       environment/credentials need to be sorted out before implementing. See
       design.md "Open Questions".

## 4. Documentation
- [ ] 4.1 Update README/setup docs with the absolute daemon state root and the
       `docker-credential-devsy` setup prerequisites. (Not done in this slice —
       left for a follow-up docs commit.)
- [ ] 4.2 Run `go test ./...` (no Docker daemon, no E2E) and record the Devsy CLI
       version and the `docker-credential-devsy` behavior in the change notes.
- [ ] 4.3 Commit the new artifacts, code, and tests as a single atomic slice:
       `fix(provider): resolve devcontainer startup credential and daemon-root
       errors`; verify the commit contains only this change and its tests.
