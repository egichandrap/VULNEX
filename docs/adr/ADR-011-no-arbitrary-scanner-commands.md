# ADR-011: No Arbitrary Scanner Command Execution

- Status: Accepted
- Decision: Users select reviewed scanner profiles with typed allowlisted options. No API/UI accepts arbitrary shell commands or unrestricted scanner flags.

## Context
Arbitrary execution would create command injection, privilege escalation, scope bypass, and abuse risks.

## Consequences
New scanner options require code review, validation, tests, documentation, and profile versioning. Advanced operator-only configuration must still be constrained and audited.
