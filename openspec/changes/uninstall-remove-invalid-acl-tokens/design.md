# Design

## Context

`scripts/uninstall.sh` Phase 3 removes the `devsy@pve` service account. Its
current order is:

1. ACLs referencing both `devsy@pve` and `devsy@pve!devsy`
2. Token `devsy@pve!devsy`
3. User `devsy@pve`

The ACL filter uses `awk` on `pveum acl list` text output, which alphabetically
sorts the property keys — `path` before `propagate`, not `path` before `type`.
This means the awk split into `$1`..`$4` is column-order dependent and
fragile across Proxmox versions. After the filter succeeds, the token-ACL
removal step invokes `pvesh set /access/acl ... --tokens ... --delete 1` for
each row. If any removal fails, `set -e` aborts the script, leaving surviving
token ACLs and orphaned tokens in `user.cfg`.

The warning itself comes from `parse_user_config` in Proxmox VE's
`PVE::AccessControl`: when it encounters an ACL entry pointing at a token that
is not present in the current user's `tokens` map it emits
`user config - ignore invalid acl token '$ug'` (confirmed at
`pve-access-control/src/PVE/AccessControl.pm:1551`). The `delete_user_acl`
subroutine at line 1020 explicitly only removes user bindings, never token
bindings, so the only reliable remediation path is to delete token ACLs before
the token disappears.

## Goals / Non-Goals

**Goals:**
- Make Phase 3 idempotent: re-running after partial teardown does not fail.
- Remove all token ACLs before removing the token so Proxmox never sees a
  dangling token reference in `user.cfg`.
- Tolerate manual or prior cleanup: each step is already-removed-tolerant.
- Make ACL parsing robust across Proxmox releases by using structured JSON
  output (`jq`) instead of text columns.

**Non-Goals:**
- Touch Proxmox itself.
- Delete or modify any DevPod-era resource.
- Change the uninstall phases outside of Phase 3.

## Decisions

1. **Switch Phase 3 ACL enumeration to JSON.** `pveum acl list
   --output-format json | jq -r '.[] | select(.ugid=="devsy@pve" or
   .ugid=="devsy@pve!devsy") | ...'` is stable across Proxmox releases.
2. **Use `pveum acl delete` (or `pvesh set /access/acl --delete 1`) for each
   token ACL entry, with `|| true`-style tolerance.** The current script already
   uses `pvesh set ... --delete 1` for ACL removal; keep that and append
   `|| true` to tolerate retries.
3. **Token removal is already tolerant.** `pveum user token list ... | grep -q`
   checks existence before `pveum user token remove`; keep this.
4. **User removal is already tolerant.** `pveum user del` after user-existence
   check; keep.
5. **Add a comment block explaining why ACL must come before token removal** so
   a future maintainer does not reorder the steps.

## Risks / Trade-offs

- **`jq` availability.** `/usr/bin/jq` ships with `proxmox-ve` package on all
  supported releases; no extra install is needed.
- **`pveum acl list --output-format json` output shape.** Per Proxmox source at
  `pve-access-control/src/PVE/API2/ACL.pm:48-96` the shape is
  `[{path, type, ugid, roleid, propagate}, ...]`. `jq` parsing is stable.
- **No-op when user doesn't exist.** Keep the outer `if pveum user list | grep
   -q ...` gate; the rest runs only when the account is present.

## Migration Plan

Replace the Phase 3 ACL iteration in `scripts/uninstall.sh`. Roll out is just a
file push: the new uninstall works on any system where `jq` and `pveum` exist;
older behavior was purely cosmetic (warnings) so the change is safe to deploy.

## Open Questions

- None.
