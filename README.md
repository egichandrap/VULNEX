# VULNEX

**Open-source Vulnerability Assessment & Penetration Testing Management Platform**

VULNEX is a defensive security assessment platform for systems that the operator owns or is explicitly authorized to assess. It combines project and asset management, controlled scanner orchestration, findings management, and reporting.

> **Authorization required:** Only assess assets that you own or are explicitly authorized to test. VULNEX is designed for controlled defensive assessment. It does not provide stealth, persistence, credential theft, arbitrary command execution, or automatic destructive exploitation.

## Product goals

- Self-hosted deployment on a workstation or server.
- One unified installer and one CLI for Local and Server deployment modes.
- Managed application runtime using Docker Engine and Docker Compose.
- Go backend, Next.js frontend, PostgreSQL, optional Redis, and isolated scanner workers.
- Integrations with approved scanner tools through constrained adapters.
- DDD, Clean Architecture, SOLID, and TDD are mandatory engineering standards.

## Quick start

The commands below describe the intended release experience. They become executable when the corresponding release artifacts and container images are published.

### 1. Install the VULNEX CLI

Run the automated POSIX installer:

```bash
curl -fsSL https://raw.githubusercontent.com/egichandrap/VULNEX/main/scripts/install.sh | sh
```

Or download manually from GitHub Releases, verify the SHA-256 checksum, and install to `/usr/local/bin`:

```bash
tar -xzf vulnex_1.1.0_linux_amd64.tar.gz
sudo install -m 0755 vulnex /usr/local/bin/vulnex
vulnex version
vulnex doctor
```

macOS and Windows packages will be distributed as platform-specific archives/packages. See [Installation Guide](docs/deployment/INSTALLATION.md).

### 2. Install locally

```bash
vulnex install --mode local
vulnex status
vulnex dashboard
```

Local mode binds the web interface to loopback by default. It must not expose the dashboard or scanner services to the public network by default.

### 3. Install on a server

```bash
vulnex doctor
vulnex install --mode server --domain security.example.com
vulnex status
```

Server mode requires an explicit network/TLS configuration. Do not expose PostgreSQL, Redis, or scanner worker ports publicly.

### 4. Uninstall VULNEX

VULNEX requires an explicit data retention choice to prevent accidental data loss:

```bash
# Option A: Stop/remove containers while strictly preserving persistent data & databases (Recommended)
vulnex uninstall --keep-data

# Option B: Complete wipe removing containers and permanently deleting persistent data
vulnex uninstall --delete-data

# (Optional) Remove the CLI binary from your executable path
sudo rm -f /usr/local/bin/vulnex
```

> **Tip:** You can create a database backup prior to uninstalling with `vulnex backup --output ./vulnex-backup`.

### 5. Run tests

For contributors building from source:

```bash
make test
make lint
make test-integration
make test-e2e
```

See [Testing Guide](docs/development/TESTING.md) for prerequisites, test layers, coverage expectations, and safe test fixtures.

## CLI command reference

| Command | Description |
|---|---|
| `vulnex install --mode local` | Install local deployment |
| `vulnex install --mode server` | Install server deployment |
| `vulnex doctor` | Diagnose OS, resources, Docker and Compose |
| `vulnex status` | Show component health |
| `vulnex start` / `stop` / `restart` | Manage application lifecycle |
| `vulnex logs` | Read application logs |
| `vulnex dashboard` | Print/open dashboard URL |
| `vulnex update` | Verify and apply an application update |
| `vulnex backup` / `restore` | Backup and restore persistent data |
| `vulnex config validate` | Validate configuration |
| `vulnex uninstall` | Remove application with explicit data-retention choice |

Commands and flags are part of the product specification; release documentation must reflect the actual CLI implementation. Do not advertise a command before it is implemented.

## Architecture and documentation

- [Product Requirements Document](docs/PRD.md)
- [Installer Architecture](docs/installer/INSTALLER_ARCHITECTURE.md)
- [Installer Specification](docs/installer/INSTALLER_SPECIFICATION.md)
- [Distribution Strategy](docs/installer/DISTRIBUTION_STRATEGY.md)
- [Deployment Guide](docs/deployment/INSTALLATION.md)
- [Testing Guide](docs/development/TESTING.md)
- [Installer Development Tasks](docs/installer/INSTALLER_TASKS.md)
- [Installer Acceptance Criteria](docs/installer/INSTALLER_ACCEPTANCE_CRITERIA.md)
- [System Architecture](docs/SYSTEM_ARCHITECTURE.md)
- [Security Requirements](docs/SECURITY_REQUIREMENTS.md)

## Development status

This repository documentation defines the target product and intended workflows. It is not a claim that all listed application features or installer commands are already implemented. Use release notes and the project issue tracker to identify completed functionality.

## Security reporting

Do not publish exploit details or sensitive target information in public issues. Follow the project's private vulnerability disclosure process once configured.
