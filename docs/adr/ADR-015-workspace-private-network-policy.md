# ADR-015: Workspace-Level Private Network Testing Policy

- Status: Accepted
- Decision: Private/internal network scanning is disabled by default across the entire platform, but may be explicitly enabled at the workspace level by an authorized administrator, subject to immutable safety guardrails.

## Context
VULNEX is designed for defensive assessment. Many legitimate enterprise security teams need to assess internal infrastructure (RFC 1918 subnets, internal corporate domains). However, unconditional private network scanning introduces critical Server-Side Request Forgery (SSRF) and network pivot risks in multi-tenant or shared server deployments.

## Decision
1. **Disabled by Default:**
   - All network destinations in private IPv4 ranges (`10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`), IPv6 unique local addresses (`fc00::/7`), and loopback are denied by default.
2. **Explicit Workspace-Level Enablement:**
   - A workspace administrator can toggle `allow_private_network_scan` on the Workspace aggregate.
   - Enabling this setting requires:
     - Explicit written authorization uploaded and bound to the workspace/project.
     - Documented asset and target scope allowlisting.
     - Comprehensive audit logging (`actor_user_id`, timestamp, reason, IP address).
3. **Immutable Safety Guardrails (Non-Bypassable):**
   - Enabling private network scanning **NEVER** bypasses target authorization, project boundary, or scope validation rules.
   - The following destinations remain **permanently blocked** even when private network scanning is enabled:
     - Cloud provider metadata endpoints (e.g., `169.254.169.254`).
     - Host loopback addresses (`127.0.0.0/8`, `::1`).
     - Link-local addresses (`169.254.0.0/16`, `fe80::/10`).
     - Internal Docker container subnets and management plane networks (to prevent scanning the VULNEX database, redis, or API internal ports).
4. **Enforcement Layer:**
   - Validation occurs at connection time in `NetworkDestinationPolicy` after DNS resolution.
   - If DNS resolution produces mixed public and prohibited internal addresses, resolution fails closed.

## Consequences
- **Positive:** Enables legitimate enterprise internal assessments while preventing unauthorized network pivoting and protecting infrastructure metadata services.
- **Negative:** Requires continuous network-level policy filtering in worker containers and strict DNS rebinding inspection.
