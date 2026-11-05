# Installer Development Tasks

Estimates are engineering mandays (MD), not calendar duration. Dependencies are task IDs. Linux amd64 is the first production target; macOS/Windows packaging is staged after the core deployment flow is stable.

| ID | Task | Priority | Dependencies | MD |
|---|---|---|---|---:|
| INS-001 | Initialize Go CLI with Cobra and command root | Critical | - | 1 |
| INS-002 | Define CLI error model, output format, exit codes | High | INS-001 | 1 |
| INS-003 | Implement OS/architecture/resource detection | Critical | INS-001 | 2 |
| INS-004 | Implement `doctor` diagnostics | High | INS-003 | 1 |
| INS-005 | Define installer domain model and ports | Critical | INS-001 | 2 |
| INS-006 | Setup TDD, lint, CI baseline | High | INS-001 | 1 |
| INS-007 | Implement Docker runtime adapter | Critical | INS-005 | 2 |
| INS-008 | Compose v2 compatibility checker | Critical | INS-007 | 1 |
| INS-009 | Runtime prerequisite validation | Critical | INS-003, INS-007 | 1 |
| INS-010 | Config loader and schema validation | Critical | INS-005 | 2 |
| INS-011 | Secret generation and secure storage | Critical | INS-010 | 2 |
| INS-012 | Local/Server deployment profile | High | INS-010 | 2 |
| INS-013 | Create secure Compose specifications | Critical | - | 3 |
| INS-014 | Release manifest and image resolver | Critical | INS-007 | 2 |
| INS-015 | Artifact signature/digest verification | Critical | INS-014 | 2 |
| INS-016 | Implement InstallUseCase orchestration | Critical | INS-012, INS-013, INS-015 | 3 |
| INS-017 | Persistent volume and directory manager | Critical | INS-013 | 2 |
| INS-018 | Health/readiness manager | Critical | INS-016 | 2 |
| INS-019 | Idempotency, lock, and recovery state | High | INS-016 | 2 |
| INS-020 | Start/stop/restart commands | Critical | INS-016 | 2 |
| INS-021 | Status and logs commands | High | INS-020 | 1 |
| INS-022 | Dashboard URL command | Medium | INS-021 | 1 |
| INS-023 | Backup and restore | Critical | INS-017 | 3 |
| INS-024 | Upgrade and migration orchestration | Critical | INS-014, INS-023 | 4 |
| INS-025 | Safe uninstall and data-retention flow | High | INS-020, INS-023 | 2 |
| INS-026 | GoReleaser setup | High | INS-001 | 1 |
| INS-027 | Linux binary/package build | Critical | INS-026 | 2 |
| INS-028 | macOS and Windows builds | High | INS-026 | 3 |
| INS-029 | Container image release pipeline | Critical | INS-013 | 2 |
| INS-030 | Checksums, signing, SBOM | Critical | INS-026, INS-029 | 2 |
| INS-031 | Release compatibility matrix | High | INS-029 | 2 |
| INS-032 | Unit tests for domain/use cases | Critical | INS-016, INS-024 | 2 |
| INS-033 | Docker adapter integration tests | Critical | INS-007 | 2 |
| INS-034 | Linux installation E2E | Critical | INS-016, INS-020 | 2 |
| INS-035 | Upgrade/backup/restore/uninstall E2E | Critical | INS-023, INS-024, INS-025 | 3 |
| INS-036 | Installer and scanner isolation security review | Critical | INS-013, INS-016 | 3 |
| INS-037 | Installation/testing/troubleshooting docs | High | INS-027, INS-034 | 2 |
| INS-038 | Release validation | Critical | INS-030, INS-035, INS-036 | 2 |

## Estimate

| Phase | MD |
|---|---:|
| Foundation | 8 |
| Runtime & configuration | 10 |
| Installation & deployment | 16 |
| Lifecycle | 13 |
| Distribution & release | 12 |
| Testing & hardening | 16 |
| **Total** | **75** |

## Definition of Done

- Domain and application logic have tests.
- Use cases depend on interfaces, not infrastructure implementations.
- No unrelated refactor.
- Documentation and CLI help are updated.
- Security acceptance criteria are tested.
- Logs do not contain secrets.
- All relevant CI checks pass.
