# VULNEX Installation & Operator Guide

This guide documents the installation, runtime verification, and lifecycle management workflows for VULNEX.

---

## 1. Prerequisites & System Requirements

- **Supported Operating Systems**: Linux (`amd64`, `arm64`), macOS (`amd64`, `arm64`).
- **Container Runtime**: Docker Engine (v24+) and Docker Compose Plugin (v2.20+).
- **Hardware Resources**:
  - Minimum: 2 CPU cores, 2 GB RAM, 10 GB free disk space.
  - Recommended: 4 CPU cores, 4 GB+ RAM for parallel vulnerability assessments.
- **Network Ports**:
  - Local Mode: `127.0.0.1:3000` (loopback only, zero public exposure).
  - Server Mode: `80` and `443` (for Caddy automated ACME TLS).

> [!NOTE]
> VULNEX CLI does not bundle Docker Engine. Users must install Docker Engine and possess group permissions to communicate with `/var/run/docker.sock` without root escalation.

---

## 2. Installing the VULNEX CLI Control Plane

### Automated POSIX Install Script (Linux & macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/egichandrap/VULNEX/main/scripts/install.sh | sh
```

The installation script performs:
1. Host OS and CPU architecture detection (`linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`).
2. Release tarball download from official GitHub Releases.
3. Cryptographic SHA-256 checksum verification against published release digests.
4. Safe binary extraction and installation to `/usr/local/bin/vulnex` (`0755` permissions).

---

## 3. Pre-flight Environment Diagnostics (`vulnex doctor`)

Before initiating deployment, run diagnostics to verify runtime prerequisites:

```bash
vulnex doctor
```

Or format as JSON for automated CI/CD runners:
```bash
vulnex doctor --json
```

Checks conducted:
- Supported OS & CPU architecture.
- Docker daemon status and socket connectivity permissions.
- Docker Compose v2 plugin availability.
- Available system memory and storage capacity.
- Port availability (`3000` for Local mode; `80`/`443` for Server mode).
- Container registry reachability.

---

## 4. Platform Provisioning (`vulnex install`)

### Interactive Text User Interface (TUI) Mode

Simply run without arguments to start the interactive onboarding wizard:

```bash
vulnex install
```

The wizard guides through:
1. Selecting **Local Mode** or **Server Mode**.
2. Setting domain name and Ingress type (Embedded Caddy with ACME TLS or External Proxy with trusted CIDRs).
3. Configuring persistent data directory (default `./vulnex-data` or `/var/lib/vulnex`).
4. Displaying the **single-use ephemeral administrator setup token** (ADR-013).

### Non-Interactive Local Deployment

```bash
vulnex install --mode local --non-interactive
```

### Non-Interactive Server Deployment

```bash
vulnex install --mode server --domain security.example.com --non-interactive
```

Using an external reverse proxy (ALB / Cloudflare / Nginx):
```bash
vulnex install --mode server --domain security.example.com --external-proxy --trusted-proxies "172.16.0.0/12,10.0.0.0/8" --non-interactive
```

---

## 5. Opening the Dashboard (`vulnex dashboard`)

```bash
vulnex dashboard
```

Opens the local (`http://localhost:3000`) or server (`https://<domain>`) dashboard URL in the default browser.

---

## 6. Stack Lifecycle Management

```bash
# Check container status and component health
vulnex status

# Start stopped stack
vulnex start

# Stop running stack
vulnex stop

# Restart stack
vulnex restart

# Stream real-time container logs
vulnex logs --follow
vulnex logs --service api
```

---

## 7. Configuration Inspection (`vulnex config`)

```bash
# Validate active configuration without applying changes
vulnex config validate --config ./vulnex-data/config.json

# Display configuration with all database secrets redacted
vulnex config show --config ./vulnex-data/config.json
```

---

## 8. Backup and Restoration (`vulnex backup` & `vulnex restore`)

```bash
# Export persistent database and configuration to timestamped SQL dump
vulnex backup --output ./backups

# Restore database from backup archive
vulnex restore --file ./backups/vulnex-backup-20261105-180000.sql
```

---

## 9. Updating the Platform (`vulnex update`)

```bash
vulnex update --version 1.0.1
```

---

## 10. Uninstallation (`vulnex uninstall`)

VULNEX enforces an explicit data retention policy during uninstallation:

```bash
# Tear down containers while strictly preserving database volumes and assets:
vulnex uninstall --keep-data

# Or completely remove containers and delete all persistent data:
vulnex uninstall --delete-data
```
