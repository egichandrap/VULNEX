# ADR-012: Unified Monorepo Repository Structure

- Status: Accepted
- Decision: Place backend, frontend, installer CLI, deployment specifications, tests, and documentation into a single Git repository while enforcing independent module boundaries and release artifacts.

## Context
VULNEX consists of multiple deployable components: the Go backend (API and worker), the Next.js frontend, the Go CLI installer (`vulnex`), and Docker Compose deployment manifests. Operating separate repositories would complicate cross-component version coordination, atomic integration testing, schema synchronization, and end-to-end release automation during early-stage development.

## Decision
Adopt a unified monorepo structure:
- `backend/`: Go modular monolith containing API (`cmd/api`) and worker (`cmd/worker`), domain contexts (`internal/`), and migrations (`migrations/`).
- `frontend/`: Next.js, TypeScript, and Tailwind CSS dashboard.
- `installer/`: Standalone Go CLI installer application (`cmd/vulnex`) with independent Go module, isolated domain/application layers, and separate release packaging.
- `deploy/`: Docker Compose configurations, Caddyfile configurations, and environment profiles.
- `docs/`: Comprehensive specifications, ADRs, user stories, and guides.
- `tests/`: End-to-end test suites and shared fixtures.

Each subsystem must maintain distinct package/module boundaries. Specifically:
- The installer CLI must not import backend business logic or database drivers.
- The backend domain must not import installer or deployment packages.
- CI workflows must validate both isolated unit tests for each component and monorepo-wide integration tests.

## Consequences
- **Positive:** Single commit history, simplified cross-component changes (e.g. API contracts and UI), unified issue tracking, and synchronized release tags.
- **Negative:** Requires rigorous CI pipeline hygiene, path filtering for targeted builds, and strict linting/architecture tests to prevent cross-component leakage.
