# Development Tasks

Estimates are engineering mandays. Tasks may run in parallel only when dependencies and interfaces are stable. Every task follows requirement analysis -> DDD design -> tests first -> implementation -> QA -> documentation.

| ID | Task | Priority | Dependencies | Estimate |
|---|---|---|---|---:|
| VUL-001 | Initialize repository, Go modules, lint/format | P0 | - | 1 |
| VUL-002 | Define module boundaries and architecture tests | P0 | VUL-001 | 1 |
| VUL-003 | PostgreSQL config and migration framework | P0 | VUL-001 | 1 |
| VUL-004 | CI pipeline: tests, lint, dependency/secret scan | P0 | VUL-001 | 2 |
| VUL-005 | Next.js app shell and typed API client | P0 | - | 2 |
| VUL-006 | Identity domain and repository ports | P0 | VUL-002 | 2 |
| VUL-007 | Registration/login/session lifecycle | P0 | VUL-006 | 3 |
| VUL-008 | Workspace aggregate and membership | P0 | VUL-006 | 2 |
| VUL-009 | RBAC and resource authorization | P0 | VUL-008 | 3 |
| VUL-010 | Project context and CRUD | P0 | VUL-009 | 2 |
| VUL-011 | Asset value objects and canonicalization | P0 | VUL-010 | 2 |
| VUL-012 | Asset CRUD and duplicate constraints | P0 | VUL-011 | 2 |
| VUL-013 | Authorization aggregate and evidence metadata | P0 | VUL-010 | 3 |
| VUL-014 | Scope matching and exclusion policy | P0 | VUL-011,VUL-013 | 3 |
| VUL-015 | Network destination policy and SSRF test suite | P0 | VUL-014 | 4 |
| VUL-016 | Scan aggregate and state machine | P0 | VUL-014 | 2 |
| VUL-017 | Transactional outbox/job queue | P0 | VUL-016 | 3 |
| VUL-018 | Worker lease, retry, recovery, idempotency | P0 | VUL-017 | 3 |
| VUL-019 | Scanner profile catalog and immutable versions | P0 | VUL-016 | 2 |
| VUL-020 | Isolated scanner executor and resource limits | P0 | VUL-018 | 4 |
| VUL-021 | Scanner adapter contract and conformance tests | P0 | VUL-019 | 2 |
| VUL-022 | Nmap constrained adapter | P0 | VUL-020,VUL-021 | 3 |
| VUL-023 | ZAP baseline adapter | P0 | VUL-020,VUL-021 | 3 |
| VUL-024 | Bounded output parser and normalization | P0 | VUL-022 | 3 |
| VUL-025 | Finding aggregate and fingerprint deduplication | P0 | VUL-024 | 3 |
| VUL-026 | Finding status and audit workflow | P0 | VUL-025 | 2 |
| VUL-027 | Explainable versioned risk scoring | P0 | VUL-025 | 2 |
| VUL-028 | HTML and JSON report generation | P0 | VUL-025 | 3 |
| VUL-029 | Audit trail and redaction | P0 | VUL-007 | 2 |
| VUL-030 | Frontend workspace/project/assets screens | P0 | VUL-005,VUL-010,VUL-012 | 4 |
| VUL-031 | Frontend authorization/scope screens | P0 | VUL-013,VUL-014 | 3 |
| VUL-032 | Frontend scan management and status | P0 | VUL-016,VUL-023 | 3 |
| VUL-033 | Frontend findings and report screens | P0 | VUL-025,VUL-028 | 3 |
| VUL-034 | Docker Compose and deployment docs | P0 | VUL-003,VUL-020 | 2 |
| VUL-035 | Integration, security, and recovery tests | P0 | Core modules | 5 |
| VUL-036 | Backup/restore verification and release docs | P0 | VUL-035 | 2 |
| VUL-037 | Curated Nuclei adapter | P1 | VUL-021,VUL-020 | 3 |
| VUL-038 | CVE/CWE/CVSS enrichment | P1 | VUL-025 | 3 |
| VUL-039 | Remediation assignment and due dates | P1 | VUL-026 | 3 |
| VUL-040 | PDF reporting | P1 | VUL-028 | 2 |
| VUL-041 | Adapter SDK documentation | P1 | VUL-021 | 2 |
| VUL-042 | Webhooks and CI integration | P1 | VUL-009,VUL-016 | 4 |
| VUL-043 | SaaS readiness review and tenant isolation hardening | P1 | VUL-035 | 5 |
| VUL-044 | External plugin runtime security design | P2 | VUL-041 | 5 |
| VUL-045 | AI finding explanation design/prototype | P2 | VUL-038 | 5 |

## Task template

Each Jira task must include:
- Requirement IDs and business rules.
- Bounded context and aggregate.
- Sequence diagram or interaction notes where useful.
- Acceptance criteria and negative/security tests.
- Manday estimate and dependencies.
- Files/modules expected to change.
- Definition of Done.
- Out-of-scope notes to prevent unrelated refactoring.

## Execution rule

Do not implement the entire backlog in one pass. Start with VUL-001 through VUL-004, review architecture and CI, then proceed by dependency order. Before each task, inspect existing code and preserve unrelated legacy behavior.
