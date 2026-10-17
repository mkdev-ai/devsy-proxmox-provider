# Devsy Proxmox Provider

A [Devsy](https://github.com/mkdev-ai/devsy-proxmox-provider) machine provider that provisions KVM virtual machines on [Proxmox VE](https://www.proxmox.com/en/proxmox-virtual-environment) to host development workspaces.

Each machine gets a dedicated VM cloned from a cloud-init template. Devsy installs Docker on the VM and runs your devcontainer inside it — the VM is the machine, the workspace is a container on top of it.

## Requirements

- Proxmox VE 7.0 or later
- Devsy CLI installed
- SSH root access to the Proxmox host (for one-time setup)
- The target node reachable from your machine over HTTPS (port 8006)
- An existing network bridge and storage pool for the template and workspace VMs
- The existing `local` storage configured to support snippets for the cloud-init vendor file

## Quick Start

### 1. Prepare Proxmox

Run the setup script once on the Proxmox host as root. The script creates or verifies only its Devsy account, API token, ACLs, and cloud-init template. It does not create or reconfigure the selected network bridge or storage pool; choose resources that already exist. The Proxmox `local` storage must already support snippets.

```bash
curl -fsSL https://raw.githubusercontent.com/mkdev-ai/devsy-proxmox-provider/main/scripts/setup.sh \
  | ssh root@proxmox-host 'bash -s --' \
      --node pve \
      --storage local-lvm \
      --bridge vmbr0
```

The default account is `devsy@pve`, its API token ID is `devsy`, and the template is `devsy-ubuntu-24-04`. The script prints a newly created API token secret once; store it securely because it cannot be retrieved later. Existing DevPod-created machines and credentials are unsupported by this provider and are left untouched.

### 2. Install the provider

```bash
devsy provider add github.com/mkdev-ai/devsy-proxmox-provider
devsy provider use proxmox
```

### 3. Configure it

```bash
devsy provider set-options proxmox \
  -o PROXMOX_HOST=your-proxmox-host \
  -o PROXMOX_USER=devsy@pve \
  -o PROXMOX_TOKEN='devsy@pve!devsy=<secret>' \
  -o PROXMOX_NODE=pve \
  -o PROXMOX_TEMPLATE=devsy-ubuntu-24-04
```

### 4. Start a workspace

```bash
devsy up git@github.com:myorg/myrepo --provider proxmox
```

## Setup Script

`scripts/setup.sh` prepares a Proxmox node for use with this provider. Run it manually, once as root on the Proxmox host before configuring the provider in Devsy. Provider installation does not invoke this privileged script. It is idempotent — safe to re-run.

```bash
# Run directly on the host
sudo bash scripts/setup.sh [options]

# Or pipe over SSH
ssh root@proxmox-host 'bash -s -- [options]' < scripts/setup.sh
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
ssh root@proxmox-host 'bash -s' < scripts/uninstall.sh
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
| `VM_DISK_SIZE` | `50` | VM disk size in GB |

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
4. If `VM_DISK_SIZE` exceeds the template disk, the disk is resized
5. The VM is started; the provider polls until the QEMU guest agent reports an IP
6. Devsy connects via SSH, installs its agent, and launches the devcontainer
7. Your IDE connects to the workspace over SSH

## Permissions

`setup.sh` creates a `devsy@pve` user and grants these roles:

| Path | Role | Key privileges |
|---|---|---|
| `/vms` | `PVEVMAdmin` | Clone, create, delete, configure, start/stop VMs |
| `/nodes/<node>` | `PVEAuditor` | Read node/network/storage info (`Sys.Audit`) |
| `/storage/<storage>` | `PVEDatastoreUser` | Allocate disk space |
| `/sdn/zones` | `PVESDNUser` | Use SDN-managed network bridges (`SDN.Use`) |

Both the user and token receive identical grants — required because Proxmox privilege separation (`privsep=1`) computes effective permissions as the intersection of user and token ACLs.

## Building from Source

Requires Go 1.22+.

```bash
git clone https://github.com/mkdev-ai/devsy-proxmox-provider
cd devsy-proxmox-provider
go build -o dist/proxmox-provider .
```

Run tests:

```bash
go test ./...
```

## Contributing

Bug reports and pull requests are welcome. Please open an issue before starting significant work so we can discuss the approach.
