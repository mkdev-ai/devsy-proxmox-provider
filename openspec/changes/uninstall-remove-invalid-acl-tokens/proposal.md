# Proposal

## Why

`scripts/uninstall.sh` Phase 3 leaves orphaned ACL entries that reference the
`devsy@pve!devsy` token after the token is removed. On every subsequent
`pveum user list` (and other commands that parse `user.cfg`) Proxmox VE emits
`user config - ignore invalid acl token 'devsy@pve!devsy'` for each surviving
entry. The provider does not cause these warnings — the uninstall script does,
and they persist until someone manually cleans the ACL or recreates the user.

The root cause is confirmed in Proxmox VE source
(`pve-access-control/src/PVE/AccessControl.pm`, `delete_user_acl` at line 1020):
it only removes `acl_node->{users}->{$username}` and never touches
`acl_node->{tokens}->{...}`. So `pveum user token remove` followed by
`pveum user del` is not sufficient — token ACLs are intentionally not tied to
user lifecycle in Proxmox.

## What Changes

- Uninstall Phase 3 removes **token ACLs before the token**, so no orphaned ACL
  entry ever exists while the token is gone.
- ACL removal uses `pveum acl list --output-format json` parsed with `jq`
  instead of fragile text `awk`, because the Proxmox CLI text formatter sorts
  property keys alphabetically (`path` before `propagate`, not `path` before
  `type`). `jq` is already available in `/usr/bin/jq` on Proxmox VE 8.
- Every removal step tolerates "already removed" (exit 0 or known failure) so
  that partial runs, re-runs, and interactive manual cleanup do not fail under
  `set -e`.
- Script help and comments document the required order and why it exists.

## Capabilities

### Modified Capabilities

- `devsy-proxmox-provider`: uninstall script Phase 3 guarantees no token-ACL
  warnings remain after complete teardown.

## Impact

`scripts/uninstall.sh` only. No provider binary, manifest, or documentation
changes are required for this fix; the existing README description of uninstall
behavior remains correct.
