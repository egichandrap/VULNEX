# Product Roadmap

## Guiding strategy

Deliver a trustworthy end-to-end authorized assessment workflow before expanding scanner breadth. Self-hosted MVP first; SaaS-ready architecture, but managed SaaS operations are a later release.

| Phase | Scope | Exit criteria | Estimate |
|---|---|---|---:|
| 1. Foundation | Monorepo setup, Go/Next.js, DB, auth, admin setup token, RBAC, CI | Secure login, admin bootstrap, workspace isolation, quality gates | 8–12 MD |
| 2. Project & assets | Workspace/project/asset CRUD, authorization evidence, scope model | Cannot create valid scan without registered asset and authorization | 10–15 MD |
| 3. Scan engine | Job lifecycle, queue, worker isolation, destination policy, Caddy ingress, Nmap and ZAP baseline adapters | End-to-end safe scan with cancellation, SSRF protection, and audit | 15–20 MD |
| 4. Assessment | Normalization, Nuclei curated profile (0.2), CVE/CWE mapping (0.2) | Stable normalized results with provenance | 12–18 MD |
| 5. Findings & risk | Deduplication, triage, risk model, remediation | Findings lifecycle and explainable priority | 8–12 MD |
| 6. Reporting | HTML/JSON reports (PDF in 0.2) | Reproducible report with scope and limitations | 6–10 MD |
| 7. Extensibility | Adapter SDK contract, plugin metadata, compatibility | Built-in adapter conformance and safe extension design | 10–15 MD |
| 8. Hardening/release | Threat review, backup/restore, docs, release pipeline | Release checklist passed | 10–15 MD |

Estimated total: 79–117 engineering mandays. Calendar duration depends on team size, review, and integration complexity.

## Release sequence

### MVP 0.1
- Unified Monorepo with standalone Go CLI installer (`vulnex`).
- Self-hosted Docker Compose (Local loopback and Server modes).
- Embedded Caddy reverse proxy with automatic ACME TLS for Server mode (trusted CIDR validation for external proxies).
- Ephemeral one-time setup token for initial administrator bootstrap and web onboarding.
- Authentication, workspace, project, assets.
- Authorization and scope controls (private network scanning disabled by default, opt-in per workspace by admin with written authorization).
- Nmap constrained discovery adapter.
- ZAP baseline adapter.
- Scan jobs, leases, cancellation, and timeout.
- Findings normalization and triage.
- Basic HTML/JSON report.
- Audit trail.

### 0.2
- Nuclei curated safe template integration.
- PDF reporting.
- CVE/CWE enrichment.
- Remediation tracking.
- Improved dashboard and filters.

### 0.3
- Stable scanner adapter SDK.
- Webhooks and CI integration.
- Additional safe checks and reporting improvements.

### 1.0
- Security review and hardening.
- Documented upgrade/backup/restore.
- Stable API and migration policy.
- Community contribution process.
- SaaS deployment readiness review.

### Post-MVP
- Managed SaaS billing/subscriptions and tenant provisioning.
- AI finding explanation and remediation assistance with human review.
- Advanced API security testing.
- Continuous monitoring and scheduling.
- External plugin marketplace after security model is proven.

## Prioritization rules

P0: necessary for safe end-to-end MVP.
P1: improves usefulness after the core workflow is stable.
P2: expansion or advanced capability.
Do not promote a feature to P0 without a documented user need, security impact, and acceptance criteria.
