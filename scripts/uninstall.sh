#!/usr/bin/env bash
# uninstall.sh — Remove Devsy Proxmox resources created by setup.sh.
# Targets only Devsy-named resources; legacy DevPod resources are unsupported
# and are intentionally left untouched. No flags are required.
# Run as root (or with sudo) on the Proxmox host.
#
# Usage:
#   sudo bash uninstall.sh [--dry-run]
#
# Example (pipe over SSH):
#   ssh user@proxmox-host 'sudo bash -s' < uninstall.sh
set -euo pipefail

DRY_RUN=0

# ── helpers ───────────────────────────────────────────────────────────────────
die()  { echo "ERROR: $*" >&2; exit 1; }
log()  { echo "    $*"; }
step() { echo ""; echo "==> $*"; }

run() {
    if [[ $DRY_RUN -eq 1 ]]; then
        log "[dry-run] $*"
    else
        "$@"
    fi
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --dry-run) DRY_RUN=1; shift ;;
        --help|-h) grep '^#' "$0" | sed 's/^# \?//'; exit 0 ;;
        *) die "Unknown option: $1" ;;
    esac
done

# ── preflight ─────────────────────────────────────────────────────────────────
[[ "$EUID" -ne 0 ]] && die "Must be run as root. Try: sudo bash $0"
command -v pveum &>/dev/null || die "pveum not found — run this script on a Proxmox VE host"
command -v qm    &>/dev/null || die "qm not found — run this script on a Proxmox VE host"

[[ $DRY_RUN -eq 1 ]] && echo "" && log "Mode: dry-run (no changes will be made)"

DEVSY_USER="devsy@pve"
DEVSY_TOKEN_NAME="devsy"
DEVSY_TOKEN_ID="${DEVSY_USER}!${DEVSY_TOKEN_NAME}"

# ── phase 1: VMs ──────────────────────────────────────────────────────────────
step "Phase 1: VMs"

DEVSY_VMIDS=$(qm list 2>/dev/null | awk 'NR>1 && $2 ~ /^devsy-/ {print $1}')

if [[ -z "$DEVSY_VMIDS" ]]; then
    log "No Devsy VMs found"
else
    while IFS= read -r vmid; do
        name=$(qm config "$vmid" 2>/dev/null | awk -F': ' '/^name:/{print $2}')
        if [[ "$name" != devsy-* ]]; then
            log "Skipping VM ${vmid}: name '${name:-unknown}' is not Devsy-prefixed"
            continue
        fi
        status=$(qm status "$vmid" 2>/dev/null | awk '{print $2}')
        if [[ "$status" == "running" ]]; then
            run qm stop "$vmid"
            [[ $DRY_RUN -eq 0 ]] && log "Stopped VM ${vmid}"
        fi
        run qm destroy "$vmid" --purge
        [[ $DRY_RUN -eq 0 ]] && log "Destroyed VM ${vmid} (${name})"
    done <<< "$DEVSY_VMIDS"
fi

# ── phase 2: cloud-init snippets ──────────────────────────────────────────────
step "Phase 2: Cloud-init snippets"

DEVSY_SNIPPETS=$(find /var/lib/vz/snippets -name 'devsy-*-vendor.yaml' 2>/dev/null || true)

if [[ -z "$DEVSY_SNIPPETS" ]]; then
    log "No Devsy snippets found"
else
    while IFS= read -r f; do
        run rm -f "$f"
        [[ $DRY_RUN -eq 0 ]] && log "Removed ${f}"
    done <<< "$DEVSY_SNIPPETS"
fi

# ── phase 3: service account ──────────────────────────────────────────────────
step "Phase 3: Service account"

if ! pveum user list --output-format json 2>/dev/null | grep -q "\"userid\":\"${DEVSY_USER}\""; then
    log "User ${DEVSY_USER} not found — skipping"
else
    # Remove ACLs for the token BEFORE removing the token itself.
    # Proxmox VE warns "user config - ignore invalid acl token" whenever
    # /etc/pve/user.cfg contains an ACL entry that references a token that
    # no longer exists on any user. Proxmox's built-in user deletion path
    # (delete_user_acl) only removes the user bindings, not token bindings,
    # so we must delete the token ACLs explicitly while the token is still
    # present.
    #
    # Order (must not be changed without reading the above):
    #   1. Token ACLs
    #   2. User ACLs
    #   3. Token
    #   4. User

    # 1. Token ACLs — tolerate already-removed entries so re-runs are safe
    while IFS= read -r acl_json; do
        path=$(jq -r '.path' <<< "$acl_json")
        role=$(jq -r '.roleid' <<< "$acl_json")
        run pvesh set /access/acl --path "$path" --roles "$role" --tokens "${DEVSY_TOKEN_ID}" --delete 1 || true
        [[ $DRY_RUN -eq 0 ]] && log "Removed token ACL: ${role} on ${path} for ${DEVSY_TOKEN_ID}"
    done < <(pveum acl list --output-format json 2>/dev/null | jq -c '.[] | select(.type=="token" and .ugid=="'"${DEVSY_TOKEN_ID}"'")')

    # 2. User ACLs — tolerate already-removed entries so re-runs are safe
    while IFS= read -r acl_json; do
        path=$(jq -r '.path' <<< "$acl_json")
        role=$(jq -r '.roleid' <<< "$acl_json")
        run pvesh set /access/acl --path "$path" --roles "$role" --users  "${DEVSY_USER}" --delete 1 || true
        [[ $DRY_RUN -eq 0 ]] && log "Removed user ACL: ${role} on ${path} for ${DEVSY_USER}"
    done < <(pveum acl list --output-format json 2>/dev/null | jq -c '.[] | select(.type=="user" and .ugid=="'"${DEVSY_USER}"'")')

    # 3. Token — tolerate already-removed token
    if pveum user token list "${DEVSY_USER}" --output-format json 2>/dev/null | grep -q "\"tokenid\":\"${DEVSY_TOKEN_NAME}\""; then
        run pveum user token remove "${DEVSY_USER}" "${DEVSY_TOKEN_NAME}"
        [[ $DRY_RUN -eq 0 ]] && log "Removed token ${DEVSY_TOKEN_ID}"
    else
        log "Token ${DEVSY_TOKEN_ID} not found — skipping"
    fi

    # 4. User — tolerate already-removed user
    run pveum user del "${DEVSY_USER}" || true
    [[ $DRY_RUN -eq 0 ]] && log "Removed user ${DEVSY_USER}"
fi

echo ""
if [[ $DRY_RUN -eq 1 ]]; then
    echo "✓ Dry run complete — no changes were made"
else
    echo "✓ Uninstall complete"
fi
echo ""
