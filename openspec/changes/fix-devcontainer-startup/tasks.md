# Tasks

## 1. Reproduce and isolate the blocking error
- [x] 1.1 Reproduce `install Devsy daemon: resolve daemon state location: daemon
       state root must be absolute: "~/.devsy/data"` at `devsy up` time.
- [x] 1.2 Confirm the source: `provider.yaml.tmpl` agent `dataPath` is pinned to a
       literal `~`.
- [x] 1.3 Confirm Devsy's daemon installer rejects a non-expanded tilde and that
       the provider should instead rely on Devsy's absolute default state root,
       overridden via `DEVSY_HOME`/`--home`.

## 2. Fix the manifest so the daemon installer never sees a tilde
- [x] 2.1 Remove `dataPath` (or pin an absolute path) so no `~` reaches Devsy's
       daemon installer in `provider.yaml.tmpl`.
- [x] 2.2 Verify `devsy provider add` and provider readiness no longer emit
       `state root must be absolute`.

## 3. Relax the provider-manifest contract test
- [x] 3.1 Update `cmd/provider_manifest_test.go` so it no longer pins the obsolete
       `dataPath: ~/.devsy/data` as the correct contract.
- [x] 3.2 Add checks that reject any literal `~` in the agent `dataPath` and that
       required lifecycle entries (`init`/`create`/...) are present.

## 4. Documentation
- [ ] 4.1 Update README/setup docs with the daemon data-root override
       (`DEVSY_HOME` / `--home`), explaining why the manifest must not hardcode a
       per-user path.

## 5. Verification
- [ ] 5.1 Run the provider-manifest contract test and the full unit suite
       (`go test ./...`).
- [ ] 5.2 Hold a live `devsy up` end-to-end check (or CI-gated equivalent) to
       confirm the workspace starts without the `state root must be absolute`
       error.

