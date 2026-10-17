# Design

## Context

See `proposal.md` for motivation and `specs/devsy-proxmox-provider/spec.md` for behavior requirements. The current manifest routes `exec.command` through a Go subcommand that prints a DevPod helper template; the provider does not execute that template. Devsy's provider contract instead supplies `${COMMAND}` to the configured command for execution on the machine. The new create flow will use Devsy-branded identifiers; existing DevPod-named VMs are explicitly outside the migrated provider's lookup and lifecycle scope.

## Goals / Non-Goals

**Goals:**
- Make the Devsy machine provider's command path execute the requested remote command over SSH and preserve the agent's stdio protocol.
- Keep Proxmox provisioning and lifecycle management for newly created Devsy-named resources only.
- Ensure the documented configuration flow leaves a valid provider initialized after the required Proxmox options are supplied.

**Non-Goals:**
- Preserve compatibility with the DevPod CLI or its helper protocol.
- Adopt, resume, rename, delete, or automatically migrate VMs and credentials created with DevPod-era identifiers.
- Invoke privileged Proxmox host setup implicitly from `devsy provider add`; host setup remains a separate, administrator-run root step.
- Create or reconfigure Proxmox network bridges or storage pools; the setup script uses the selected resources that already exist.
- Rename the Linux guest account or existing Proxmox resources solely to match the Devsy product name.

## Decisions

1. **Use Devsy's agent and command contract directly.** Configure the agent with the Devsy binary provided by the CLI and pass the Devsy-supplied `${COMMAND}` to the provider's machine command operation. That operation will resolve the VM from the machine identity, obtain its reachable IP from Proxmox, invoke SSH with the generated private key and existing guest user, and connect stdin/stdout/stderr directly to the SSH process. It must not print a command template to stdout. This keeps the dynamic VM lookup in the existing provider, rather than embedding an unknown host in a static manifest shell script; unit tests will cover command execution and byte-preserving stdio behavior.

2. **Use Devsy-only identifiers.** Make new account/token, cloud-init template, guest user, VM prefix, and tags Devsy-branded. All provider lookup and lifecycle operations use only Devsy identifiers; never probe a `devpod-*` alias or adopt a legacy VM. Leave any legacy account, token, template, or VM untouched, and keep it outside setup/uninstall targeting. This deliberately gives up automatic continuity to prevent a Devsy workspace from silently attaching to a previous machine.

3. **Keep privileged host bootstrap separate from provider initialization.** Preserve the root-run Proxmox setup script as an explicit step before adding/configuring the provider in Devsy. It creates or verifies the service account, API token/ACLs, and cloud-init template; it uses the selected existing bridge and storage pool without creating or reconfiguring them. Devsy provider `init` then validates the configured options, API reachability, node, and prepared template, returning a useful error on failure. Document this order and keep the privileged setup out of `devsy provider add`.

4. **Verify against Devsy end to end, not only manifest text.** Add focused Go tests for command forwarding, lifecycle mapping, and initialization, plus a Devsy-driven acceptance path that installs/configures the provider, creates a VM, confirms agent/SSH command execution, resumes it, and deletes it. The live Proxmox test must be environment-gated because it requires a real API endpoint, credentials, and template.

## Risks / Trade-offs

- **Devsy changes its provider contract** → Pin the documented compatibility target to a tested Devsy CLI release, keep manifest-level contract tests, and update them when supporting a newer CLI.
- **VM IP discovery or SSH readiness is slower than expected** → Reuse guest-agent polling with bounded timeouts and return actionable errors; do not emit diagnostics to the agent's stdout stream.
- **Existing DevPod VMs cannot be resumed through the migrated provider** → State clearly that legacy machines are unsupported and left untouched; validate that Devsy operations do not select or modify them. Any manual data recovery or migration is outside this change.
- **SSH host-key handling weakens security for ephemeral VM addresses** → Keep key authentication and make host-key verification behavior explicit and narrowly scoped; do not add a global SSH configuration change.
- **Root setup is run against the wrong host or has incomplete prerequisites** → Keep it an explicit administrator action, preserve its dry-run and idempotency behavior, and validate selected existing node, bridge, and storage before creating the template.

## Migration Plan

1. On the intended Proxmox host, an administrator runs the existing setup script as root to create/verify the service account, API token/ACLs, and cloud-init template. Select a bridge and storage pool that already exist; retain the generated token.
2. Install/select the provider through Devsy and supply the Proxmox options, including the prepared template; provider initialization must complete before workspace creation. It must not invoke the privileged host script itself.
3. Create a new Devsy-named VM and verify agent injection/SSH before using the provider for new workspaces. Existing DevPod-named VMs remain untouched and are not resumed by this provider.
4. After validation, use the Devsy provider for normal create, stop, start, and delete operations on Devsy-named resources only. Do not fall back to the prior provider, its CLI, or its resource identifiers.

## Open Questions

- Which Devsy release will be declared the minimum supported version? Confirm during implementation against the CLI used for the integration test.
- Can the existing Proxmox test environment support a repeatable live end-to-end job in CI, or must that acceptance test remain an opt-in operator-run check?