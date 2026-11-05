# ADR-003: PostgreSQL as Primary Database and MVP Job Store

- Status: Accepted
- Decision: Use PostgreSQL for transactional data, outbox, and initial job queue.

## Context
MVP does not yet demonstrate a need for a separate broker.

## Consequences
Positive: fewer components and transactional job creation.
Negative: queue throughput and locking must be load-tested; introduce a broker only with evidence.
Use leases, bounded polling, `SKIP LOCKED` or equivalent safe claim strategy, idempotency, and recovery.
