# ADR-005: Integrate Existing Scanners Through Adapters

- Status: Accepted
- Decision: Integrate Nmap and OWASP ZAP for MVP 0.1, with curated Nuclei profiles scheduled for Release 0.2 (P1), behind an extensible scanner port.

## Context
Reimplementing mature scanner engines is outside product scope.

## Consequences
VULNEX owns authorization, execution controls, normalization, workflow, and reporting. Initial MVP 0.1 focuses on Nmap and OWASP ZAP; Nuclei is planned for Release 0.2. Tool licenses, versions, templates, and output compatibility require review and pinning. The scanner architecture must remain extensible to accommodate Nuclei without refactoring core scan orchestration.
