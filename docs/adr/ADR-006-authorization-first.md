# ADR-006: Authorization-First Scanning

- Status: Accepted
- Decision: No scan may execute without valid explicit authorization, scope match, and execution-time destination validation. Private/internal networks are denied by default, with opt-in workspace-level administrator enablement subject to strict non-bypassable guardrails (ADR-015).

## Context
The platform can initiate network activity and must prevent unauthorized use.

## Consequences
Fail closed on missing/ambiguous authorization or target policy. Enforce policy in application and network egress layers. Private/internal networks remain denied by default unless explicitly authorized at the workspace level, while metadata and loopback destinations remain permanently prohibited.
