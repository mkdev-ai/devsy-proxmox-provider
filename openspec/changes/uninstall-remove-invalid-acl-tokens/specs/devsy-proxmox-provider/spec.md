# Spec Delta

## ADDED Requirements

### Requirement: Uninstall leaves no invalid token ACL warnings

`scripts/uninstall.sh` Phase 3 SHALL remove all Devsy ACL entries that reference the `devsy@pve!devsy` token before removing the token itself, and SHALL NOT leave any ACL entry that references a removed token.

#### Scenario: Complete teardown
- **WHEN** the administrator runs `scripts/uninstall.sh` against a Proxmox host where `devsy@pve`, `devsy@pve!devsy`, and all Devsy ACLs exist
- **THEN** the script removes all Devsy token ACLs, then the Devsy user ACLs, then the token, then the user
- **AND** Proxmox does not emit `user config - ignore invalid acl token 'devsy@pve!devsy'` after the script completes

#### Scenario: Re-run after partial teardown
- **WHEN** the administrator re-runs `scripts/uninstall.sh` after an earlier partial run that left orphaned ACLs
- **THEN** the script tolerates already-removed resources and completes without failing under `set -e`

### Requirement: Uninstall tolerates partial or repeated cleanup

`scripts/uninstall.sh` Phase 3 SHALL tolerate resources that were already removed by an earlier run or by manual intervention.

#### Scenario: Re-run after manual token removal
- **WHEN** the administrator manually removes the `devsy@pve!devsy` token before running `scripts/uninstall.sh`
- **THEN** the script skips the already-removed token and continues to remove ACLs and the user without failing

#### Scenario: Re-run after user deletion
- **WHEN** the administrator manually deletes the `devsy@pve` user before running `scripts/uninstall.sh`
- **THEN** the script detects the missing user, skips all Phase 3 steps, and reports "User devsy@pve not found — skipping"
