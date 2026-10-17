# Design

## Context

See `proposal.md` for motivation and `specs/proxmox-devsy-provider/spec.md` for behavior requirements. The current manifest routes `exec.command` through a Go subcommand that prints a DevPod helper template; the provider does not execute that template. Devsy's provider contract instead supplies `${COMMAND}` to the configured command for execution on the machine. The existing create flow generates an SSH keypair, injects its public key through cloud-init, and identifies VMs by `MACHINE_ID` and the `devpod-<machine-id>` name.

## Goals / Non-Goals

**Goals:**
- Make the Devsy machine provider's command path execute the requested remote command over SSH and preserve the agent's stdio protocol.
- Keep Proxmox provisioning and lifecycle management, and maintain access to existing VMs where their identity is still discoverable.
- Ensure the documented configuration flow leaves a valid provider initialized after the required Proxmox options are supplied.

**Non-Goals:**
- Preserve compatibility with the DevPod CLI or its helper protocol.
- Invoke privileged Proxmox host setup implicitly from `devsy provider add`; host setup remains a separate, administrator-run root step.
- Create or reconfigure Proxmox network bridges or storage pools; the setup script uses the selected resources that already exist.
- Rename the Linux guest account or existing Proxmox resources solely to match the Devsy product name.

## Decisions

1. **Use Devsy's agent and command contract directly.** Configure the agent with the Devsy binary provided by the CLI and pass the Devsy-supplied `${COMMAND}` to the provider's machine command operation. That operation will resolve the VM from the machine identity, obtain its reachable IP from Proxmox, invoke SSH with the generated private key and existing guest user, and connect stdin/stdout/stderr directly to the SSH process. It must not print a command template to stdout. This keeps the dynamic VM lookup in the existing provider, rather than embedding an unknown host in a static manifest shell script; unit tests will cover command execution and byte-preserving stdio behavior.

2. **Use Devsy names for new resources and preserve legacy lookup.** Make new account/token, cloud-init template, guest user, VM prefix, and tags Devsy-branded. Keep lookup and lifecycle handling for existing `devpod-<machine-id>` VMs, and make setup/uninstall recognize both generations. Do not rename or delete legacy resources automatically; use an explicit preference/order when a machine could match more than one resource so resume never creates or selects a duplicate unexpectedly.

3. **Keep privileged host bootstrap separate from provider initialization.** Preserve the root-run Proxmox setup script as an explicit step before adding/configuring the provider in Devsy. It creates or verifies the service account, API token/ACLs, and cloud-init template; it uses the selected existing bridge and storage pool without creating or reconfiguring them. Devsy provider `init` then validates the configured options, API reachability, node, and prepared template, returning a useful error on failure. Document this order and keep the privileged setup out of `devsy provider add`.

4. **Verify against Devsy end to end, not only manifest text.** Add focused Go tests for command forwarding, lifecycle mapping, and initialization, plus a Devsy-driven acceptance path that installs/configures the provider, creates a VM, confirms agent/SSH command execution, resumes it, and deletes it. The live Proxmox test must be environment-gated because it requires a real API endpoint, credentials, and template.

## Risks / Trade-offs

- **Devsy changes its provider contract** → Pin the documented compatibility target to a tested Devsy CLI release, keep manifest-level contract tests, and update them when supporting a newer CLI.
- **VM IP discovery or SSH readiness is slower than expected** → Reuse guest-agent polling with bounded timeouts and return actionable errors; do not emit diagnostics to the agent's stdout stream.
- **Existing VMs are not discoverable under the Devsy machine identity** → Preserve the current machine-ID/name mapping and test resume against a VM created before migration; never silently clone when lookup fails during resume.
- **SSH host-key handling weakens security for ephemeral VM addresses** → Keep key authentication and make host-key verification behavior explicit and narrowly scoped; do not add a global SSH configuration change.
- **Root setup is run against the wrong host or has incomplete prerequisites** → Keep it an explicit administrator action, preserve its dry-run and idempotency behavior, and validate selected existing node, bridge, and storage before creating the template.

## Migration Plan

1. On the intended Proxmox host, an administrator runs the existing setup script as root to create/verify the service account, API token/ACLs, and cloud-init template. Select a bridge and storage pool that already exist; retain the generated token.
2. Install/select the provider through Devsy and supply the Proxmox options, including the prepared template; provider initialization must complete before workspace creation. It must not invoke the privileged host script itself.
3. Resume an existing provider VM and verify Devsy agent injection/SSH before creating new workspaces. Keep rollback possible by retaining the VM and its disk until the migrated provider has passed validation.
4. After validation, use the Devsy provider for normal create, stop, start, and delete operations. New resources use Devsy names; existing DevPod-named resources remain manageable until an administrator chooses to remove them. Rollback is to restore the prior provider source/CLI for users who still require DevPod; the migrated release itself does not promise DevPod CLI compatibility.

## Open Questions

- Which Devsy release will be declared the minimum supported version? Confirm during implementation against the CLI used for the integration test.
- Can the existing Proxmox test environment support a repeatable live end-to-end job in CI, or must that acceptance test remain an opt-in operator-run check?