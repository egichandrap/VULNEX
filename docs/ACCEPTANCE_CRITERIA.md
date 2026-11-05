# Acceptance Criteria

Acceptance criteria are normative and must be automated where practical.

## Authentication and tenant isolation
- AC-001: Valid credentials authenticate; invalid credentials return a generic authentication error.
- AC-002: Every protected endpoint rejects unauthenticated requests.
- AC-003: A user cannot read, update, delete, or infer the existence of another workspace's resources.
- AC-004: Role and resource permission checks are enforced server-side.
- AC-004a: Initial administrator bootstrapping verifies an ephemeral setup token, records only its cryptographic hash, marks the token as used, and permanently disables the bootstrap endpoint once the first administrator is created (ADR-013).
- AC-004b: API validates trusted upstream reverse proxy CIDRs; untrusted forwarding headers are stripped/ignored (ADR-014).

## Assets and authorization
- AC-005: Invalid domain, IP, CIDR, URL, scheme, and port inputs are rejected.
- AC-006: Authorization must identify project, targets, permitted methods, approver, evidence reference, and validity.
- AC-007: Missing, expired, revoked, or insufficient authorization prevents scan submission.
- AC-008: Excluded targets override included targets.
- AC-009: Private destinations (RFC 1918, RFC 4193) are denied by default. An administrator may enable private network scanning only at the workspace level with documented authorization and audit logging. Cloud metadata (`169.254.169.254`), loopback (`127.0.0.0/8`, `::1`), link-local, and Docker internal networks remain permanently blocked regardless of workspace settings (ADR-015).
- AC-010: A queued job is rejected if authorization expires or is revoked before execution.

## Scope and network safety
- AC-011: Worker revalidates scope immediately before execution.
- AC-012: DNS results are validated at connection time; any prohibited result causes fail-closed behavior.
- AC-013: Redirect destinations are revalidated.
- AC-014: IPv4, IPv6, IPv4-mapped IPv6, and CIDR representations are canonicalized and tested.
- AC-015: No API accepts arbitrary shell commands or unrestricted scanner flags.

## Scan lifecycle
- AC-016: Scan request creates a durable job and returns a scan identifier.
- AC-017: Valid state transitions are enforced; terminal states cannot return to RUNNING.
- AC-018: Duplicate delivery does not create duplicate persisted findings or duplicate state transitions.
- AC-019: Timeout produces TIMED_OUT and terminates the scanner process.
- AC-020: Cancellation produces CANCELLED after worker confirmation/termination.
- AC-021: Worker crash/restart does not silently lose the job; recovery behavior is documented and tested.
- AC-022: Resource and output limits are enforced.

## Scanner and findings
- AC-023: Each adapter records tool version, profile version, exit status, and execution timestamps.
- AC-024: Malformed or oversized scanner output is rejected safely and does not crash API.
- AC-025: Findings retain source, scan, asset, evidence reference, and provenance.
- AC-026: Repeated observations with the same fingerprint are deduplicated according to documented rules.
- AC-027: Severity and risk are separate fields; risk calculation records model version and inputs.

## Reporting and audit
- AC-028: HTML/JSON report includes authorization reference, exact scope, profile, scan time, limitations, and findings.
- AC-029: Report export requires permission and creates an audit event.
- AC-030: Audit events include actor, action, resource, timestamp, and correlation ID.
- AC-031: Sensitive values are redacted from logs and default report output.

## Release
- AC-032: A clean environment can install the MVP using documented Docker Compose steps.
- AC-032a: Server mode deploys with embedded Caddy providing automated ACME TLS certificate provisioning; external proxy mode enforces trusted CIDR validation (ADR-014).
- AC-033: Database migrations can be applied from an empty database and upgraded from the previous release.
- AC-034: Backup/restore procedure is tested.
- AC-035: CI runs formatting, linting, unit tests, integration tests, dependency scanning, and container scanning across all monorepo modules (ADR-012).
