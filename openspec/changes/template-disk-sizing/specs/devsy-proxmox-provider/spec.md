# Spec Delta

## Purpose

Defines the observable contract for template disk sizing in the Proxmox provider so the workspace is built at the requested capacity from the start, not the small cloud-image default.

## ADDED Requirements

### Requirement: Template base disk sized at creation

The Proxmox template's root disk SHALL be created at a configurable size, defaulting to the Ubuntu cloud image default, and NOT at the small cloud-image default when a larger size is requested.

#### Scenario: Template created at the desired size
- **WHEN** an administrator runs `scripts/setup.sh --template-disk-size <size-mib>`
- **THEN** the template's root disk (`virtio0`) is created at exactly that capacity, not at the Ubuntu cloud image default of ~3.5 GB (3584 MiB).

#### Scenario: Default template retains the cloud image default
- **WHEN** an administrator runs `scripts/setup.sh` without `--template-disk-size`
- **THEN** the template base disk is the Ubuntu cloud image default (3584 MiB), and the user is told to pass `--template-disk-size` if they want more.

### Requirement: Workspace VM resized before install

The provider SHALL resize the cloned workspace VM to `VM_DISK_SIZE` immediately after cloning and **before** the VM starts, so the git clone and Docker install that follow always have the requested capacity.

#### Scenario: Clone sized to the requested capacity before install
- **WHEN** a workspace is created with `VM_DISK_SIZE` greater than the template base
- **THEN** the provider resizes the clone to exactly `VM_DISK_SIZE` before starting the VM, and the git clone / docker pull that follow see adequate disk space.

#### Scenario: Resize runs before start, not after
- **WHEN** the provider starts the clone and runs the workspace install
- **THEN** the resize task completes (or is rejected with an actionable error) before the VM starts; the install never races the resize.

### Requirement: Template rebuild documented

The template base disk can only be changed by rebuilding the template on the Proxmox host; the provider MUST NOT resize an existing template in place.

#### Scenario: Template rebuild required to change base size
- **WHEN** the desired workspace disk size differs from the current template base
- **THEN** the administrator rebuilds the template with `scripts/setup.sh --template-disk-size <size-mib>` on the Proxmox host before creating new workspaces.
