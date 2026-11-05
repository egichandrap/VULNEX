# Domain Design — DDD

## 1. Bounded contexts

| Context | Responsibility | Core aggregates |
|---|---|---|
| Identity | Users, credentials, sessions | User |
| Workspace | Membership and workspace policy | Workspace |
| Project | Assessment project lifecycle | Project |
| Asset | Asset inventory and identity | Asset |
| Authorization | Written authorization and scope | Authorization |
| Scanning | Scan request and lifecycle | Scan |
| Scanner Catalog | Approved profiles and adapter metadata | ScannerProfile |
| Findings | Normalized observations and triage | Finding |
| Vulnerability Intelligence | CVE/CWE/CVSS reference data | Vulnerability |
| Risk | Explainable risk assessment | RiskAssessment |
| Reporting | Generated reports and export lifecycle | Report |
| Audit | Append-only security activity | AuditRecord |

Avoid a single aggregate spanning multiple contexts. Cross-context consistency is coordinated by application use cases and domain events.

## 2. Aggregates and invariants

### Workspace
- Has one owner.
- Membership is unique per user/workspace.
- Cannot remove the last owner.
- All child projects belong to the workspace.

### Project
- Belongs to exactly one workspace.
- Has lifecycle status and project-level policy.
- Cannot reference assets from another project.

### Asset
- Has canonical type and value.
- Belongs to one project.
- Duplicate canonical assets are prevented within the project unless explicitly modeled as aliases.

### Authorization
- References project and evidence metadata.
- Validity period must be coherent.
- Includes allowed targets/methods and exclusions.
- Revocation is terminal for that authorization record; create a new record for renewed approval.

### Scan
- References project, authorization, immutable scanner profile version, and target snapshot.
- State transitions are guarded.
- Terminal states cannot transition back to active.
- A scan cannot start without successful execution-time policy validation.

### Finding
- Has a stable fingerprint, source provenance, affected asset, scan, severity, confidence, and status.
- Status transitions are controlled.
- Evidence references are access-controlled and bounded.

### Report
- Captures immutable report input snapshot/version.
- Export access is authorized and audited.

## 3. Entities

User, WorkspaceMember, Project, Asset, AuthorizationScope, ScanTarget, ScannerProfileVersion, ScanJob, FindingEvidence, Remediation, AuditActor.

## 4. Value objects

UserID, WorkspaceID, ProjectID, AssetID, AuthorizationID, ScanID, FindingID, ReportID, EmailAddress, AssetType, CanonicalAssetValue, ScopeRule, AuthorizationPeriod, ScanStatus, FindingStatus, Severity, Confidence, CVSSScore, RiskScore, ScannerType, ScanTimeout, ResourceLimit, FindingFingerprint.

Value objects validate at construction and expose immutable behavior.

## 5. Domain events

UserRegistered, WorkspaceCreated, MemberInvited, MemberRoleChanged, ProjectCreated, AssetRegistered, AuthorizationRecorded, AuthorizationRevoked, ScanRequested, ScanQueued, ScanStarted, ScanCompleted, ScanFailed, ScanCancelled, FindingDiscovered, FindingStatusChanged, RiskAssessed, ReportGenerated.

Use transactional outbox for events that must be delivered reliably outside the owning transaction.

## 6. Repository ports

Interfaces belong to the owning domain/application boundary. Example:

```go
type ScanRepository interface {
    Save(ctx context.Context, scan *Scan) error
    Get(ctx context.Context, id ScanID) (*Scan, error)
    Update(ctx context.Context, scan *Scan) error
}
```

```go
type AuthorizationRepository interface {
    Get(ctx context.Context, id AuthorizationID) (*Authorization, error)
    IsValidForExecution(ctx context.Context, id AuthorizationID, at time.Time) (bool, error)
}
```

```go
type FindingRepository interface {
    Save(ctx context.Context, finding *Finding) error
    FindByFingerprint(ctx context.Context, projectID ProjectID, fingerprint FindingFingerprint) (*Finding, error)
}
```

Infrastructure adapters implement these interfaces. Domain types must not import SQL drivers, ORM, HTTP framework, scanner SDKs, or cloud SDKs.

## 7. Application use cases

Identity: RegisterUser, AuthenticateUser, RevokeSession.
Workspace: CreateWorkspace, InviteMember, ChangeMemberRole.
Project: CreateProject, ArchiveProject.
Asset: RegisterAsset, UpdateAsset, ListProjectAssets.
Authorization: RecordAuthorization, RevokeAuthorization, ValidateAuthorization.
Scanning: RequestScan, ValidateScanRequest, QueueScan, StartScan, CancelScan, CompleteScan, FailScan.
Scanner catalog: PublishProfile, GetApprovedProfile.
Findings: IngestScannerResult, NormalizeFinding, DeduplicateFinding, ChangeFindingStatus.
Risk: AssessFindingRisk.
Reporting: GenerateReport, ExportReport.

## 8. Domain service boundaries

ScopePolicy evaluates canonical target membership, exclusions, authorization method, and policy context. NetworkDestinationPolicy validates resolved addresses at the connection boundary. RiskScoringService calculates explainable risk using a versioned model. These services must not perform network I/O or persistence directly.

## 9. Dependency rules

Delivery -> Application -> Domain.
Infrastructure -> Application ports / Domain interfaces.
Domain -> no outer-layer dependency.

Use DTOs at application/delivery boundaries. Do not expose persistence models as domain entities. Do not place business rules in HTTP handlers or repositories.
