# ADR-007: Isolated Scanner Workers

- Status: Accepted
- Decision: Scanner execution runs outside the API process in restricted workers/containers.

## Context
Scanners are external programs and untrusted output/process behavior must not compromise the API.

## Consequences
Use least privilege, resource limits, bounded output, no arbitrary command interface, and controlled egress. Do not expose Docker socket to the general API.
