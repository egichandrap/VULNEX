# ADR-002: Go Backend

- Status: Accepted
- Decision: Use Go for API and worker.

## Context
VULNEX requires concurrent job processing, process supervision, and efficient resource use.

## Consequences
Positive: strong concurrency primitives, straightforward static binaries, good container support.
Negative: team must establish consistent error, validation, dependency injection, and testing conventions.
