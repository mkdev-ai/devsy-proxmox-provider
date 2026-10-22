# Tasks

## 1. Add template disk size option

- [x] 1.1 Add `TEMPLATE_DISK_SIZE` (MiB, default `3584`) to `scripts/setup.sh` defaults alongside `DISK_FORMAT`.
- [x] 1.2 Add `--template-disk-size` argument parsing to `scripts/setup.sh` (reject non-numeric input with an actionable error).

## 2. Size the template base disk

- [x] 2.1 Pass `--size "${TEMPLATE_DISK_SIZE}"` to `qm importdisk` in `scripts/setup.sh` so the template base lands at the requested capacity.
- [x] 2.2 Document `--template-disk-size` in the script's help text and config table.

## 3. Resize the clone before install

- [x] 3.1 Rewrite the Step 4 resize in `cmd/create.go` to compute the target size from the actual template base and apply an absolute `VM_DISK_SIZE` resize **before** the VM starts.
- [x] 3.2 Remove the hard-coded `templateDiskGB := 17` assumption.

## 4. Verify and document

- [x] 4.1 Run `bash -n scripts/setup.sh scripts/uninstall.sh` (syntax check).
- [x] 4.2 Run `gofmt -l cmd/ pkg/ openspec/` and `go build ./...`.
- [x] 4.3 Run `go test ./cmd/... ./pkg/...` to confirm no regressions.
- [x] 4.4 Update the OpenSpec spec (`specs/devsy-proxmox-provider/spec.md`) with the template-disk behavior requirement.

## Implementation + integration (PR to main)

- [ ] 5.1 Commit task 1–4 changes as `feat(provider): size template base disk and resize clone before install` on `feature/template-disk-sizing`.
- [ ] 5.2 Push `feature/template-disk-sizing` and open a PR to `main`.
- [ ] 5.3 Merge the PR on `main`, then tag `v<next>` (release automation handles publishing).
