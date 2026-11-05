# Functional Requirements

Priority: P0 = MVP blocker; P1 = next iteration; P2 = later.

## Identity and workspace
- FR-IAM-001 P0: User can register and authenticate securely.
- FR-IAM-002 P0: User can create a workspace.
- FR-IAM-003 P0: Workspace owner/admin can invite, remove, and assign members.
- FR-IAM-004 P0: API enforces workspace and resource-level authorization.
- FR-IAM-005 P1: Scoped API tokens support automation and revocation.
- FR-IAM-006 P0: Initial administrator onboarding is secured via an ephemeral one-time setup token (ADR-013).

## Project and asset
- FR-ASSET-001 P0: User can create, update, archive projects.
- FR-ASSET-002 P0: User can register domain, hostname, IP, URL, and API base URL assets.
- FR-ASSET-003 P0: System validates canonical form and prevents malformed targets.
- FR-ASSET-004 P0: System records asset ownership/context and tags.
- FR-ASSET-005 P1: Asset discovery history and deduplication.

## Authorization and scope
- FR-SCOPE-001 P0: User can record authorization evidence, approver, validity, allowed methods, and exclusions.
- FR-SCOPE-002 P0: System requires valid authorization before scan submission.
- FR-SCOPE-003 P0: System validates each target against explicit scope.
- FR-SCOPE-004 P0: Scope is revalidated by worker immediately before execution.
- FR-SCOPE-005 P0: DNS resolution and destination IP are checked against network policy at connection time; private networks are denied by default unless explicitly authorized at the workspace level (ADR-015); metadata and loopback are permanently blocked.
- FR-SCOPE-006 P0: Expired/revoked authorization prevents queued scan execution.
- FR-SCOPE-007 P1: Approval workflow with separate approver permission.

## Scan orchestration
- FR-SCAN-001 P0: User can select only published scanner profiles.
- FR-SCAN-002 P0: API creates asynchronous scan jobs.
- FR-SCAN-003 P0: Worker executes job outside API process.
- FR-SCAN-004 P0: Job supports queued, running, completed, failed, cancelled, timed-out states.
- FR-SCAN-005 P0: Worker enforces timeout, concurrency, CPU, memory, output-size, and wall-clock limits.
- FR-SCAN-006 P0: User can cancel a running scan.
- FR-SCAN-007 P0: Job retries are bounded and idempotent.
- FR-SCAN-008 P1: Recurring scan schedules.
- FR-SCAN-009 P1: Scan progress events and notifications.

## Scanner integrations
- FR-ADAPTER-001 P0: Nmap adapter supports constrained approved discovery profile.
- FR-ADAPTER-002 P0: ZAP adapter supports baseline/passive-oriented profile.
- FR-ADAPTER-003 P1: Nuclei adapter uses curated, reviewed safe templates only.
- FR-ADAPTER-004 P0: Adapter output is parsed as untrusted data and normalized.
- FR-ADAPTER-005 P0: Scan records capture tool version and profile version.
- FR-ADAPTER-006 P1: Adapter conformance test suite.
- FR-ADAPTER-007 P2: Third-party plugin runtime after sandbox/security design approval.

## Findings and risk
- FR-FIND-001 P0: System stores normalized findings with source, asset, scan, evidence reference, and timestamps.
- FR-FIND-002 P0: System deduplicates repeat observations using stable fingerprints.
- FR-FIND-003 P0: User can filter and search findings.
- FR-FIND-004 P0: User can change finding status with audit event.
- FR-FIND-005 P1: CVE/CWE/CVSS enrichment with source and retrieval timestamp.
- FR-RISK-001 P0: System calculates explainable risk score using versioned model.
- FR-RISK-002 P1: User can configure asset criticality.

## Remediation and reporting
- FR-REMED-001 P1: User can assign finding owner, due date, and remediation notes.
- FR-REMED-002 P1: User can mark remediation for retest.
- FR-REPORT-001 P0: User can generate HTML and JSON reports.
- FR-REPORT-002 P1: User can export PDF.
- FR-REPORT-003 P0: Report includes scope, authorization reference, scan profile, timestamps, limitations, and findings.
- FR-REPORT-004 P0: Report access is workspace-authorized and export is audited.

## Audit and platform
- FR-AUDIT-001 P0: System records security-sensitive actions in append-only audit records.
- FR-AUDIT-002 P0: Audit records identify actor, action, resource, time, and correlation ID.
- FR-OPS-001 P0: Application emits structured logs and health endpoints.
- FR-OPS-002 P1: Metrics and distributed tracing.
- FR-PLUGIN-001 P1: Stable adapter contract and built-in adapter registration.
- FR-PLUGIN-002 P2: Signed external plugin catalog and isolated runtime.
