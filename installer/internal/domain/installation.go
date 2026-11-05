package domain

import (
	"time"
)

type InstallationStatus string

const (
	InstallationStatusPending    InstallationStatus = "pending"
	InstallationStatusInstalling InstallationStatus = "installing"
	InstallationStatusReady      InstallationStatus = "ready"
	InstallationStatusFailed     InstallationStatus = "failed"
	InstallationStatusStopped    InstallationStatus = "stopped"
	InstallationStatusUpdating   InstallationStatus = "updating"
)

type DeploymentPhase string

const (
	DeploymentPhasePreflight       DeploymentPhase = "preflight"
	DeploymentPhaseConfiguration   DeploymentPhase = "configuration"
	DeploymentPhaseMaterialization DeploymentPhase = "materialization"
	DeploymentPhaseDeployment      DeploymentPhase = "deployment"
	DeploymentPhaseHealthCheck     DeploymentPhase = "health_check"
	DeploymentPhaseCommitted       DeploymentPhase = "committed"
	DeploymentPhaseRolledBack      DeploymentPhase = "rolled_back"
)

// InstallationState models durable state across installation phases (ADR-012).
type InstallationState struct {
	ID          string             `json:"id"`
	Mode        DeploymentMode     `json:"mode"`
	Version     string             `json:"version"`
	Status      InstallationStatus `json:"status"`
	Phase       DeploymentPhase    `json:"phase"`
	DataDir     string             `json:"data_dir"`
	ConfigPath  string             `json:"config_path,omitempty"`
	InstalledAt time.Time          `json:"installed_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
	Services    map[string]string  `json:"services,omitempty"`
	LastError   string             `json:"last_error,omitempty"`
}

// NewInstallationState initializes a pending deployment state record.
func NewInstallationState(mode DeploymentMode, version string, dataDir string) *InstallationState {
	now := time.Now().UTC()
	return &InstallationState{
		ID:          "vulnex-" + string(mode),
		Mode:        mode,
		Version:     version,
		Status:      InstallationStatusPending,
		Phase:       DeploymentPhasePreflight,
		DataDir:     dataDir,
		InstalledAt: now,
		UpdatedAt:   now,
		Services:    make(map[string]string),
	}
}

// AdvancePhase transitions state to an active intermediate deployment phase.
func (s *InstallationState) AdvancePhase(phase DeploymentPhase) {
	s.Phase = phase
	s.Status = InstallationStatusInstalling
	s.UpdatedAt = time.Now().UTC()
}

// MarkReady transitions state to committed healthy state.
func (s *InstallationState) MarkReady() {
	s.Status = InstallationStatusReady
	s.Phase = DeploymentPhaseCommitted
	s.LastError = ""
	s.UpdatedAt = time.Now().UTC()
}

// MarkFailed transitions state to failed with recorded error.
func (s *InstallationState) MarkFailed(reason string) {
	s.Status = InstallationStatusFailed
	s.LastError = reason
	s.UpdatedAt = time.Now().UTC()
}

// SetServices updates component statuses.
func (s *InstallationState) SetServices(services map[string]string) {
	s.Services = services
	s.UpdatedAt = time.Now().UTC()
}
