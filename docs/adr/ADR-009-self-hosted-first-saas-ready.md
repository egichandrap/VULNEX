# ADR-009: Self-Hosted MVP, SaaS-Ready Design

- Status: Accepted
- Decision: Ship self-hosted Docker Compose first; maintain tenant-aware boundaries and deployment abstractions for future SaaS.

## Context
Both individual and enterprise users are targets, but operating a public scanning SaaS introduces substantial abuse, egress, tenant isolation, billing, and support requirements.

## Consequences
Do not launch public SaaS scanning in MVP. Build tenant isolation and quotas into design; defer billing, provisioning, operator console, and managed egress until a separate security and operational readiness review.
