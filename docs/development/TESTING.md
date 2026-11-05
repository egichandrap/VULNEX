# VULNEX Development & Testing Guide

## 1. Purpose

This guide defines the test strategy for contributors and maintainers. The CLI and application are separate deployable components but share release versioning and integration contracts.

## 2. Development prerequisites

The exact tool versions must be pinned in CI and documented in the repository toolchain files. Typical requirements:
- Supported Go toolchain.
- Node.js and package manager for frontend development.
- Docker Engine and Docker Compose v2 for integration/E2E tests.
- Make.
- Linters and security scanners used by CI.

End users installing a release do not need Go or Node.js.

## 3. Test layers

### Unit tests
- Domain entities, value objects, invariants.
- Application use cases with mocked ports.
- Configuration validation.
- CLI argument parsing and exit codes.
- Scope validation and authorization policies.

### Adapter/integration tests
- Docker runtime adapter against a disposable test environment.
- Registry manifest parsing and artifact verification.
- Filesystem permissions and atomic configuration writes.
- PostgreSQL repository and migration tests.
- Scanner adapter contract tests using safe fixtures.

### End-to-end tests
- Fresh Local installation.
- Fresh Server installation in an isolated test network.
- Health/readiness.
- Start, stop, restart, logs.
- Backup and restore.
- Upgrade between compatible fixture releases.
- Uninstall with data retained.
- Uninstall with data deletion in disposable test environment only.

### Security tests
- Reject missing/invalid authorization.
- Reject out-of-scope target.
- SSRF and DNS rebinding test cases.
- Reject arbitrary scanner command injection.
- Enforce scan timeout and resource limits.
- Validate cancellation.
- Verify secrets are redacted from logs and diagnostics.
- Verify scanner containers cannot access Docker socket or management-plane secrets.

## 4. Commands

The repository should provide these Make targets once implemented:

```bash
make fmt
make lint
make test
make test-race
make test-integration
make test-e2e
make test-security
make build
```

Suggested direct Go commands:

```bash
go test ./...
go test -race ./...
go vet ./...
```

Frontend:

```bash
npm ci
npm run lint
npm run test
npm run build
```

Do not claim a target exists until the Makefile/package scripts implement it.

## 5. TDD requirements

- Write failing tests before implementation (Red).
- Implement the smallest correct behavior (Green).
- Refactor without changing behavior (Refactor).
- Test domain rules independently from Docker, HTTP, SQL, and filesystem.
- Use table-driven tests and subtests for Go.
- Use fakes for application-level tests; use mocks only where interaction verification is needed.
- Prefer real disposable dependencies for adapter integration tests.

## 6. Coverage and quality gates

- Domain and application packages: target 90%+ statement coverage, with meaningful branch and invariant tests.
- Overall backend/installer coverage: target 80%+ as an initial quality gate.
- No untested security-critical scope/authorization decision.
- `go test -race` for concurrent components.
- Lint, static analysis, dependency vulnerability scan, and secret scan in CI.
- Coverage percentage never substitutes for test quality.

## 7. Safe test data

Use loopback test servers, local fixtures, mock scanner outputs, and intentionally vulnerable lab containers in an isolated network. Never run integration or E2E tests against public targets or assets without explicit authorization. Avoid destructive tests against shared infrastructure.

## 8. CI pipeline

1. Format and lint.
2. Unit tests.
3. Race tests.
4. Build CLI and application.
5. Integration tests.
6. Security tests.
7. Build container images.
8. Generate SBOM and scan artifacts.
9. Sign and publish only from protected release workflows.

## 9. Test evidence

For each release, retain:
- CI run reference.
- Supported OS/architecture matrix.
- Unit/integration/E2E result.
- Image scan and SBOM.
- Artifact checksum/signature verification result.
- Upgrade and restore test result.
