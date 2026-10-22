# Devsy Proxmox Provider

A [Devsy](https://github.com/mkdev-ai/devsy-proxmox-provider) machine provider that provisions KVM virtual machines on [Proxmox VE](https://www.proxmox.com/en/proxmox-virtual-environment) to host development workspaces.

Each machine gets a dedicated VM cloned from a cloud-init template. Devsy installs Docker on the VM and runs your devcontainer inside it — the VM is the machine, the workspace is a container on top of it.

## Requirements

- Proxmox VE 7.0 or later
- Devsy CLI installed
- An existing network bridge and storage pool for the template and workspace VMs
- The existing `local` storage configured to support snippets for the cloud-init vendor file

## End-to-End Acceptance Test

The repository includes an environment-gated end-to-end acceptance test
(`cmd/acceptance_test.go`) that uses an emulated Proxmox API to exercise the
full provider lifecycle and verify that pre-existing legacy DevPod-named VMs
are never discovered, adopted, or modified. It is environment-gated: when
`DEVSY_PROXMOX_HOST`, `DEVSY_PROXMOX_PORT`, `DEVSY_PROXMOX_USER`,
`DEVSY_PROXMOX_TOKEN`, `DEVSY_PROXMOX_NODE`, and `DEVSY_PROXMOX_TEMPLATE` are
**unset**, the test is **skipped**, keeping CI green when no Proxmox
credentials are available.

Run the test locally with the emulated Proxmox API:

```bash
# Full suite + acceptance-tagged suite (skips by default, no credentials needed)
go test ./...
go test -tags=acceptance ./...

# Run only the acceptance test
go test -tags=acceptance -v -run TestDevsyAcceptance
```

Behavior: the emulated server answers the subset of the Proxmox API the
provider needs (cluster resources, VM clone/configure/resize/start/stop/shutdown/
delete/status + network interfaces + task polling). The test verifies:

- A pre-existing legacy `devpod-*-` VM is never selected, cloned, started, or deleted.
- The provider creates a Devsy-named VM (`devsy-<machine-id>`), starts it, runs an
  agent command over SSH, stops it, reports status, and deletes it.
- Cleanups run after every subtest, even on failure.

The emulated server is in-process; no real Proxmox endpoint or credentials are
needed to run it. To run it against a real Proxmox host, set the `DEVSY_PROXMOX_*`
variables and re-run with `-tags=acceptance` (see the [Local live endpoint](#local-live-endpoint) section).

## Local live endpoint (optional)

To run the acceptance test against a real Proxmox host instead of the embedded
emulated server, set the environment variables below and re-run with
`-tags=acceptance`:

```bash
export DEVSY_PROXMOX_HOST=your-proxmox-host
export DEVSY_PROXMOX_PORT=8006
export DEVSY_PROXMOX_USER=devsy@pve
export DEVSY_PROXMOX_TOKEN='devsy@pve!devsy=<secret>'
export DEVSY_PROXMOX_NODE=pve
export DEVSY_PROXMOX_TEMPLATE=devsy-ubuntu-24-04
go test -tags=acceptance -v -run TestDevsyAcceptance
```

The test is skipped automatically when any of the variables above is unset, so
CI stays green without live Proxmox credentials.

## Setup Script

`scripts/setup.sh` prepares a Proxmox node for use with this provider. Run it manually, once as root on the Proxmox host before configuring the provider in Devsy. Provider installation does not invoke this privileged script. It is idempotent — safe to re-run.

```bash
# Run directly on the host
sudo bash scripts/setup.sh --node pve --storage local-lvm

# Or pipe over SSH with direct input (no SSH connection line)
| bash -s -- --node pve --storage local-lvm < scripts/setup.sh

# Or fully remote — fetch and run directly, no clone needed
curl -fsSL https://raw.githubusercontent.com/mkdev-ai/devsy-proxmox-provider/main/scripts/setup.sh | bash -s -- --node pve --storage local-lvm
```

### What it does

1. Creates a `devsy@pve` service account and `devsy@pve!devsy` API token
2. Grants the minimum required ACL roles (see [Permissions](#permissions))
3. Downloads and verifies an Ubuntu cloud image
4. Creates a cloud-init VM template with QEMU guest agent pre-installed (default: `devsy-ubuntu-24-04`)

The `--bridge` and `--storage` values must name existing Proxmox resources. Setup uses them for the template and does not create or reconfigure them. The `local` storage must already support snippets; setup checks this prerequisite and does not change storage configuration.

### Options

| Flag | Default | Description |
|---|---|---|
| `--node` | `pve` | Proxmox node name |
| `--storage` | `local-lvm` | Storage pool for the template disk |
| `--bridge` | `vmbr0` | Network bridge |
| `--vmid` | `9000` | VMID for the template |
| `--image` | `ubuntu:24.04` | Distro and version (`ubuntu:20.04`, `22.04`, `24.04`, `25.04`) |
| `--disk-format` | `raw` | `raw` for LVM/Ceph; `qcow2` for directory-backed storage (local, NFS) |
| `--template-name` | `devsy-<distro>-<version>` | Override the template name |
| `--dry-run` | | Print what would be done without making any changes |

### Uninstalling

To remove Devsy-named resources managed by setup (VMs, template, service account, token):

```bash
# Pipe over SSH with direct input (no SSH connection line)
| bash -s -- < scripts/uninstall.sh

# Or fully remote — fetch and run directly, no clone needed
curl -fsSL https://raw.githubusercontent.com/mkdev-ai/devsy-proxmox-provider/main/scripts/uninstall.sh | bash
```

Uninstall targets only `devsy-*` VMs/snippets and the `devsy@pve` account/token. Legacy DevPod-created VMs and credentials are unsupported, are not adopted, and are never removed by this script.

## Provider Options

Set options with `devsy provider set-options proxmox -o KEY=VALUE`.

### Required

| Option | Description |
|---|---|
| `PROXMOX_HOST` | Proxmox VE hostname or IP address |
| `PROXMOX_USER` | Proxmox user (e.g. `devsy@pve`) |
| `PROXMOX_TOKEN` | API token in the format `USER@REALM!TOKENID=SECRET` |
| `PROXMOX_NODE` | Proxmox node name to create VMs on |
| `PROXMOX_TEMPLATE` | Name of the cloud-init VM template created by `setup.sh` |

### Optional

| Option | Default | Description |
|---|---|---|
| `PROXMOX_PORT` | `8006` | Proxmox API port |
| `PROXMOX_STORAGE` | `local-lvm` | Existing storage pool for VM disks; not created or reconfigured by the provider |
| `PROXMOX_NETWORK` | `vmbr0` | Existing network bridge to attach VMs to; not created or reconfigured by the provider |
| `PROXMOX_VM_START_ID` | `2000` | Starting VMID for workspace VMs (increments until a free ID is found) |
| `VM_MEMORY` | `4096` | VM memory in MB |
| `VM_CORES` | `2` | Number of vCPU cores |
|| `VM_DISK_SIZE` | `50` | Workspace VM root disk size in GB. The template is pre-sized at template creation (see `--template-disk-size` in setup.sh); this clamps the cloned VM to this size before install. |

## Usage

```bash
# Create and start a workspace
devsy up git@github.com:myorg/myrepo --provider proxmox

# Stop a workspace (VM is shut down, data persists)
devsy stop my-workspace

# Resume a stopped workspace
devsy up my-workspace

# Delete a workspace and its VM
devsy delete my-workspace
```

New workspace VMs are named `devsy-{machine-id}` and tagged `devsy` in Proxmox. This provider does not discover, adopt, resume, rename, or delete legacy DevPod-named VMs.

## How It Works

1. `devsy up` calls the provider's `create` command
2. The provider clones the cloud-init template to a new VM
3. Cloud-init is applied: SSH key, DHCP network, `devsy` user
4. The clone is resized to `VM_DISK_SIZE` **before** the VM starts, so the workspace install (git clone / docker pull) always has the requested capacity
5. The VM is started; the provider polls until the QEMU guest agent reports an IP
6. Devsy connects via SSH, installs its agent, and launches the devcontainer
7. Your IDE connects to the workspace over SSH

## Permissions

The provider should never run with live root access. Use this least-privilege setup:

1. Create a dedicated service account on the Proxmox host (`root@pam` only):

   ```bash
   pveum user create devsy@pve -comment="Devsy provider service account"
   ```

2. Grant only the minimum node-scoped roles:

   ```bash
   pveum acl set /nodes/<NODE> -user devsy@pve -roles PVEVMUser,Sys.Audit --propagate 1
   ```

   - `PVEVMUser` covers VM lifecycle operations the provider performs (clone, configure, resize, start, stop, shutdown, delete, status, guest-agent network info).
   - `Sys.Audit` covers cluster/resource enumeration needed for `findVMByMachineID`.
   - No `Storage.*`, `Network.*`, `User.*`, or host-level operations are granted.

3. Create an API token for that account (never trust a password):

   ```bash
   pveum user token add devsy@pve devsy-proxmox-provider
   ```

   Store the printed `devsy@pve!devsy=<SECRET>` value securely — it is returned only once and cannot be retrieved later.

4. (Optional) Restrict the API to a dedicated management interface and source IPs. Serve the Proxmox API over TLS with a proper certificate.

Local verification (no live Proxmox needed) — run `bash -n scripts/setup.sh scripts/uninstall.sh` and review the ACL/role semantics above. The `pveum` commands are standard Proxmox practice and do not require a live endpoint to reason about.

## Release and Distribution

Every tagged release publishes **four compiled binaries** (one per platform) plus a
`provider.yaml` manifest that ties them together with checksums. This is what Devsy
fetches when installing the provider — it is **not** the raw source tree.

### What a release contains

| Asset | OS | Arch | Format |
|---|---|---|---|
| `proxmox-provider-linux-amd64` | Linux | amd64 | ELF, statically linked |
| `proxmox-provider-linux-arm64` | Linux | arm64 | ELF, statically linked |
| `proxmox-provider-darwin-amd64` | macOS | amd64 | Mach-O, x86_64 |
| `proxmox-provider-darwin-arm64` | macOS | arm64 | Mach-O, arm64 |

Each binary is built with `CGO_ENABLED=0` and linked statically, so it has no
shared-library dependencies and runs anywhere that platform is available.

### Installing the provider (recommended)

Devsy installs the provider by downloading the manifest from the release and
verifying each binary against its SHA-256 checksum:

```bash
# Install the latest provider on your machine
devsy provider add github.com/mkdev-ai/devsy-proxmox-provider
```

Devsy reads `provider.yaml` from the release; the checksums let it prove the
platform binary has not been tampered with. You do not need to build from source.

### Building and publishing a release (developers)

```bash
# Build all 4 platform binaries (CGO_ENABLED=0, static, version pinned)
GOOS=linux GOARCH=amd64 go build -ldflags="-X github.com/mkdev-ai/devsy-proxmox-provider/cmd.versionStr=v0.2.0" -o dist/proxmox-provider-linux-amd64 .
GOOS=linux GOARCH=arm64 go build -ldflags="-X github.com/mkdev-ai/devsy-proxmox-provider/cmd.versionStr=v0.2.0" -o dist/proxmox-provider-linux-arm64 .
GOOS=darwin GOARCH=amd64 go build -ldflags="-X github.com/mkdev-ai/devsy-proxmox-provider/cmd.versionStr=v0.2.0" -o dist/proxmox-provider-darwin-amd64 .
GOOS=darwin GOARCH=arm64 go build -ldflags="-X github.com/mkdev-ai/devsy-proxmox-provider/cmd.versionStr=v0.2.0" -o dist/proxmox-provider-darwin-arm64 .

# Generate provider.yaml with checksums
python3 scripts/generate_provider_yaml.py

# Commit and tag, then push the tag -- the CI release workflow publishes assets
git add provider.yaml provider.yaml.tmpl
git commit -m "release(v0.2.0): build 4-platform provider binaries and manifest"
git tag v0.2.0
git push origin v0.2.0
```

### Building from source (alternative)

```bash
git clone https://github.com/mkdev-ai/devsy-proxmox-provider
cd devsy-proxmox-provider
go build -o dist/proxmox-provider .
go test ./...
```

## Contributing

Bug reports and pull requests are welcome. Please open an issue before starting significant work so we can discuss the approach.
