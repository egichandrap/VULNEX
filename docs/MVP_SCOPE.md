# MVP Scope — Balanced

## MVP objective

Deliver one complete, safe, useful workflow: create workspace/project, register assets, record authorization, define scope, run an approved scanner profile, review normalized findings, and export a report.

## Included (P0)

### Platform
- Unified Monorepo containing backend, frontend, installer CLI, deployment manifests, and tests.
- Go API and worker modular monolith.
- Next.js + TypeScript UI.
- PostgreSQL and versioned migrations.
- Docker Compose self-hosted deployment (Local and Server profiles).
- Embedded Caddy reverse proxy with automatic ACME TLS for Server mode (optional external proxy with trusted CIDR enforcement).
- Initial administrator provisioning via ephemeral single-use setup token and web onboarding.
- Authentication and workspace RBAC.
- Project CRUD.

### Asset and authorization
- Domain, hostname, IP, URL asset types.
- Asset canonicalization and duplicate prevention.
- Authorization evidence metadata, validity, permitted methods.
- Include/exclude scope rules (exclusions strictly override inclusions).
- Execution-time scope and destination validation.
- Private/internal destinations denied by default; explicit workspace-level administrator enablement permitted with written authorization and audit logging; cloud metadata and loopback permanently blocked.

### Scanning
- Durable asynchronous job lifecycle.
- Worker isolation and resource limits.
- Nmap constrained discovery profile (MVP 0.1).
- OWASP ZAP baseline/passive-oriented profile (MVP 0.1).
- Extensible adapter architecture for future scanner engines (Nuclei scheduled for Release 0.2 / P1).
- Scan cancellation and timeout.
- Audit events.
- Bounded scanner output parsing.

### Findings and reports
- Normalized findings.
- Basic deduplication.
- Severity and confidence.
- Open, in-progress, resolved, accepted-risk, false-positive states.
- Basic explainable risk score.
- HTML and JSON report.
- Audit trail.

### Engineering
- Unit/integration/security tests.
- OpenAPI baseline.
- CI quality gates.
- README, CONTRIBUTING, SECURITY, license.
- Backup/restore documentation.

## Explicitly excluded from MVP
- Nuclei scanner integration (scheduled for Release 0.2 / P1).
- PDF report generation (scheduled for Release 0.2 / P1).
- CVE/CWE external vulnerability intelligence enrichment (scheduled for Release 0.2 / P1).
- Remediation assignment with due dates (scheduled for Release 0.2 / P1).
- Managed SaaS billing and subscription.
- Public SaaS scanning until dedicated abuse/egress review.
- AI assistant.
- Arbitrary plugin installation or marketplace.
- Arbitrary Nuclei template upload.
- Destructive testing and automatic exploitation.
- Credential attack workflows.
- Advanced API fuzzing.
- Continuous monitoring and recurring scans.
- Enterprise SSO/SCIM.
- Multi-region deployment.

## MVP release gates
1. No scan can bypass authorization.
2. No scan can escape approved scope through redirects, DNS resolution, or address encoding.
3. Worker isolation and resource limits verified.
4. Cancellation and timeout verified.
5. Cross-workspace access tests pass.
6. Scanner output parser handles malformed/oversized input safely.
7. Report access is authorized and audited.
8. Docker Compose clean-install and upgrade paths pass.
9. Security review has no unresolved critical/high issue affecting scan safety or tenant isolation.
10. Documentation explains deployment, authorization, scope, limitations, and safe operation.
