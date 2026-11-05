# VULNEX Installation Guide

> This guide documents the intended installer experience. Replace example release URLs, registry names, minimum runtime versions, and platform support statements with verified release values before publishing a production release.

## 1. Requirements

- Supported OS and CPU architecture listed in the release manifest.
- Docker Engine and Docker Compose Plugin v2.
- Network access to the official VULNEX release and container registry for online installation.
- Adequate CPU, memory, and disk according to the release's published sizing guide.
- Administrator privileges only for installing the CLI or managing system-level runtime components.

The VULNEX CLI does not bundle Docker Engine. If Docker is missing, `vulnex doctor` provides installation guidance. Any automated system package installation must be explicitly approved.

## 2. Download and verify

Download the correct artifact from the official Releases page. Verify the SHA-256 checksum and signature using the verification instructions published with that release.

Do not execute artifacts from unofficial mirrors.

## 3. Linux CLI installation

Example for Linux amd64:

```bash
tar -xzf vulnex-linux-amd64.tar.gz
sudo install -m 0755 vulnex /usr/local/bin/vulnex
vulnex version
```

For a package-based release, use the package manager and exact package published for your distribution.

## 4. Diagnose prerequisites

```bash
vulnex doctor
```

Resolve any reported unsupported OS, missing Docker Engine, missing Compose plugin, insufficient disk/memory, permission, or registry connectivity issue before continuing.

## 5. Local installation

```bash
vulnex install --mode local
vulnex status
vulnex dashboard
```

Local mode must bind the dashboard to loopback by default. Do not configure public ingress for a local installation.

## 6. Server installation

Before installation:
- Configure a supported hostname.
- Configure TLS and the supported reverse proxy/ingress path.
- Ensure inbound firewall rules expose only the required web entry point.
- Keep database, Redis, and worker ports private.
- Prepare persistent storage and backup destination.

Then run:

```bash
vulnex doctor
vulnex install --mode server --domain security.example.com
vulnex status
```

The exact TLS and reverse-proxy flags depend on the implemented release. Consult `vulnex install --help` and the release-specific server configuration guide.

## 7. Lifecycle

```bash
vulnex status
vulnex start
vulnex stop
vulnex restart
vulnex logs --follow
vulnex dashboard
```

## 8. Configuration

```bash
vulnex config validate
vulnex config show
```

Secret values must be redacted. Edit configuration only using documented fields and run validation before applying changes.

## 9. Backup and restore

```bash
vulnex backup --output ./vulnex-backup
vulnex restore --file ./vulnex-backup
```

Backups may contain sensitive assessment data. Store them encrypted, restrict access, and test restoration periodically. The exact archive format and encryption behavior must be documented by the implemented release.

## 10. Update

```bash
vulnex update
```

Before upgrading, the CLI must verify release authenticity, check database migration compatibility, and create or require a valid recovery point when the migration can affect persistent data.

## 11. Uninstall

Choose explicitly whether to retain or delete persistent data:

```bash
vulnex uninstall --keep-data
```

or:

```bash
vulnex uninstall --delete-data
```

The destructive option must require an additional confirmation. VULNEX uninstall must not remove unrelated Docker images, containers, networks, or volumes.

## 12. Troubleshooting

| Symptom | Action |
|---|---|
| Docker unavailable | Run `vulnex doctor`; install/enable supported Docker runtime |
| Compose plugin missing | Install supported Compose v2 plugin |
| Permission denied | Check current user permissions and Docker access; do not run the whole application as root |
| Registry unreachable | Check DNS, proxy, TLS trust, and outbound connectivity |
| Health check timeout | Inspect `vulnex status` and `vulnex logs` |
| Disk space low | Check persistent data and logs; do not manually delete database volumes |
| Upgrade failed | Preserve logs and backup; follow recovery instructions for that exact release |

## 13. Safe operation

Only assess systems you own or are explicitly authorized to test. Configure scope and authorization in VULNEX before scheduling scans. Never expose scanner worker control ports to untrusted networks.
