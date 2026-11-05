# ADR-013: Initial Administrator Bootstrapping via Ephemeral Setup Token

- Status: Accepted
- Decision: Bootstrap the initial system administrator using an ephemeral, cryptographically secure one-time setup token displayed once in the installer terminal, with registration completed via a web onboarding flow.

## Context
In self-hosted and server deployments, provisioning the initial super-administrator securely without hardcoded credentials, cleartext environment variables, or exposing open registration endpoints is critical.

## Decision
1. During installation (`vulnex install --mode server` or local setup), the installer generates a high-entropy, cryptographically secure random token (e.g. 32 bytes hex/base64-encoded).
2. The installer outputs the setup token and the web onboarding URL exactly once in the terminal stdout.
3. The token is never stored in plaintext in the database or logs; only its cryptographic hash (e.g., Argon2id or SHA-256) is stored in a dedicated `setup_tokens` record with an explicit expiration (e.g., 24 hours).
4. The user completes administrator creation via the web onboarding endpoint (`POST /api/v1/auth/setup`), submitting the token alongside the initial administrator's email, display name, and password.
5. Upon successful creation of the first administrator:
   - The token is marked used (`used_at`) and invalidated.
   - The setup endpoint is permanently disabled.
   - Future registrations require normal invite/workspace authorization flows.

## Consequences
- **Positive:** No hardcoded passwords, zero default credentials, no plaintext secret persistence in configuration files, and zero-trust web onboarding.
- **Negative:** If the terminal output is cleared or the administrator fails to record the token before expiration, an installer recovery command (`vulnex admin reset-setup-token`) must be provided via the local CLI control plane.
