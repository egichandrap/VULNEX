# ADR-001: Modular Monolith for MVP

- Status: Accepted
- Decision: Build core capabilities as bounded modules in one Go codebase; run API and worker as separate processes.

## Context
The product is early-stage and needs fast iteration without distributed transaction and operational overhead.

## Decision
Use a modular monolith with explicit DDD boundaries and separate API/worker entrypoints. Scanner processes are isolated runtime units.

## Consequences
Positive: simple deployment, strong local transactions, easier refactoring within clear boundaries.
Negative: requires architecture tests and discipline to prevent module coupling.
Revisit when independent ownership/scaling/deployment is demonstrated.
