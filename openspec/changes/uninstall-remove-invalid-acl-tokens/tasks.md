# Tasks

- [x] 1.0 Update Phase 3 ACL enumeration in `scripts/uninstall.sh` to use `pveum acl list --output-format json | jq` instead of fragile `awk` text parsing
- [x] 1.1 Remove token ACLs (`devsy@pve!devsy`) BEFORE the token removal step, with `|| true` tolerance so re-runs do not abort under `set -e`
- [x] 1.2 Keep user ACL removal (`devsy@pve`) tolerant and idempotent
- [x] 1.3 Add a comment block in Phase 3 explaining why ACL-before-token order is required (orphaned token ACLs cause persistent `ignore invalid acl token` warnings)
- [x] 1.4 Run `bash -n scripts/uninstall.sh` and review the Phase 3 ordering; verify the script remains idempotent
