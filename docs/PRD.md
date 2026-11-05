# VULNEX — Product Requirements Document

**Document status:** Product baseline / implementation specification  
**Product type:** Open-source defensive security assessment platform  
**License target:** Apache License 2.0  
**MVP deployment:** Self-hosted Local and Server through one unified CLI installer  
**Backend:** Go | **Frontend:** Next.js + TypeScript | **Runtime:** Docker Engine + Docker Compose

## 1. Executive summary

VULNEX is an open-source platform for managing authorized vulnerability assessments. It provides a unified workflow to register projects and assets, record authorization and scope, orchestrate approved scanner integrations, normalize findings, manage remediation, and produce reports.

VULNEX must be installable without requiring users to clone the source repository or install development toolchains. Users download a signed platform-specific CLI artifact, run `vulnex install`, choose Local or Server mode, and let the CLI provision and manage the application runtime.

## 2. Product vision

Make controlled security assessment accessible to individuals and organizations while retaining ownership of deployment, data, and assessment scope.

### Product principles

1. Authorization before assessment.
2. Explicit and enforceable asset scope.
3. Safe defaults; no destructive behavior by default.
4. Self-hosted first; hosted cloud may be introduced later.
5. One application release for Local and Server modes.
6. Installer is a separate control-plane application, not part of backend business logic.
7. DDD, Clean Architecture, SOLID, and TDD are mandatory.
8. Modular monolith for MVP; split services only with evidence-based justification.
9. Scanner integrations use constrained adapters and isolated workers.
10. Transparent operations, audit logs, and recoverable upgrades.

## 3. Goals and success measures

| Goal | Success measure |
|---|---|
| Simple installation | A supported user can install without Git, Go, Node.js, or manual database setup |
| Unified deployment | Local and Server use the same release images and schema |
| Safe assessment | Every scan is linked to authorization and a validated scope |
| Reliable operations | Health checks, logs, backup, restore, and upgrade are available |
| Extensible scanning | Scanner adapters can be added without changing core domain rules |
| Open-source readiness | Reproducible releases, license notices, SBOM, checksums, signatures, and documentation |

## 4. Target users

- Individual security practitioners assessing their own systems.
- Developers and platform engineers validating applications they own.
- Internal security teams.
- Consultants conducting assessments under written authorization.
- Organizations that require self-hosted deployment and control over assessment data.

## 5. User problems

- Security tools are fragmented and results are difficult to consolidate.
- Teams lack a consistent asset scope and authorization record.
- Findings are duplicated, inconsistently prioritized, and hard to track.
- Self-hosted security tooling can require substantial manual setup.
- Scanner execution needs isolation, resource limits, and auditability.

## 6. Product scope

### 6.1 MVP capabilities

1. Authentication and role-based access.
2. Organization/workspace and project management.
3. Asset inventory and scope allowlists.
4. Authorization record and assessment approval.
5. Scan profile management with safe presets.
6. Scanner orchestration through approved adapters.
7. Initial integrations: Nmap for authorized discovery and OWASP ZAP baseline/passive-safe workflows for MVP 0.1; extensible adapter architecture for curated Nuclei templates in Release 0.2 (P1).
8. Finding normalization, deduplication, severity, status, and remediation notes.
9. Assessment history and audit trail.
10. Reports export (HTML and JSON).
11. Unified VULNEX CLI installer.
12. Local and Server deployment modes using managed Docker Compose runtime.
13. CLI lifecycle, diagnostics, logs, backup, restore, update, and uninstall.
14. Secure initial administrator onboarding via ephemeral one-time setup token.
15. Documentation for installation, configuration, testing, backup, and troubleshooting.

### 6.2 Out of scope for MVP

- Stealth or evasion features.
- Persistence, backdoors, credential theft, or password spraying.
- Automatic exploitation or destructive testing.
- Arbitrary shell command execution from the UI or API.
- Public multi-tenant SaaS billing.
- AI-based vulnerability analysis (post-MVP).
- Kubernetes-native installer as the primary deployment path.
- Air-gapped/offline bundle as a release blocker (planned extension).
- Native non-container deployment of every dependency.

## 7. Deployment and distribution requirements

### 7.1 Final installer decisions

| Decision | Final value |
|---|---|
| Repository topology | Unified Monorepo (ADR-012) |
| Installer model | Unified |
| Deployment modes | Local and Server |
| Server Ingress / TLS | Embedded Caddy (automatic ACME) default; external proxy supported with trusted CIDRs (ADR-014) |
| Runtime | Managed Docker Engine / Docker Compose integration |
| Interface | CLI only |
| Source checkout required | No (for release consumers) |
| Installer implementation | Go CLI |
| Application packaging | Versioned OCI images |
| Distribution | Signed platform-specific CLI artifacts and packages |
| First production target | Linux x86_64 server |
| Later platform targets | Linux ARM64, macOS Intel/Apple Silicon, Windows x86_64 |

### 7.2 Installation experience

The user downloads a release artifact, verifies it, installs the CLI, runs `vulnex doctor`, then executes one of:

```bash
vulnex install --mode local
vulnex install --mode server --domain security.example.com
```

The CLI checks prerequisites, validates configuration, creates secure secrets and persistent directories, verifies image provenance, deploys the application, waits for health checks, and prints the dashboard URL.

The installer may guide the administrator through installing Docker where supported, but must not silently elevate privileges or make system changes without consent. Docker Engine is not bundled inside the CLI binary.

### 7.3 Local mode

- Dashboard bound to loopback by default.
- No public ingress by default.
- Persistent application data under an OS-appropriate application data directory.
- Suitable for single-user evaluation and personal use.
- Authentication remains enabled.

### 7.4 Server mode

- Explicit domain or network exposure configuration.
- HTTPS required for remote access; Embedded Caddy reverse proxy provides automatic ACME TLS provisioning by default (ADR-014).
- Alternative external reverse proxy supported with strict trusted upstream IP/CIDR validation to prevent header spoofing.
- Database, Redis, and worker ports remain private.
- Persistent volumes and backup location explicitly configured.
- Initial administrator provisioning via ephemeral, cryptographically secure one-time setup token displayed once in terminal and completed via web onboarding (ADR-013).

## 8. Functional requirements

| ID | Requirement | Priority |
|---|---|---|
| FR-001 | Users can authenticate and manage sessions | Must |
| FR-002 | Users can create organizations/workspaces and projects | Must |
| FR-003 | Users can register assets and asset types | Must |
| FR-004 | Users can define allowlisted domains, IPs, CIDRs, and URLs | Must |
| FR-005 | A scan requires an active authorization record | Must |
| FR-006 | Scan targets are validated against the approved scope before execution | Must |
| FR-007 | Users can select safe scanner profiles | Must |
| FR-008 | Scanner execution is isolated and resource-limited | Must |
| FR-009 | Scanner outputs are validated and normalized | Must |
| FR-010 | Findings can be triaged, assigned, and tracked | Must |
| FR-011 | Users can export assessment reports | Must |
| FR-012 | All sensitive actions are audited | Must |
| FR-013 | CLI supports Local and Server installation modes | Must |
| FR-014 | CLI detects prerequisites and reports actionable diagnostics | Must |
| FR-015 | CLI manages application lifecycle | Must |
| FR-016 | CLI supports backup and restore | Must |
| FR-017 | CLI supports verified application upgrades | Must |
| FR-018 | CLI supports safe uninstall with explicit data-retention choice | Must |
| FR-019 | CLI supports configuration validation without exposing secrets | Must |
| FR-020 | Future hosted deployment can reuse application domain logic | Should |

## 9. Non-functional requirements

- Security: least privilege, secure defaults, strict scope enforcement, secret protection, image verification.
- Reliability: health checks, graceful shutdown, restart policy, backup/restore procedure.
- Performance: asynchronous scan execution; configurable concurrency and resource ceilings.
- Portability: platform-specific CLI binaries; application images shared across Local and Server.
- Maintainability: DDD, Clean Architecture, SOLID, TDD; modular monolith.
- Observability: structured logs, metrics, health/readiness endpoints, audit events.
- Usability: clear CLI help, actionable errors, non-interactive mode for automation.
- Accessibility: web dashboard should follow WCAG 2.1 AA where practical.
- Privacy: assessment data stays in the user's infrastructure for self-hosted deployments.

## 10. Security and safety requirements

1. Require explicit authorization before assessment.
2. Require allowlisted scope and reject out-of-scope targets.
3. Validate DNS resolution and resolved IP addresses to mitigate SSRF and DNS rebinding.
4. Block loopback, link-local, cloud metadata, and private network ranges by default. Private/internal scanning may only be explicitly enabled at the workspace level by an administrator with documented authorization and audit logging (ADR-015). Cloud metadata (`169.254.169.254`), loopback (`127.0.0.0/8`, `::1`), link-local, and internal Docker networks remain permanently prohibited under all circumstances.
5. Do not expose scanner control interfaces publicly.
6. Do not accept arbitrary scanner command strings from users.
7. Use curated scanner templates and constrained adapter options.
8. Enforce scan timeout, CPU, memory, process, and concurrency limits.
9. Support emergency cancellation.
10. Log actor, project, authorization reference, target scope, scanner profile, start/end, and outcome.
11. Sanitize and size-limit scanner output.
12. Use non-root containers where supported and drop unnecessary capabilities.
13. Keep scanner egress policy separate from management-plane networking.
14. Do not enable destructive or exploitative tests by default.
15. Require explicit approval for any future active test that can alter target state.

## 11. Architecture requirements

### Application architecture

- Repository structure: Unified Monorepo (`backend/`, `frontend/`, `installer/`, `deploy/`, `docs/`, `tests/`) (ADR-012).
- Backend: Go modular monolith (`cmd/api`, `cmd/worker`).
- Frontend: Next.js + TypeScript + Tailwind CSS.
- Database: PostgreSQL (transactional storage and MVP job queue via `SKIP LOCKED`).
- Ingress (Server mode): Embedded Caddy with automatic ACME TLS; external reverse proxy supported with trusted CIDR enforcement (ADR-014).
- Queue/cache: Redis where asynchronous workload requires it (post-MVP).
- Scanner execution: isolated worker containers.
- API: REST with OpenAPI specification.
- Observability: structured logs and health endpoints; OpenTelemetry/Prometheus as justified.
- Deployment: Docker Compose managed by VULNEX CLI.

### Installer architecture

The CLI is a separate Go application with its own bounded contexts:
- Environment
- Installation
- Runtime
- Configuration
- Deployment
- Lifecycle
- Maintenance
- Diagnostics

Installer ports abstract Docker, filesystem, process execution, registry, and OS package manager. Domain and application layers must not import Docker SDK, shell execution, or platform-specific packages directly.

## 12. Data and domain concepts

Core domains:
- Identity & Access
- Organization / Workspace
- Project
- Asset & Scope
- Authorization
- Assessment / Scan
- Scanner Integration
- Finding
- Remediation
- Reporting
- Audit

Installer domain:
- Installation
- DeploymentProfile
- RuntimeEnvironment
- ApplicationRelease
- Component
- PersistentData
- Backup
- UpgradePlan

## 13. Acceptance criteria

- A user can install from a release artifact without cloning source.
- Local and Server modes use the same application image versions.
- Installer detects missing prerequisites and explains remediation.
- Installer does not silently install system packages or escalate privileges.
- Install is idempotent and preserves existing persistent data.
- Health checks must pass before installation is reported successful.
- Upgrade verifies artifact provenance and schema compatibility.
- Backup can be restored in a documented test.
- Uninstall requires an explicit data-retention decision.
- No scan can start without valid authorization and scope.
- Out-of-scope targets are rejected before scanner invocation.
- Scanner worker is isolated and resource-limited.
- CI runs unit, integration, security, and end-to-end tests.

## 14. Release strategy

Unified release synchronization between installer CLI and application platform:
- 0.1 (MVP): Monorepo setup; Linux x86_64 CLI installer (`doctor`, `install --mode local|server`, lifecycle); Embedded Caddy TLS & setup token bootstrap; Authentication, Workspaces, Projects, Assets; Scope & Authorization engine (private scanning disabled by default, opt-in at workspace level); Nmap and OWASP ZAP adapters; Normalized findings & basic triage; HTML and JSON reports; Audit logs.
- 0.2: Nuclei curated safe template integration (P1); PDF reports (P1); CVE/CWE enrichment; Remediation assignment and due dates.
- 0.3: Stable Scanner Adapter SDK; Webhooks & CI integration; Recurring scan scheduling.
- 0.4: macOS and Windows CLI release packaging; signed distribution.
- 1.0: Production hardening, comprehensive security audit, and open-source release.

## 15. Risks and mitigations

| Risk | Mitigation |
|---|---|
| Docker unavailable or incompatible | `doctor`, compatibility matrix, guided setup |
| Image registry unavailable | clear retry/error behavior; later offline bundle |
| Upgrade damages data | backup, migration compatibility, recovery procedure |
| Scanner escapes intended scope | centralized scope validation, isolated worker, egress controls |
| Supply-chain compromise | signed artifacts, SBOM, pinned image digests |
| Platform differences | staged platform support and automated E2E matrix |
| Large scanner resource use | quotas, concurrency limits, timeout, cancellation |

## 16. Technical decisions status

### 16.1 Confirmed & Approved Decisions
- Repository Topology: Unified Monorepo with independent module boundaries and release artifacts (ADR-012).
- Scanner Scope for MVP 0.1: Nmap and OWASP ZAP in MVP 0.1; Nuclei deferred to Release 0.2 (P1) (ADR-005).
- Initial Admin Bootstrap: Ephemeral single-use setup token displayed once in terminal, web onboarding, hashed storage (ADR-013).
- Server Ingress & TLS: Embedded Caddy with automatic ACME default; external proxy supported with trusted CIDR check (ADR-014).
- Private Network Testing Policy: Disabled by default; explicit workspace-level administrator enablement with written authorization; metadata and loopback permanently blocked (ADR-015).

### 16.2 Remaining Open Decisions
The following implementation details are to be confirmed prior to final release packaging:
- Official organization and container registry namespace (e.g. `ghcr.io/vulnex` vs `docker.io/vulnex`).
- Minimum supported Docker Engine (e.g., 24.0+) and Compose (v2.20+) versions.
- Pinned scanner container base image tags and digest verification strategy.
- Backup encryption archive cipher details (e.g., age vs AES-256-GCM).
- Supported database upgrade/rollback policy and toolchain.
