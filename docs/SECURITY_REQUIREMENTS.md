# Security Requirements and Threat Model

## 1. Security principles

- Authorization-first: no scan without explicit, current authorization.
- Deny by default: unknown target, scope, network, or scanner option is rejected.
- Least privilege: API, worker, and scanner containers have separate permissions.
- Defense in depth: application policy plus network egress enforcement.
- Safe defaults: non-destructive, bounded, non-stealth profiles.
- Auditability: security-sensitive decisions and state transitions are recorded.

## 2. Trust boundaries

1. Browser to API.
2. API to PostgreSQL/job queue.
3. API to worker.
4. Worker to scanner container.
5. Scanner container to target network.
6. Scanner output to parser/normalizer.
7. Application to external intelligence feeds.
8. Tenant/user to tenant/user data boundary.

## 3. Threats and mitigations

| Threat | Required mitigation |
|---|---|
| Unauthorized scanning | Authorization record, RBAC, approval policy |
| Scope bypass | Central scope service; request-time and execution-time checks |
| DNS rebinding | Resolve at execution; validate every resolved destination; enforce egress at connection layer |
| SSRF | Strict URL parser, scheme/port policy, deny metadata/link-local/loopback/private ranges by default, controlled egress proxy |
| Private network pivoting | Denied by default; explicit workspace-level admin enablement requires written authorization and target allowlisting; cloud metadata, loopback, and Docker internal subnets permanently blocked (ADR-015) |
| Setup token interception | Ephemeral expiry (24h), single-use, hashed storage (Argon2id/SHA-256), terminal display once, setup endpoint permanently disabled post-bootstrap (ADR-013) |
| Reverse proxy header spoofing | Enforce explicit trusted proxy IP/CIDR validation in API; ignore untrusted client-supplied `X-Forwarded-*` headers (ADR-014) |
| Redirect escape | Revalidate every redirect destination; bounded redirect count |
| IPv4/IPv6 bypass | Canonicalize and validate both address families, mapped addresses, CIDRs |
| Command injection | No arbitrary command strings; fixed executable and typed allowlisted arguments |
| Scanner escape | Non-root, read-only filesystem where possible, dropped capabilities, seccomp/AppArmor, resource limits |
| Malicious scanner output | Treat as untrusted; size limits, parser isolation, HTML escaping |
| Credential leakage | Secret store, redaction, no secrets in command arguments/logs |
| Cross-tenant access | Workspace-scoped authorization and database query constraints |
| Queue replay/duplicate | Idempotency key, unique constraints, state transition guards |
| Resource exhaustion | Per-job and global limits, quotas, queue backpressure |
| Plugin compromise | External plugin runtime deferred; signed artifacts and sandbox required |
| Evidence disclosure | Object-level authorization, encryption, short-lived download links |
| Dependency compromise | Pin versions, SBOM, vulnerability scanning, provenance review |

## 4. Authorization record requirements

Store:
- authorizing organization/person and contact reference;
- project and assets covered;
- evidence document reference/checksum;
- start/end time and timezone;
- allowed test categories and intensity;
- excluded systems and networks;
- approval actor and timestamp;
- revocation status and reason.

The platform stores evidence references and metadata; it should avoid collecting unnecessary personal data.

## 5. Scope enforcement

- Canonicalize domain, hostname, IP, CIDR, URL, and port.
- Use exact matching and carefully defined wildcard semantics.
- Exclusions always override inclusions.
- Do not trust client-supplied resolved IPs.
- Re-resolve names immediately before connection.
- Validate every A/AAAA result and redirect.
- Enforce destination policy at network egress, not only in application code.
- If DNS resolution is ambiguous, mixed, or fails validation, fail closed.
- Private/internal targets (RFC 1918, RFC 4193) are denied by default across the platform. Workspace administrators may explicitly enable private network scanning at the workspace level (ADR-015), subject to mandatory written authorization, strict target allowlisting, and audit logging. Enabling private network scanning NEVER bypasses target authorization or scope validation. Cloud metadata (`169.254.169.254`), host loopback (`127.0.0.0/8`, `::1`), link-local (`169.254.0.0/16`, `fe80::/10`), and internal Docker container network subnets remain permanently blocked under all circumstances.
- SaaS public egress must not reach tenant metadata services, control planes, or internal infrastructure.

## 6. Scanner execution controls

- Use fixed, reviewed profiles and typed options.
- No shell interpolation or arbitrary command execution.
- Run each scan in a constrained container/process.
- Apply CPU, memory, PID, disk, output, network, and wall-clock limits.
- Use non-root UID, minimal filesystem, dropped Linux capabilities, and restricted syscalls where available.
- Scanner worker has no access to application signing keys or unrelated tenant secrets.
- Cancellation sends cooperative stop and then enforces termination after a grace period.
- Capture exit code, tool version, profile version, timestamps, and bounded logs.

## 7. Data classification

- Public: product documentation and public vulnerability metadata.
- Internal: workspace/project metadata.
- Confidential: asset inventory, scan configuration, findings.
- Restricted: credentials, authorization evidence, raw request/response evidence, exported reports.

Restricted data must be encrypted at rest where supported, access-controlled, redacted from logs, and excluded from analytics by default.

## 8. Security testing requirements

- Unit tests for scope matching, canonicalization, exclusions, authorization expiry, role checks, state transitions.
- Integration tests for tenant isolation and repository filters.
- Negative tests for localhost, private ranges, link-local, metadata IPs, IPv6, redirects, DNS rebinding scenarios, malformed URLs, and mixed DNS answers.
- Tests proving worker revalidation occurs after queue delay.
- Tests proving cancellation and timeout terminate scanner processes.
- Fuzz tests for URL, hostname, IP/CIDR, and scanner output parsers.
- Dependency, container, secret, and static analysis in CI.
- Independent security review before public SaaS scanning.

## 9. Incident response

Provide documented procedures for disabling scan execution globally, revoking a scanner profile, suspending a workspace, rotating secrets, preserving audit evidence, and notifying affected operators/users.
