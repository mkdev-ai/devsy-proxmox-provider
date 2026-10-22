#!/usr/bin/env bash
# setup.sh — Configure Proxmox for the Devsy provider.
# Legacy DevPod resources are unsupported and are intentionally left untouched.
# Run as root on the Proxmox host.
#
# Usage:
#   bash setup.sh [options]
#
# Options:
#   --node          Proxmox node name          (default: pve)
#   --storage       Storage pool               (default: local-lvm)
#   --bridge        Network bridge             (default: vmbr0)
#   --vmid          Template VMID              (default: 9000)
#   --image         distro:version             (default: ubuntu:24.04)
#   --disk-format   raw | qcow2               default: raw
#                   raw  — required for LVM/Ceph storage
#                   qcow2 — use for directory-backed storage (local, NFS)
#   --template-disk-size   template root disk size in MiB    default: 3584
#   --template-name   Override template name                 default: devsy-<distro>-<version>
#   --dry-run       Print what would be done without making any changes
#
# Example (run directly on the host):
#   sudo bash setup.sh --node pve --storage local-lvm --bridge vmbr0 --image ubuntu:24.04
#
# Example (pipe over SSH):
#   ssh root@proxmox-host 'bash -s -- --node pve --storage local-lvm' < setup.sh
set -euo pipefail

# ── defaults ──────────────────────────────────────────────────────────────────
NODE="pve"
STORAGE="local-lvm"
BRIDGE="vmbr0"
VMID="9000"
IMAGE="ubuntu:24.04"
DISK_FORMAT="raw"
TEMPLATE_DISK_SIZE="3584"   # template root disk size in MiB (Ubuntu cloud image default)
TEMPLATE_NAME=""
DRY_RUN=0

DEVSY_USER="devsy@pve"
DEVSY_TOKEN_NAME="devsy"
DEVSY_TOKEN_ID="${DEVSY_USER}!${DEVSY_TOKEN_NAME}"

# ── helpers ───────────────────────────────────────────────────────────────────
die()  { echo "ERROR: $*" >&2; exit 1; }
log()  { echo "    $*"; }
step() { echo ""; echo "==> $*"; }

# run: execute a command normally, or print it in dry-run mode.
run() {
    if [[ $DRY_RUN -eq 1 ]]; then
        log "[dry-run] $*"
    else
        "$@"
    fi
}

usage() {
    sed -n '/^# Usage:/,/^[^#]/p' "$0" | grep '^#' | sed 's/^# \?//'
    exit 0
}

# ── argument parsing ──────────────────────────────────────────────────────────
while [[ $# -gt 0 ]]; do
    case "$1" in
        --node)          NODE="$2";          shift 2 ;;
        --storage)       STORAGE="$2";       shift 2 ;;
        --bridge)        BRIDGE="$2";        shift 2 ;;
        --vmid)          VMID="$2";          shift 2 ;;
        --image)         IMAGE="$2";         shift 2 ;;
        --disk-format)   DISK_FORMAT="$2";   shift 2 ;;
        --template-disk-size) TEMPLATE_DISK_SIZE="$2"; shift 2 ;;
        --template-name) TEMPLATE_NAME="$2"; shift 2 ;;
        --dry-run)       DRY_RUN=1;          shift   ;;
        --help|-h)       usage ;;
        *) die "Unknown option: $1" ;;
    esac
done

# ── parse image ───────────────────────────────────────────────────────────────
DISTRO="${IMAGE%%:*}"
VERSION="${IMAGE##*:}"

[[ "$DISTRO" != "ubuntu" ]] && die "Unsupported distro '${DISTRO}'. Only 'ubuntu' is supported."

ubuntu_codename() {
    case "$1" in
        20.04) echo "focal"  ;;
        22.04) echo "jammy"  ;;
        24.04) echo "noble"  ;;
        25.04) echo "plucky" ;;
        *) die "Unsupported Ubuntu version: $1. Supported: 20.04, 22.04, 24.04, 25.04" ;;
    esac
}

CODENAME=$(ubuntu_codename "$VERSION")
VERSION_SLUG="${VERSION//./-}"

[[ -z "$TEMPLATE_NAME" ]] && TEMPLATE_NAME="devsy-${DISTRO}-${VERSION_SLUG}"

IMG_URL="https://cloud-images.ubuntu.com/${CODENAME}/current/${CODENAME}-server-cloudimg-amd64.img"
IMG_PATH="/tmp/${CODENAME}-server-cloudimg-amd64.img"
VENDOR_YAML="/var/lib/vz/snippets/${TEMPLATE_NAME}-vendor.yaml"

# ── preflight ─────────────────────────────────────────────────────────────────
[[ "$EUID" -ne 0 ]] && die "Must be run as root. Try: sudo bash $0"
command -v pveum &>/dev/null || die "pveum not found — run this script on a Proxmox VE host"
command -v qm    &>/dev/null || die "qm not found — run this script on a Proxmox VE host"

[[ "$DISK_FORMAT" != "raw" && "$DISK_FORMAT" != "qcow2" ]] && \
    die "Invalid --disk-format '${DISK_FORMAT}'. Use 'raw' or 'qcow2'."

step "Configuration"
log "Node:          ${NODE}"
log "Storage:       ${STORAGE}"
log "Bridge:        ${BRIDGE}"
log "VMID:          ${VMID}"
log "Image:         ${IMAGE} (codename: ${CODENAME})"
log "Disk format:   ${DISK_FORMAT}"
log "Template disk: ${TEMPLATE_DISK_SIZE} MiB"
log "Template name: ${TEMPLATE_NAME}"
[[ $DRY_RUN -eq 1 ]] && log "Mode:          dry-run (no changes will be made)"

# Refuse to touch a VMID owned by a different VM before changing any resources.
VM_PRESENT=0
VM_CONFIG=""
if [[ $DRY_RUN -eq 0 ]] && qm list 2>/dev/null | awk -v vmid="${VMID}" 'NR > 1 && $1 == vmid { found = 1 } END { exit !found }'; then
    if ! VM_CONFIG=$(qm config "${VMID}" 2>/dev/null); then
        die "Could not inspect VMID ${VMID}; no resources were changed."
    fi
    VM_NAME=$(awk '/^name: / { sub(/^name: /, ""); print; exit }' <<< "$VM_CONFIG")
    if [[ "$VM_NAME" != "$TEMPLATE_NAME" ]]; then
        die "VMID ${VMID} belongs to '${VM_NAME:-an unnamed VM}', not configured Devsy template '${TEMPLATE_NAME}'. It was not modified; choose another --vmid."
    fi
    VM_PRESENT=1
fi

# ── phase 1: service account ──────────────────────────────────────────────────
step "Phase 1: Service account (${DEVSY_USER})"

# User
if pveum user list --output-format json 2>/dev/null | grep -q "\"userid\":\"${DEVSY_USER}\""; then
    log "User ${DEVSY_USER} already exists"
else
    run pveum user add "${DEVSY_USER}" --comment "Devsy service account"
    [[ $DRY_RUN -eq 0 ]] && log "Created user ${DEVSY_USER}"
fi

# Token
TOKEN_VALUE=""
if pveum user token list "${DEVSY_USER}" --output-format json 2>/dev/null | grep -q "\"tokenid\":\"${DEVSY_TOKEN_NAME}\""; then
    log "Token ${DEVSY_TOKEN_ID} already exists"
else
    if [[ $DRY_RUN -eq 1 ]]; then
        log "[dry-run] pvesh create /access/users/${DEVSY_USER}/token/${DEVSY_TOKEN_NAME} --expire 0 --privsep 1"
    else
        TOKEN_VALUE=$(pvesh create "/access/users/${DEVSY_USER}/token/${DEVSY_TOKEN_NAME}" \
            --expire 0 --privsep 1 --output-format json \
            | grep -o '"value":"[^"]*"' | cut -d'"' -f4)
        log "Created token ${DEVSY_TOKEN_ID}"
    fi
fi

# ACLs — pvesh set /access/acl is idempotent (PUT)
# /vms — VM operations (VM.Audit, VM.Clone, VM.Config.*, VM.PowerMgmt, etc.)
run pvesh set /access/acl --path "/vms"                --roles "PVEVMAdmin"            --users  "${DEVSY_USER}"
run pvesh set /access/acl --path "/vms"                --roles "PVEVMAdmin"            --tokens "${DEVSY_TOKEN_ID}"
# /nodes/${NODE} — node-level audit (Sys.Audit for network/storage listing)
run pvesh set /access/acl --path "/nodes/${NODE}"      --roles "PVEAuditor"            --users  "${DEVSY_USER}"
run pvesh set /access/acl --path "/nodes/${NODE}"      --roles "PVEAuditor"            --tokens "${DEVSY_TOKEN_ID}"
# /storage/${STORAGE} — disk allocation for clones and cloud-init drives
run pvesh set /access/acl --path "/storage/${STORAGE}" --roles "PVEDatastoreUser"      --users  "${DEVSY_USER}"
run pvesh set /access/acl --path "/storage/${STORAGE}" --roles "PVEDatastoreUser"      --tokens "${DEVSY_TOKEN_ID}"
# /sdn/zones — SDN.Use required when cloning a VM onto a bridge managed by Proxmox SDN
run pvesh set /access/acl --path "/sdn/zones"          --roles "PVESDNUser"            --users  "${DEVSY_USER}"
run pvesh set /access/acl --path "/sdn/zones"          --roles "PVESDNUser"            --tokens "${DEVSY_TOKEN_ID}"
[[ $DRY_RUN -eq 0 ]] && log "ACL grants applied on /vms, /nodes/${NODE}, /storage/${STORAGE}, and /sdn/zones"

if [[ -n "$TOKEN_VALUE" ]]; then
    echo ""
    echo "  ┌─────────────────────────────────────────────────────────────────┐"
    echo "  │  New API token — save securely for your Devsy provider config:  │"
    echo "  │                                                                 │"
    echo "  │    export PROXMOX_TOKEN=${DEVSY_TOKEN_ID}=${TOKEN_VALUE}"
    echo "  │                                                                 │"
    echo "  └─────────────────────────────────────────────────────────────────┘"
    echo ""
fi

# ── phase 2: template ─────────────────────────────────────────────────────────
step "Phase 2: Template (${TEMPLATE_NAME}, VMID ${VMID})"

# Idempotency: matching template → done; matching partial VM → destroy and retry.
if [[ $VM_PRESENT -eq 1 ]]; then
    if grep -q "^template:" <<< "$VM_CONFIG"; then
        log "Template VMID ${VMID} already exists — nothing to do"
        echo ""
        echo "✓ Setup complete"
        exit 0
    fi
    log "Found partial VM ${VMID} (not yet a template) — destroying to recreate"
    qm destroy "${VMID}" --purge
fi

# On error after VM creation: destroy the partial VM but keep the image
PARTIAL_VM_CREATED=0
cleanup_on_error() {
    local rc=$?
    if [[ $rc -ne 0 && $PARTIAL_VM_CREATED -eq 1 ]]; then
        echo "" >&2
        echo "  Setup failed — cleaning up partial VM ${VMID}..." >&2
        qm destroy "${VMID}" --purge 2>/dev/null || true
        echo "  Partial VM removed. Re-run to retry." >&2
    fi
}
trap cleanup_on_error EXIT

# The snippet storage must already be configured; dry-run must not create files.
if [[ $DRY_RUN -eq 1 ]]; then
    log "[dry-run] Would ensure /var/lib/vz/snippets exists"
else
    mkdir -p /var/lib/vz/snippets
fi
LOCAL_CONTENT=$(pvesh get /storage/local --output-format json 2>/dev/null \
    | grep -o '"content":"[^"]*"' | cut -d'"' -f4 || echo "")
if [[ -z "$LOCAL_CONTENT" ]]; then
    die "Could not read existing local storage content. Ensure local storage exists and supports snippets; setup does not configure storage."
elif echo "$LOCAL_CONTENT" | grep -q snippets; then
    log "Snippets already enabled on local storage"
else
    die "Snippets are not enabled on existing local storage. Enable them in Proxmox first; setup does not reconfigure storage."
fi

# Write vendor cloud-init config (installs qemu-guest-agent AND pre-installs
# the Devsy agent so workspace creation is fast)
if [[ $DRY_RUN -eq 1 ]]; then
    log "[dry-run] Would write ${VENDOR_YAML}"
else
cat > "${VENDOR_YAML}" <<'EOF'
#cloud-config
package_update: true
packages:
    - qemu-guest-agent
runcmd:
    # Install qemu-guest-agent
    - systemctl enable --now qemu-guest-agent

    # Pre-install the Devsy agent binary so workspace creation is fast
    # (skips the ~10 min download+install during create)
    - |
        set -e
        DEVSY_VERSION="v1.24.0"
        DEVSY_URL="https://github.com/devsy-org/devsy/releases/download/v1.24.0/devsy-linux-amd64"
        DEVSY_PATH="/usr/local/bin/devsy"
        if [ ! -x "$DEVSY_PATH" ]; then
            echo "Downloading Devsy agent v1.24.0..."
            curl -fsSL "$DEVSY_URL" -o "$DEVSY_PATH"
            chmod +x "$DEVSY_PATH"
            echo "Devsy agent installed to $DEVSY_PATH"
        else
            echo "Devsy agent already installed at $DEVSY_PATH"
        fi
EOF
    log "Written ${VENDOR_YAML}"
fi

# Download cloud image and verify checksum
IMG_FILENAME="${CODENAME}-server-cloudimg-amd64.img"
CHECKSUMS_URL="https://cloud-images.ubuntu.com/${CODENAME}/current/SHA256SUMS"

if [[ $DRY_RUN -eq 1 ]]; then
    log "[dry-run] Would download ${IMG_URL} to ${IMG_PATH}"
    log "[dry-run] Would verify checksum from ${CHECKSUMS_URL}"
else
    log "Downloading ${IMG_URL} ..."
    wget -q -O "${IMG_PATH}" "${IMG_URL}"
    log "Image downloaded"

    log "Verifying checksum ..."
    EXPECTED=$(wget -q -O - "${CHECKSUMS_URL}" | grep "${IMG_FILENAME}" | awk '{print $1}')
    [[ -z "$EXPECTED" ]] && die "Could not retrieve checksum for ${IMG_FILENAME}"
    ACTUAL=$(sha256sum "${IMG_PATH}" | awk '{print $1}')
    [[ "$EXPECTED" == "$ACTUAL" ]] || die "Checksum mismatch — download is corrupt (expected ${EXPECTED}, got ${ACTUAL})"
    log "Checksum OK"
fi

# Create VM (4 cores / 4GB to speed up cloud-init and agent install)
run qm create "${VMID}" \
    --name "${TEMPLATE_NAME}" \
    --memory 4096 \
    --cores 4 \
    --ostype l26 \
    --net0 "virtio,bridge=${BRIDGE}" \
    --agent 1 \
    --serial0 socket \
    --vga serial0
[[ $DRY_RUN -eq 0 ]] && PARTIAL_VM_CREATED=1

# Import disk (size configurable so templates land at the desired capacity)
if [[ $DRY_RUN -eq 1 ]]; then
    log "[dry-run] qm importdisk ${VMID} ${IMG_PATH} ${STORAGE} --format ${DISK_FORMAT}"
else
    qm importdisk "${VMID}" "${IMG_PATH}" "${STORAGE}" --format "${DISK_FORMAT}" > /dev/null
fi

# Attach disk — read the volume name Proxmox assigned rather than guessing it
if [[ $DRY_RUN -eq 1 ]]; then
    DISK_VOL="${STORAGE}:vm-${VMID}-disk-0"
    log "[dry-run] qm set ${VMID} --scsihw virtio-scsi-pci --virtio0 ${DISK_VOL},discard=on"
else
    DISK_VOL=$(qm config "${VMID}" | awk -F': ' '/^unused0:/{print $2}')
    [[ -z "$DISK_VOL" ]] && die "Could not find imported disk in VM ${VMID} config (unused0 missing)"
    qm set "${VMID}" --scsihw virtio-scsi-pci --virtio0 "${DISK_VOL},discard=on"
fi
run qm set "${VMID}" --boot order=virtio0
run qm set "${VMID}" --scsi1 "${STORAGE}:cloudinit"
run qm set "${VMID}" --cicustom "vendor=local:snippets/${TEMPLATE_NAME}-vendor.yaml"
run qm set "${VMID}" --ciupgrade 0

# Convert to template
run qm template "${VMID}"
PARTIAL_VM_CREATED=0

# Cleanup downloaded image
run rm -f "${IMG_PATH}"

echo ""
if [[ $DRY_RUN -eq 1 ]]; then
    echo "✓ Dry run complete — no changes were made"
else
    echo "✓ Setup complete"
    echo ""
    echo "  Template '${TEMPLATE_NAME}' (VMID ${VMID}) is ready."
    echo ""
    echo "  Set in your Devsy provider config:"
    echo "    PROXMOX_TEMPLATE=${TEMPLATE_NAME}"
fi
echo ""
