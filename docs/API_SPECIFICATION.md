# API Specification — Baseline

Base path: `/api/v1`. JSON over HTTPS. OpenAPI 3.x is the contract source.

## 1. Conventions
- UUID identifiers.
- UTC ISO-8601 timestamps.
- Pagination: `limit`, `cursor`.
- Idempotency-Key for scan creation and report generation where applicable.
- Consistent error envelope.
- Authentication via secure session or bearer token; choose one primary browser strategy during implementation and document CSRF/session protections.
- All resource operations enforce workspace and object-level authorization.

## 2. Endpoints

| Method | Path | Purpose |
|---|---|---|
| POST | `/auth/setup` | Initial admin onboarding with one-time setup token (ADR-013) |
| POST | `/auth/register` | Register (disabled if self-registration closed) |
| POST | `/auth/login` | Login |
| POST | `/auth/logout` | Logout |
| GET | `/me` | Current user |
| GET/POST | `/workspaces` | List/create workspaces |
| GET/PATCH | `/workspaces/{workspaceId}` | Workspace details/update (includes private scan policy, ADR-015) |
| GET/POST | `/workspaces/{workspaceId}/members` | List/invite members |
| PATCH/DELETE | `/workspaces/{workspaceId}/members/{memberId}` | Change/remove membership |
| GET/POST | `/workspaces/{workspaceId}/projects` | List/create projects |
| GET/PATCH | `/projects/{projectId}` | Project detail/update |
| GET/POST | `/projects/{projectId}/assets` | List/register assets |
| GET/PATCH | `/assets/{assetId}` | Asset detail/update |
| GET/POST | `/projects/{projectId}/authorizations` | List/record authorization |
| POST | `/authorizations/{authorizationId}/revoke` | Revoke authorization |
| GET | `/scanner-profiles` | List published profiles |
| POST | `/projects/{projectId}/scans` | Request scan |
| GET | `/scans/{scanId}` | Scan detail/status |
| POST | `/scans/{scanId}/cancel` | Cancel scan |
| GET | `/scans/{scanId}/findings` | Scan findings |
| GET | `/projects/{projectId}/findings` | Project findings |
| GET/PATCH | `/findings/{findingId}` | Finding detail/update |
| POST | `/findings/{findingId}/remediations` | Add remediation |
| POST | `/projects/{projectId}/reports` | Generate report |
| GET | `/reports/{reportId}` | Report status/detail |
| GET | `/audit-logs` | Authorized audit search |

## 3. Scan request

```json
{
  "authorization_id": "uuid",
  "scanner_profile_version_id": "uuid",
  "asset_ids": ["uuid"]
}
```

Do not accept raw command strings, arbitrary executable paths, unrestricted flags, or unregistered arbitrary destinations. Server derives target values from registered assets and validates authorization/scope.

## 4. Scan response

```json
{
  "data": {
    "scan_id": "uuid",
    "status": "QUEUED",
    "created_at": "2026-11-05T08:00:00Z"
  }
}
```

## 5. Error envelope

```json
{
  "error": {
    "code": "SCAN_SCOPE_VIOLATION",
    "message": "One or more targets are outside the authorized scope",
    "correlation_id": "uuid"
  }
}
```

Suggested codes: `AUTHENTICATION_REQUIRED`, `PERMISSION_DENIED`, `RESOURCE_NOT_FOUND`, `VALIDATION_FAILED`, `AUTHORIZATION_REQUIRED`, `AUTHORIZATION_EXPIRED`, `SCAN_SCOPE_VIOLATION`, `NETWORK_POLICY_DENIED`, `PROFILE_NOT_AVAILABLE`, `SCAN_ALREADY_TERMINAL`, `RATE_LIMITED`, `INTERNAL_ERROR`.

## 6. API requirements
- Validate content type and request size.
- Apply rate limits.
- Do not reveal cross-tenant resource existence.
- Use optimistic concurrency/version checks for mutable high-risk resources where appropriate.
- Generate OpenAPI examples and contract tests.
- API responses must not expose secrets, raw credentials, or unredacted restricted evidence.
