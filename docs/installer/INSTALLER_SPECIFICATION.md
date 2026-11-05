# Installer Specification

## 1. CLI contract

The executable name is `vulnex`. Commands are designed for both interactive use and automation.

```text
vulnex
  version
  doctor [--json]
  install --mode <local|server> [options]
  status [--json]
  start
  stop
  restart
  logs [--service <name>] [--follow]
  dashboard
  config validate
  config show
  update [--version <version>]
  backup [--output <path>]
  restore --file <path>
  uninstall [--keep-data|--delete-data]
```

All command behavior must be documented and tested before being included in a stable release.

## 2. Install options

| Option | Meaning |
|---|---|
| `--mode` | Required: local or server |
| `--domain` | Public hostname for server mode (used for Caddy ACME TLS) |
| `--external-proxy` | Optional: offload TLS to an external proxy instead of embedded Caddy |
| `--trusted-proxies` | CIDR ranges of trusted reverse proxies when using external proxy |
| `--version` | Exact VULNEX release |
| `--data-dir` | Persistent data directory |
| `--config` | Configuration file path |
| `--non-interactive` | Disable prompts; missing required values cause failure |
| `--yes` | Confirm only explicitly documented non-destructive defaults; never bypass destructive confirmation |

Server mode must reject unsafe configurations such as public database bindings or missing TLS configuration for remote access.

## 3. Runtime requirements

- Docker Engine version: define minimum supported version in release compatibility manifest.
- Docker Compose Plugin v2: required for initial managed runtime.
- Sufficient CPU, memory, and disk: exact minimums to be established from benchmarked release profile.
- Supported OS/architecture: only those listed in release manifest.
- Network access to the official registry for online installation.

The CLI must distinguish missing runtime, unsupported runtime, insufficient resources, permission problems, and registry connectivity errors.

## 4. Managed runtime behavior

- Detect Docker; do not assume it exists.
- If missing, show platform-specific installation guidance.
- Automated package installation is opt-in, platform-specific, and requires administrator consent.
- Do not silently alter firewall, SSH, system service, or kernel settings.
- Do not bundle Docker Engine in the CLI binary.
- Use Compose project name and labels unique to VULNEX.
- Never prune global Docker resources as part of ordinary uninstall.

## 5. Configuration

Configuration is validated against a versioned schema. Secrets are generated using a cryptographically secure random source and stored with restrictive filesystem permissions or a supported OS secret store.

Never print secret values. `config show` must redact secrets. Configuration writes should be atomic and preserve the previous valid configuration on failure.

## 6. Local profile

- Bind web dashboard to loopback by default.
- Do not expose database, Redis, or worker ports.
- Use persistent named volumes or configured persistent directories.
- Authentication remains enabled.
- Provide dashboard URL after readiness.

## 7. Server profile

- Require explicit network exposure configuration.
- Require TLS for remote dashboard access: provision embedded Caddy container with automatic ACME certificate management by default (ADR-014).
- If `--external-proxy` is selected, enforce `--trusted-proxies` CIDR validation to prevent client-supplied header spoofing.
- Keep database and Redis on private container networks.
- Use least-privilege service accounts and container settings.
- Require secure administrator bootstrap: generate an ephemeral, single-use setup token, store only its hash in database, and display it once in the terminal alongside onboarding instructions (`https://<domain>/setup`) (ADR-013).
- Provide backup configuration and health checks.

## 8. Release manifest

The release manifest should include:
- Product version and release channel.
- Supported OS and architecture.
- Component image names and immutable digests.
- Compose schema/version.
- Database migration compatibility range.
- Minimum runtime versions.
- Artifact checksums and signature metadata.
- SBOM references.

## 9. Exit codes

| Code | Meaning |
|---:|---|
| 0 | Success |
| 1 | General failure |
| 2 | Invalid arguments/configuration |
| 3 | Unsupported environment |
| 4 | Runtime unavailable |
| 5 | Artifact verification failed |
| 6 | Deployment/health check failed |
| 7 | Backup/restore failed |
| 8 | Upgrade compatibility/recovery failure |
| 9 | Operation cancelled |

## 10. Idempotency and concurrency

- Repeated install must detect an existing installation.
- Concurrent mutating commands must be serialized with an installation lock.
- Interrupted operations must leave a recoverable state record.
- Start/stop/restart must be safe to retry.
- Backup and restore must verify integrity and target installation identity.
