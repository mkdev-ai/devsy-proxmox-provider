# Tasks

## 1. Devsy Manifest and SSH Transport

- [x] 1.1 Update `provider.yaml.tmpl` for the Devsy agent and machine-provider contract, use Devsy agent paths/data paths and descriptions, and add a manifest contract test that rejects obsolete CLI syntax and checks all required lifecycle commands.
- [x] 1.2 Replace the command-template output with an SSH command runner that resolves the VM from `MACHINE_ID`, uses its current IP and generated key, and relays stdin/stdout/stderr and exit status; verify it with fake SSH and Proxmox clients, including that stdout contains no provider diagnostics.
- [x] 1.3 Update user-visible provider command descriptions, errors, and help text to Devsy terminology; add focused checks that obsolete DevPod product/CLI syntax is absent except for a clear unsupported-migration notice.
- [x] 1.4 After tasks 1.1–1.3 pass, commit only the manifest/SSH/help slice as `feat(provider): migrate SSH transport to Devsy`; verify the commit summary contains only this slice and its tests.

## 2. Lifecycle and Provider Readiness

- [x] 2.1 Verify Devsy's machine create-result and environment contract; use Devsy-branded names for accounts, tokens, templates, guest users, tags, and VMs; cover Devsy create/resume/status/stop/delete and verify a matching legacy DevPod-named VM is never discovered, adopted, or modified.
- [x] 2.2 Make provider initialization validate required options, Proxmox API access, node, and template; add tests proving valid configuration succeeds and missing/unavailable prerequisites fail with actionable errors, then run `go test ./cmd/... ./pkg/options/... ./pkg/proxmox/...`.
- [x] 2.3 After tasks 2.1–2.2 pass, commit only lifecycle, resource naming, and readiness changes as `feat(provider): add Devsy lifecycle support`; verify the commit contains its implementation and tests only.

## 3. Devsy Installation and Documentation

- [x] 3.1 Update every user-facing document and install/configure/workspace command example to Devsy terminology and syntax; document root-running `scripts/setup.sh`, Devsy-branded defaults, token handling, existing bridge/storage prerequisites, and that old DevPod-created machines are unsupported and untouched; point self-referencing install/docs links at the current canonical repo when the companion module/URL alignment is applied.
- [x] 3.2 Update setup/uninstall help and output to Devsy branding and Devsy defaults; limit resource targeting to Devsy identifiers and verify legacy DevPod resources are never adopted or deleted, bridge/storage are only used (not created/reconfigured), reruns remain idempotent, and run `bash -n scripts/setup.sh scripts/uninstall.sh` plus a mocked or dry-run check.
- [x] 3.3 After tasks 3.1–3.2 pass, commit only documentation and setup/uninstall changes as `docs(provider): document Devsy setup and legacy resources`; verify no provider implementation files are included.

## 4. Devsy Integration Verification

- [ ] 4.1 Establish and document the minimum supported Devsy CLI release, then validate provider installation and manifest initialization with that release and the current supported release.
- [ ] 4.2 Add an environment-gated Proxmox acceptance test that configures the provider, creates a Devsy-named workspace VM, runs an SSH command through the injected agent, resumes that Devsy VM, and deletes it; verify a pre-existing DevPod-named VM is not selected or altered, guarantee cleanup, and document how to run the test when CI lacks Proxmox credentials.
- [ ] 4.3 Run `go test ./...` and the Devsy acceptance test where Proxmox credentials are available; record the tested Devsy version and confirm the original `unexpected start line` failure no longer occurs.
- [x] 4.4 Coordinate with `rename-proxmox-provider-to-devsy` to set the Go module path/imports and self-referencing provider/docs URLs to `github.com/mkdev-ai/devsy-proxmox-provider`; leave upstream `iamveen/devpod-proxmox-provider` untouched.
- [ ] 4.5 After task 4.3 passes, commit any remaining integration-only changes separately as `test(provider): verify Devsy Proxmox workspace flow`; verify the commit contains only end-to-end/integration coverage and required fixes.