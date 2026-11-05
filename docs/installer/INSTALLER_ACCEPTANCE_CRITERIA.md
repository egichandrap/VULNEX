# Installer Acceptance Criteria

| ID | Acceptance criteria | Verification |
|---|---|---|
| AC-INS-001 | Install from release artifact without Git/source checkout | Clean-machine E2E |
| AC-INS-002 | Detect supported OS and architecture | Unit + matrix test |
| AC-INS-003 | Detect Docker Engine and Compose capability | Adapter integration |
| AC-INS-004 | Report actionable prerequisite errors | CLI tests |
| AC-INS-005 | Support Local and Server profiles | E2E |
| AC-INS-006 | Local dashboard binds loopback by default | Network test |
| AC-INS-007 | Server mode requires safe remote-access configuration | Config tests |
| AC-INS-008 | Database, Redis, worker ports are not publicly exposed | Compose policy test |
| AC-INS-009 | Verify image manifest and digest before deployment | Signature/digest test |
| AC-INS-010 | Install is idempotent | Repeated-install E2E |
| AC-INS-011 | Persistent data survives restart and upgrade | E2E |
| AC-INS-012 | Mutating CLI operations are serialized | Concurrency test |
| AC-INS-013 | Health checks gate successful install result | Failure-injection E2E |
| AC-INS-014 | Backup integrity is verified | Backup/restore test |
| AC-INS-015 | Restore rejects incompatible/corrupt backup | Negative test |
| AC-INS-016 | Upgrade checks schema compatibility | Migration test |
| AC-INS-017 | Uninstall requires explicit retention decision | CLI tests |
| AC-INS-018 | Uninstall never prunes unrelated Docker resources | Integration test |
| AC-INS-019 | Secrets are never emitted in logs or config output | Secret-leak test |
| AC-INS-020 | Scanner workers have no Docker socket and limited resources | Compose/security test |
| AC-INS-021 | Installer cancellation leaves recoverable state | Failure-injection test |
| AC-INS-022 | Published artifacts have checksum, signature, and SBOM | Release pipeline check |
