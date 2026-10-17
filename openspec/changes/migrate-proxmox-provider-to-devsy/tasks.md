# Tasks

## 1. Devsy Manifest and SSH Transport

- [ ] 1.1 Update `provider.yaml.tmpl` for the Devsy agent and machine-provider contract, use Devsy agent paths/data paths and descriptions, and add a manifest contract test that rejects obsolete CLI syntax and checks all required lifecycle commands.
- [ ] 1.2 Replace the command-template output with an SSH command runner that resolves the VM from `MACHINE_ID`, uses its current IP and generated key, and relays stdin/stdout/stderr and exit status; verify it with fake SSH and Proxmox clients, including that stdout contains no provider diagnostics.
- [ ] 1.3 Update user-visible provider command descriptions, errors, and help text to Devsy terminology; add focused checks that no obsolete product or CLI wording remains outside explicit legacy-compatibility references.
- [ ] 1.4 After tasks 1.1–1.3 pass, commit only the manifest/SSH/help slice as `feat(provider): migrate SSH transport to Devsy`; verify the commit summary contains only this slice and its tests.

## 2. Lifecycle and Provider Readiness

- [ ] 2.1 Verify Devsy's machine create-result and environment contract; use Devsy-branded names for new accounts, tokens, templates, guest users, tags, and VMs while retaining lookup/lifecycle support for existing `devpod-<machine-id>` VMs; cover both naming generations and create/resume/status/stop/delete with focused unit tests.
- [ ] 2.2 Make provider initialization validate required options, Proxmox API access, node, and template; add tests proving valid configuration succeeds and missing/unavailable prerequisites fail with actionable errors, then run `go test ./cmd/... ./pkg/options/... ./pkg/proxmox/...`.
- [ ] 2.3 After tasks 2.1–2.2 pass, commit only lifecycle, resource naming, and readiness changes as `feat(provider): add Devsy lifecycle support`; verify the commit contains its implementation and tests only.

## 3. Devsy Installation and Documentation

- [ ] 3.1 Update every user-facing document and install/configure/workspace command example to Devsy terminology and syntax; document root-running `scripts/setup.sh`, its Devsy-branded defaults, token handling, existing bridge/storage prerequisites, and legacy resource compatibility; point self-referencing install/docs links at the current canonical repo when the companion module/URL alignment is applied.
- [ ] 3.2 Update setup/uninstall script help and output to Devsy branding and Devsy defaults; retain safe detection/removal behavior for both Devsy and legacy resource identifiers, verify bridge/storage are only used (not created/reconfigured), reruns remain idempotent, and run `bash -n scripts/setup.sh scripts/uninstall.sh` plus a mocked or dry-run check.
- [ ] 3.3 After tasks 3.1–3.2 pass, commit only documentation and setup/uninstall changes as `docs(provider): document Devsy setup and legacy resources`; verify no provider implementation files are included.

## 4. Devsy Integration Verification

- [ ] 4.1 Establish and document the minimum supported Devsy CLI release, then validate provider installation and manifest initialization with that release and the current supported release.
- [ ] 4.2 Add an environment-gated Proxmox acceptance test that configures the provider, creates a workspace, runs an SSH command through the injected agent, resumes the existing VM, and deletes it; guarantee cleanup and document how to run it when CI lacks Proxmox credentials.
- [ ] 4.3 Run `go test ./...` and the Devsy acceptance test where Proxmox credentials are available; record the tested Devsy version and confirm the original `unexpected start line` failure no longer occurs.
- [ ] 4.4 Coordinate with `rename-proxmox-provider-to-devsy` to set the Go module path/imports and self-referencing provider/docs URLs to `github.com/mkdev-ai/devsy-proxmox-provider`; leave upstream `iamveen/devpod-proxmox-provider` untouched.
- [ ] 4.5 After task 4.3 passes, commit any remaining integration-only changes separately as `test(provider): verify Devsy Proxmox workspace flow`; verify the commit contains only end-to-end/integration coverage and required fixes.