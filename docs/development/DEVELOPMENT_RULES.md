# VULNEX Project Development Rules

These rules apply to every development task, every agent, and every implementation session. They are mandatory and non-negotiable.

## 1. Always read the relevant project context before starting a task
Before proposing or writing code, review the corresponding domain design, architectural specifications, database schema, and task definitions.

## 2. Domain-Driven Design (DDD) is mandatory
Model the software around business concepts and invariants. Maintain ubiquitous language across code, tests, and documentation. Respect bounded context boundaries.

## 3. Clean Architecture is mandatory
Maintain strict dependency direction: `Delivery → Application → Domain ← Infrastructure`. High-level policy must never depend on low-level technical details.

## 4. SOLID principles are mandatory
- **Single Responsibility:** Each aggregate, service, and use case must have one clear reason to change.
- **Open/Closed:** Extend functionality (e.g., new scanner adapters) without modifying existing domain rules.
- **Liskov Substitution:** Port implementations must honor the complete behavioral contract of their interfaces.
- **Interface Segregation:** Ports must remain small, focused, and client-centric.
- **Dependency Inversion:** Depend on abstractions (interfaces), never on concrete implementations.

## 5. Test-Driven Development (TDD) is mandatory
Follow the Red → Green → Refactor cycle for all domain and application code:
1. **Red:** Write a failing test verifying a domain invariant, value object validation, or use case workflow.
2. **Green:** Implement the minimal code necessary to make the test pass.
3. **Refactor:** Clean and optimize the implementation while maintaining passing tests.

## 6. Never bypass domain rules for implementation convenience
All business invariants must be enforced within domain aggregates, entities, and value objects. Never move domain logic into HTTP handlers, database queries, or UI controllers.

## 7. Never introduce infrastructure dependencies into the domain layer
The domain package must contain zero third-party framework, database driver (`database/sql`, `pgx`), HTTP (`net/http`, routers), or CLI dependencies. The domain must remain pure Go standard library (excluding transport/storage packages).

## 8. Never refactor unrelated code
Keep every change scoped strictly to the task at hand. Avoid opportunistic refactoring of working, tested legacy code.

## 9. Prefer enhancement-only changes over unnecessary rewrites
Build upon validated structures. Add new capabilities via ports and adapters rather than rewriting existing verified components.

## 10. Every task must have acceptance criteria and tests
No task is actionable without clear acceptance criteria. Every task must produce automated unit, integration, or security tests validating those criteria.

## 11. Every implementation must follow the approved architecture
Adhere to the approved modular monolith architecture, ADRs ([`ADR-001`](../adr/ADR-001-modular-monolith.md) through [`ADR-015`](../adr/ADR-015-workspace-private-network-policy.md)), and [`ARCHITECTURE_VALIDATION.md`](../ARCHITECTURE_VALIDATION.md).

## 12. Security requirements must never be weakened
- Authorization-first: No scan can run without valid, non-revoked authorization.
- Deny by default: Unknown targets, private networks (unless workspace administrator authorized), and invalid profiles are blocked.
- Permanent bans: Cloud metadata (`169.254.169.254`), loopback (`127.0.0.0/8`, `::1`), link-local, and Docker internal subnets can never be assessed under any circumstance.
- Never mount the host Docker socket into application, web, or scanner containers.

## 13. If requirements are ambiguous, ask before making assumptions
Clarify requirements and design choices with the Product Owner before proceeding with architectural changes or implementation.

## 14. Update relevant documentation when an approved decision changes
Keep specifications, ADRs, database designs, and acceptance criteria in sync with confirmed decisions.

## 15. Never mark a task complete without verifying its tests and acceptance criteria
Run the relevant test suites (`go test -race`, linter, static analysis) and confirm that all acceptance criteria are demonstrably satisfied before closing a task.
