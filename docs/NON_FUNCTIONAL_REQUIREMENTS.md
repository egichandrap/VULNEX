# Non-Functional Requirements

## Performance
- NFR-PERF-001: Ordinary CRUD endpoints target P95 < 300 ms under a documented representative load, excluding scanner execution and report rendering.
- NFR-PERF-002: Scan concurrency is configurable globally and per workspace.
- NFR-PERF-003: Worker resource ceilings are configurable.
- NFR-PERF-004: Large scanner output is streamed or stored in bounded object/file storage, not loaded without limits into memory.

## Availability and reliability
- NFR-REL-001: API and worker health are independently observable.
- NFR-REL-002: Jobs survive API restarts and can be recovered after worker interruption.
- NFR-REL-003: Job processing is at-least-once with idempotent state transitions and result persistence.
- NFR-REL-004: Retry count and retryable error classes are bounded.
- NFR-REL-005: Backup and restore procedures are documented and tested before 1.0.

## Security
- NFR-SEC-001: All external inputs are validated.
- NFR-SEC-002: Secrets are not stored in plaintext or logged.
- NFR-SEC-003: Authorization and scope checks are mandatory for every scan path.
- NFR-SEC-004: Scanner processes run with least privilege and isolated network/filesystem boundaries.
- NFR-SEC-005: Workspace and resource isolation is tested.
- NFR-SEC-006: Dependencies and container images are scanned in CI.
- NFR-SEC-007: Security-sensitive actions are audited.

## Maintainability
- NFR-MAIN-001: DDD and Clean Architecture are mandatory.
- NFR-MAIN-002: Domain layer has no framework, database, HTTP, or scanner SDK dependency.
- NFR-MAIN-003: Domain and application behavior is developed using TDD.
- NFR-MAIN-004: Public API is documented with OpenAPI.
- NFR-MAIN-005: Architecture decisions are recorded as ADRs.

## Scalability
- NFR-SCALE-001: API and workers can be deployed as separate processes.
- NFR-SCALE-002: Worker concurrency can be increased independently of API replicas.
- NFR-SCALE-003: Tenant isolation is enforced in repository/query boundaries and authorization policy.
- NFR-SCALE-004: SaaS-specific infrastructure can be added without changing core domain rules.

## Observability
- NFR-OBS-001: Structured logs include request/job correlation IDs.
- NFR-OBS-002: Metrics cover API latency, job queue depth, scan duration, worker failures, cancellations, and adapter errors.
- NFR-OBS-003: Traces cover request-to-job-to-adapter flow where supported.
- NFR-OBS-004: Logs redact credentials, tokens, sensitive evidence, and personal data.

## Usability and accessibility
- NFR-UX-001: Dashboard supports empty, loading, error, and permission-denied states.
- NFR-UX-002: Severity is not communicated by color alone.
- NFR-UX-003: Critical scan details and authorization are visible before confirmation.
- NFR-UX-004: Keyboard navigation and accessible contrast are supported.

## Portability
- NFR-PORT-001: Self-hosted deployment uses documented Docker Compose configuration.
- NFR-PORT-002: Configuration is environment-driven and secrets are externalized.
- NFR-PORT-003: Database schema changes use versioned migrations.
