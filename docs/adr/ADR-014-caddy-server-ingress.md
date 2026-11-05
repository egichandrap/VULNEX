# ADR-014: Embedded Caddy Reverse Proxy for Server Ingress and TLS

- Status: Accepted
- Decision: Use embedded Caddy as the default ingress component for Server mode with automatic ACME TLS provisioning, while supporting external reverse proxies with explicit trusted CIDR validation.

## Context
Server deployments require HTTPS termination, secure HTTP header handling, static asset routing, and reverse proxying to backend services without manual certificate management complexity.

## Decision
1. **Default Ingress:** Embedded Caddy container integrated into the Server mode Docker Compose topology.
   - Automatically handles ACME (Let's Encrypt / ZeroSSL) certificate procurement and renewal for configured `--domain`.
   - Terminates TLS, redirects HTTP to HTTPS, and reverse-proxies requests to `frontend:3000` and `api:8080`.
2. **External Reverse Proxy Support:**
   - Operators may opt to run an external reverse proxy / load balancer (e.g. Cloudflare, AWS ALB, enterprise Nginx) by configuring VULNEX in `--external-proxy` mode.
3. **Trusted Proxy Enforcement:**
   - The Go API must enforce explicit trusted proxy IP/CIDR configuration.
   - Client forwarding headers (`X-Forwarded-For`, `X-Forwarded-Proto`, `X-Real-IP`) must only be accepted from explicitly configured trusted upstream proxy addresses.
   - Arbitrary client-supplied forwarding headers from untrusted sources must be ignored or stripped to prevent IP spoofing, rate-limit evasion, and host-header attacks.

## Consequences
- **Positive:** Zero-configuration automated TLS out of the box for server operators; standard reverse proxy capabilities; robust protection against proxy header spoofing.
- **Negative:** Requires inbound ports 80 and 443 accessible for ACME HTTP/ALPN challenges when using default automated certificates.
