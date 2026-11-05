# Installer Domain Design

## 1. Aggregates

### Installation (Aggregate Root)
Owns installation identity, selected mode, lifecycle state, version, and configured data paths.

Invariants:
- Exactly one mode: Local or Server.
- State transitions are explicit.
- Ready requires deployment health verification.
- Existing persistent data is not deleted by upgrade.

### Deployment (Aggregate Root)
Owns component release versions, deployment profile, and readiness state.

Invariants:
- All components must satisfy release compatibility.
- Only verified image digests may be deployed.
- Server profile must satisfy remote access security policy.

### Backup (Aggregate Root)
Owns backup metadata, source installation identity, version, checksum, and completion status.

Invariants:
- Restore validates integrity and compatibility.
- A backup cannot be marked valid before checksum verification.

## 2. Entities and value objects

Entities:
- Installation
- Deployment
- ComponentRelease
- Backup
- UpgradeExecution

Value objects:
- InstallationID
- DeploymentMode
- InstallationPath
- ApplicationVersion
- ImageDigest
- RuntimeVersion
- ResourceRequirement
- DomainName
- BackupChecksum
- InstallationStatus

## 3. Domain services

- `DeploymentCompatibilityPolicy`
- `InstallationReadinessPolicy`
- `UpgradeCompatibilityPolicy`
- `DataRetentionPolicy`
- `RuntimeRequirementPolicy`

## 4. Domain events

- `InstallationPlanned`
- `InstallationStarted`
- `InstallationCompleted`
- `InstallationFailed`
- `DeploymentStarted`
- `DeploymentReady`
- `BackupCreated`
- `RestoreCompleted`
- `UpgradeStarted`
- `UpgradeCompleted`
- `UninstallationCompleted`

## 5. Repository and service ports

```go
type InstallationRepository interface {
    Get(ctx context.Context, id InstallationID) (Installation, error)
    Save(ctx context.Context, installation Installation) error
}

type RuntimePort interface {
    Inspect(ctx context.Context) (RuntimeInfo, error)
    Apply(ctx context.Context, spec DeploymentSpec) error
    Stop(ctx context.Context, name string) error
    Status(ctx context.Context, name string) (DeploymentStatus, error)
}

type ArtifactVerifier interface {
    Verify(ctx context.Context, artifact ReleaseArtifact) error
}

type BackupPort interface {
    Create(ctx context.Context, destination string) (BackupMetadata, error)
    Restore(ctx context.Context, source string) error
}
```

Interfaces are illustrative contracts; refine method signatures and error types during implementation. Domain types must not include JSON or database-specific tags.
