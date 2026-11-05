-- VULNEX Database Baseline Schema (v1.0.0)
-- Conforms to docs/DATABASE_DESIGN.md and docs/ARCHITECTURE_VALIDATION.md

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================================
-- 1. Identity Context
-- ============================================================================

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email_normalized VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    display_name VARCHAR(120) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);

CREATE UNIQUE INDEX idx_users_email_normalized ON users (email_normalized);

-- Initial Admin Setup Tokens (ADR-013)
CREATE TABLE setup_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token_hash VARCHAR(128) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);

CREATE UNIQUE INDEX idx_setup_tokens_hash ON setup_tokens (token_hash);
CREATE INDEX idx_setup_tokens_pending ON setup_tokens (expires_at) WHERE used_at IS NULL;

-- ============================================================================
-- 2. Workspace Context
-- ============================================================================

CREATE TABLE workspaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(120) NOT NULL,
    owner_user_id UUID NOT NULL REFERENCES users(id),
    allow_private_network_scan BOOLEAN NOT NULL DEFAULT FALSE,
    status VARCHAR(32) NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);

CREATE TABLE workspace_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(32) NOT NULL DEFAULT 'MEMBER',
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);

CREATE UNIQUE INDEX idx_workspace_members_user ON workspace_members (workspace_id, user_id);

-- ============================================================================
-- 3. Project Context
-- ============================================================================

CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    name VARCHAR(120) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status VARCHAR(32) NOT NULL DEFAULT 'ACTIVE',
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);

CREATE INDEX idx_projects_workspace_status ON projects (workspace_id, status);

-- ============================================================================
-- 4. Asset Context
-- ============================================================================

CREATE TABLE assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    asset_type VARCHAR(32) NOT NULL,
    canonical_value VARCHAR(512) NOT NULL,
    display_value VARCHAR(512) NOT NULL,
    criticality VARCHAR(32) NOT NULL DEFAULT 'MEDIUM',
    environment VARCHAR(32) NOT NULL DEFAULT 'PRODUCTION',
    status VARCHAR(32) NOT NULL DEFAULT 'ACTIVE',
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);

CREATE UNIQUE INDEX idx_assets_project_canonical ON assets (project_id, asset_type, canonical_value);

-- ============================================================================
-- 5. Authorization Context
-- ============================================================================

CREATE TABLE authorizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    status VARCHAR(32) NOT NULL DEFAULT 'ACTIVE',
    authorized_party VARCHAR(255) NOT NULL,
    approver_user_id UUID NOT NULL REFERENCES users(id),
    evidence_reference VARCHAR(512) NOT NULL,
    evidence_checksum VARCHAR(64) NOT NULL,
    valid_from TIMESTAMPTZ NOT NULL,
    valid_until TIMESTAMPTZ NOT NULL,
    allowed_methods JSONB NOT NULL DEFAULT '[]'::jsonb,
    revoked_at TIMESTAMPTZ,
    revocation_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);

CREATE INDEX idx_authorizations_project ON authorizations (project_id, status);

CREATE TABLE authorization_scopes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    authorization_id UUID NOT NULL REFERENCES authorizations(id) ON DELETE CASCADE,
    rule_type VARCHAR(16) NOT NULL, -- 'INCLUDE' or 'EXCLUDE'
    target_type VARCHAR(32) NOT NULL,
    canonical_target VARCHAR(512) NOT NULL,
    port_range VARCHAR(64),
    notes TEXT NOT NULL DEFAULT ''
);

CREATE INDEX idx_auth_scopes_auth_id ON authorization_scopes (authorization_id);

-- ============================================================================
-- 6. Scanner Catalog Context
-- ============================================================================

CREATE TABLE scanner_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(120) NOT NULL,
    scanner_type VARCHAR(64) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status VARCHAR(32) NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);

CREATE TABLE scanner_profile_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scanner_profile_id UUID NOT NULL REFERENCES scanner_profiles(id) ON DELETE CASCADE,
    version VARCHAR(32) NOT NULL,
    configuration JSONB NOT NULL DEFAULT '{}'::jsonb,
    approved_by UUID NOT NULL REFERENCES users(id),
    approved_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    checksum VARCHAR(64) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'APPROVED'
);

CREATE UNIQUE INDEX idx_scanner_profile_version ON scanner_profile_versions (scanner_profile_id, version);

-- ============================================================================
-- 7. Scanning Context
-- ============================================================================

CREATE TABLE scans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    authorization_id UUID NOT NULL REFERENCES authorizations(id),
    scanner_profile_version_id UUID NOT NULL REFERENCES scanner_profile_versions(id),
    requested_by UUID NOT NULL REFERENCES users(id),
    status VARCHAR(32) NOT NULL DEFAULT 'QUEUED',
    idempotency_key VARCHAR(128) NOT NULL,
    target_snapshot JSONB NOT NULL DEFAULT '[]'::jsonb,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);

CREATE UNIQUE INDEX idx_scans_idempotency ON scans (workspace_id, project_id, idempotency_key);
CREATE INDEX idx_scans_status_created ON scans (workspace_id, status, created_at DESC);

CREATE TABLE scan_targets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scan_id UUID NOT NULL REFERENCES scans(id) ON DELETE CASCADE,
    asset_id UUID NOT NULL REFERENCES assets(id),
    canonical_target VARCHAR(512) NOT NULL,
    target_status VARCHAR(32) NOT NULL DEFAULT 'PENDING',
    resolved_snapshot JSONB NOT NULL DEFAULT '[]'::jsonb,
    validation_result VARCHAR(64) NOT NULL DEFAULT 'VALID'
);

CREATE TABLE scan_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scan_id UUID NOT NULL REFERENCES scans(id) ON DELETE CASCADE,
    status VARCHAR(32) NOT NULL DEFAULT 'QUEUED',
    attempt INT NOT NULL DEFAULT 0,
    lease_owner VARCHAR(128),
    lease_expires_at TIMESTAMPTZ,
    available_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    last_error_code VARCHAR(64),
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);

CREATE INDEX idx_scan_jobs_lease_claim ON scan_jobs (status, available_at, lease_expires_at);

CREATE TABLE outbox_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_type VARCHAR(64) NOT NULL,
    aggregate_id UUID NOT NULL,
    event_type VARCHAR(128) NOT NULL,
    payload JSONB NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    published_at TIMESTAMPTZ,
    attempts INT NOT NULL DEFAULT 0
);

CREATE INDEX idx_outbox_events_pending ON outbox_events (published_at, occurred_at);

-- ============================================================================
-- 8. Findings Context
-- ============================================================================

CREATE TABLE findings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    scan_id UUID NOT NULL REFERENCES scans(id) ON DELETE CASCADE,
    asset_id UUID NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    fingerprint VARCHAR(64) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    severity VARCHAR(32) NOT NULL,
    confidence VARCHAR(32) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'OPEN',
    source_scanner VARCHAR(64) NOT NULL,
    first_seen_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    risk_model_version VARCHAR(32) NOT NULL DEFAULT 'v1.0.0',
    risk_score NUMERIC(5,2) NOT NULL DEFAULT 0.00
);

CREATE UNIQUE INDEX idx_findings_project_fingerprint ON findings (project_id, fingerprint);
CREATE INDEX idx_findings_query ON findings (workspace_id, project_id, status, severity);

CREATE TABLE finding_evidence (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    finding_id UUID NOT NULL REFERENCES findings(id) ON DELETE CASCADE,
    evidence_type VARCHAR(64) NOT NULL,
    protected_reference VARCHAR(512) NOT NULL,
    checksum VARCHAR(64) NOT NULL,
    redaction_status VARCHAR(32) NOT NULL DEFAULT 'REDACTED',
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);

CREATE TABLE remediations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    finding_id UUID NOT NULL REFERENCES findings(id) ON DELETE CASCADE,
    assigned_to UUID REFERENCES users(id),
    description TEXT NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'OPEN',
    due_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);

-- ============================================================================
-- 9. Reporting Context
-- ============================================================================

CREATE TABLE reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    requested_by UUID NOT NULL REFERENCES users(id),
    format VARCHAR(16) NOT NULL, -- 'HTML', 'JSON', 'PDF'
    status VARCHAR(32) NOT NULL DEFAULT 'PENDING',
    file_reference VARCHAR(512) NOT NULL,
    checksum VARCHAR(64) NOT NULL,
    snapshot_version VARCHAR(32) NOT NULL DEFAULT 'v1.0.0',
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp(),
    completed_at TIMESTAMPTZ
);

CREATE INDEX idx_reports_project ON reports (project_id, created_at DESC);

-- ============================================================================
-- 10. Audit Context
-- ============================================================================

CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID REFERENCES workspaces(id) ON DELETE SET NULL,
    actor_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(128) NOT NULL,
    resource_type VARCHAR(64) NOT NULL,
    resource_id UUID,
    correlation_id UUID NOT NULL,
    metadata_redacted JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT clock_timestamp()
);

CREATE INDEX idx_audit_logs_workspace_time ON audit_logs (workspace_id, created_at DESC);

-- ============================================================================
-- 11. Initial Baseline Seed Data
-- ============================================================================

-- Default Administrator (Username: admin or admin@vulnex.local, Password: admin)
INSERT INTO users (id, email_normalized, password_hash, display_name, status)
VALUES (
    'd5f7f5ab-773c-4e6a-b7c1-8f51f49704c9',
    'admin@vulnex.local',
    crypt('admin', gen_salt('bf', 10)),
    'Administrator',
    'ACTIVE'
) ON CONFLICT (id) DO UPDATE SET password_hash = EXCLUDED.password_hash;

-- Default Workspace
INSERT INTO workspaces (id, name, owner_user_id, allow_private_network_scan, status)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'Default Workspace',
    'd5f7f5ab-773c-4e6a-b7c1-8f51f49704c9',
    true,
    'ACTIVE'
) ON CONFLICT (id) DO NOTHING;

-- Default Workspace Membership
INSERT INTO workspace_members (workspace_id, user_id, role)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'd5f7f5ab-773c-4e6a-b7c1-8f51f49704c9',
    'OWNER'
) ON CONFLICT (workspace_id, user_id) DO NOTHING;

-- Default Project
INSERT INTO projects (id, workspace_id, name, description, status, created_by)
VALUES (
    '00000000-0000-0000-0000-000000000002',
    '00000000-0000-0000-0000-000000000001',
    'Primary Infrastructure',
    'Default vulnerability assessment and penetration testing scope',
    'ACTIVE',
    'd5f7f5ab-773c-4e6a-b7c1-8f51f49704c9'
) ON CONFLICT (id) DO NOTHING;

-- Default Scanner Profile & Version
INSERT INTO scanner_profiles (id, name, scanner_type, description, status)
VALUES (
    '00000000-0000-0000-0000-000000000004',
    'nmap-top-ports',
    'NMAP',
    'Nmap — Top Ports Discovery (TCP)',
    'ACTIVE'
) ON CONFLICT (id) DO NOTHING;

INSERT INTO scanner_profile_versions (id, scanner_profile_id, version, configuration, approved_by, approved_at, checksum, status)
VALUES (
    '00000000-0000-0000-0000-000000000004',
    '00000000-0000-0000-0000-000000000004',
    'v1.0.0',
    '{"command": "nmap", "args": ["--top-ports", "1000"]}'::jsonb,
    'd5f7f5ab-773c-4e6a-b7c1-8f51f49704c9',
    '2024-01-01 00:00:00Z',
    'e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855',
    'APPROVED'
) ON CONFLICT (id) DO NOTHING;

-- Default Global Authorization & Scopes (allows testing against firing-range, scanme, localhost, and internet targets)
INSERT INTO authorizations (
    id,
    project_id,
    status,
    authorized_party,
    approver_user_id,
    evidence_reference,
    evidence_checksum,
    valid_from,
    valid_until,
    allowed_methods
) VALUES (
    '00000000-0000-0000-0000-000000000003',
    '00000000-0000-0000-0000-000000000002',
    'ACTIVE',
    'Default SecOps Team',
    'd5f7f5ab-773c-4e6a-b7c1-8f51f49704c9',
    'AUTH-DEFAULT-BASELINE',
    'e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855',
    '2020-01-01 00:00:00Z',
    '2099-12-31 23:59:59Z',
    '["PORT_SCAN", "VULN_ASSESSMENT"]'::jsonb
) ON CONFLICT (id) DO NOTHING;

INSERT INTO authorization_scopes (
    authorization_id,
    rule_type,
    target_type,
    canonical_target,
    notes
) VALUES
    ('00000000-0000-0000-0000-000000000003', 'INCLUDE', 'DOMAIN', 'public-firing-range.appspot.com', 'Google Firing Range test bed'),
    ('00000000-0000-0000-0000-000000000003', 'INCLUDE', 'DOMAIN', 'scanme.nmap.org', 'Nmap test host'),
    ('00000000-0000-0000-0000-000000000003', 'INCLUDE', 'DOMAIN', '*', 'Global wildcard domain scope'),
    ('00000000-0000-0000-0000-000000000003', 'INCLUDE', 'CIDR', '0.0.0.0/0', 'Global IPv4 testing scope'),
    ('00000000-0000-0000-0000-000000000003', 'INCLUDE', 'URL', 'http', 'Global HTTP scope'),
    ('00000000-0000-0000-0000-000000000003', 'INCLUDE', 'URL', 'https', 'Global HTTPS scope')
ON CONFLICT (id) DO NOTHING;
