package domain

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrEmptyFindingTitle       = errors.New("finding title cannot be empty")
	ErrEmptyFindingDescription = errors.New("finding description cannot be empty")
	ErrEmptyJustification      = errors.New("justification is mandatory when accepting risk or marking false positive")
)

type FindingID struct {
	value string
}

func NewFindingID(val string) (FindingID, error) {
	if strings.TrimSpace(val) == "" {
		return FindingID{}, errors.New("finding ID cannot be empty")
	}
	return FindingID{value: strings.TrimSpace(val)}, nil
}

func GenerateFindingID() (FindingID, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return FindingID{}, fmt.Errorf("failed to generate UUID: %w", err)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	val := fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
	return FindingID{value: val}, nil
}

func (f FindingID) String() string {
	return f.value
}

type WorkspaceID struct{ value string }

func NewWorkspaceID(val string) (WorkspaceID, error) {
	if strings.TrimSpace(val) == "" {
		return WorkspaceID{}, errors.New("workspace ID cannot be empty")
	}
	return WorkspaceID{value: strings.TrimSpace(val)}, nil
}

func (w WorkspaceID) String() string { return w.value }
func GenerateWorkspaceID() (WorkspaceID, error) {
	var b [16]byte
	rand.Read(b[:])
	return WorkspaceID{value: fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])}, nil
}

type ProjectID struct{ value string }

func NewProjectID(val string) (ProjectID, error) {
	if strings.TrimSpace(val) == "" {
		return ProjectID{}, errors.New("project ID cannot be empty")
	}
	return ProjectID{value: strings.TrimSpace(val)}, nil
}

func (p ProjectID) String() string { return p.value }
func GenerateProjectID() (ProjectID, error) {
	var b [16]byte
	rand.Read(b[:])
	return ProjectID{value: fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])}, nil
}

type ScanID struct{ value string }

func NewScanID(val string) (ScanID, error) {
	if strings.TrimSpace(val) == "" {
		return ScanID{}, errors.New("scan ID cannot be empty")
	}
	return ScanID{value: strings.TrimSpace(val)}, nil
}

func (s ScanID) String() string { return s.value }
func GenerateScanID() (ScanID, error) {
	var b [16]byte
	rand.Read(b[:])
	return ScanID{value: fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])}, nil
}

type AssetID struct{ value string }

func NewAssetID(val string) (AssetID, error) {
	if strings.TrimSpace(val) == "" {
		return AssetID{}, errors.New("asset ID cannot be empty")
	}
	return AssetID{value: strings.TrimSpace(val)}, nil
}

func (a AssetID) String() string { return a.value }
func GenerateAssetID() (AssetID, error) {
	var b [16]byte
	rand.Read(b[:])
	return AssetID{value: fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])}, nil
}

type Severity string

const (
	SeverityInfo     Severity = "INFO"
	SeverityLow      Severity = "LOW"
	SeverityMedium   Severity = "MEDIUM"
	SeverityHigh     Severity = "HIGH"
	SeverityCritical Severity = "CRITICAL"
)

func (s Severity) BaseWeight() float64 {
	switch s {
	case SeverityCritical:
		return 10.0
	case SeverityHigh:
		return 8.0
	case SeverityMedium:
		return 5.0
	case SeverityLow:
		return 2.5
	default:
		return 0.5
	}
}

type Confidence string

const (
	ConfidenceLow       Confidence = "LOW"
	ConfidenceMedium    Confidence = "MEDIUM"
	ConfidenceHigh      Confidence = "HIGH"
	ConfidenceConfirmed Confidence = "CONFIRMED"
)

func (c Confidence) Multiplier() float64 {
	switch c {
	case ConfidenceConfirmed:
		return 1.0
	case ConfidenceHigh:
		return 0.9
	case ConfidenceMedium:
		return 0.7
	default:
		return 0.5
	}
}

type AssetCriticality string

const (
	AssetCriticalityLow      AssetCriticality = "LOW"
	AssetCriticalityMedium   AssetCriticality = "MEDIUM"
	AssetCriticalityHigh     AssetCriticality = "HIGH"
	AssetCriticalityCritical AssetCriticality = "CRITICAL"
)

func (a AssetCriticality) Multiplier() float64 {
	switch a {
	case AssetCriticalityCritical:
		return 1.0
	case AssetCriticalityHigh:
		return 0.85
	case AssetCriticalityMedium:
		return 0.70
	default:
		return 0.50
	}
}

type FindingStatus string

const (
	FindingStatusOpen          FindingStatus = "OPEN"
	FindingStatusInProgress    FindingStatus = "IN_PROGRESS"
	FindingStatusResolved      FindingStatus = "RESOLVED"
	FindingStatusAcceptedRisk  FindingStatus = "ACCEPTED_RISK"
	FindingStatusFalsePositive FindingStatus = "FALSE_POSITIVE"
)

// FindingFingerprint is a deterministic SHA-256 fingerprint for deduplication.
type FindingFingerprint struct {
	value string
}

func NewFindingFingerprint(val string) (FindingFingerprint, error) {
	if strings.TrimSpace(val) == "" {
		return FindingFingerprint{}, errors.New("finding fingerprint cannot be empty")
	}
	return FindingFingerprint{value: strings.TrimSpace(val)}, nil
}

func CalculateFingerprint(projectID, canonicalAsset, scannerType, vulnKey, locationKey string) FindingFingerprint {
	raw := fmt.Sprintf("%s|%s|%s|%s|%s",
		strings.TrimSpace(projectID),
		strings.ToLower(strings.TrimSpace(canonicalAsset)),
		strings.ToLower(strings.TrimSpace(scannerType)),
		strings.ToLower(strings.TrimSpace(vulnKey)),
		strings.ToLower(strings.TrimSpace(locationKey)),
	)
	hash := sha256.Sum256([]byte(raw))
	return FindingFingerprint{value: hex.EncodeToString(hash[:])}
}

func (f FindingFingerprint) String() string {
	return f.value
}

// Finding is the Aggregate Root representing an assessment observation.
type Finding struct {
	id               FindingID
	workspaceID      WorkspaceID
	projectID        ProjectID
	scanID           ScanID
	assetID          AssetID
	fingerprint      FindingFingerprint
	title            string
	description      string
	severity         Severity
	confidence       Confidence
	status           FindingStatus
	sourceScanner    string
	firstSeenAt      time.Time
	lastSeenAt       time.Time
	riskModelVersion string
	riskScore        float64
	justification    string
}

func NewFinding(
	id FindingID,
	workspaceID WorkspaceID,
	projectID ProjectID,
	scanID ScanID,
	assetID AssetID,
	fingerprint FindingFingerprint,
	title string,
	description string,
	severity Severity,
	confidence Confidence,
	sourceScanner string,
	discoveredAt time.Time,
	riskScore float64,
	riskModelVersion string,
) (*Finding, error) {
	if strings.TrimSpace(title) == "" {
		return nil, ErrEmptyFindingTitle
	}
	if strings.TrimSpace(description) == "" {
		return nil, ErrEmptyFindingDescription
	}

	utc := discoveredAt.UTC()
	return &Finding{
		id:               id,
		workspaceID:      workspaceID,
		projectID:        projectID,
		scanID:           scanID,
		assetID:          assetID,
		fingerprint:      fingerprint,
		title:            strings.TrimSpace(title),
		description:      strings.TrimSpace(description),
		severity:         severity,
		confidence:       confidence,
		status:           FindingStatusOpen,
		sourceScanner:    strings.TrimSpace(sourceScanner),
		firstSeenAt:      utc,
		lastSeenAt:       utc,
		riskModelVersion: riskModelVersion,
		riskScore:        riskScore,
	}, nil
}

func (f *Finding) ID() FindingID                   { return f.id }
func (f *Finding) WorkspaceID() WorkspaceID        { return f.workspaceID }
func (f *Finding) ProjectID() ProjectID            { return f.projectID }
func (f *Finding) ScanID() ScanID                  { return f.scanID }
func (f *Finding) AssetID() AssetID                { return f.assetID }
func (f *Finding) Fingerprint() FindingFingerprint { return f.fingerprint }
func (f *Finding) Title() string                   { return f.title }
func (f *Finding) Description() string             { return f.description }
func (f *Finding) Severity() Severity              { return f.severity }
func (f *Finding) Confidence() Confidence          { return f.confidence }
func (f *Finding) Status() FindingStatus           { return f.status }
func (f *Finding) SourceScanner() string           { return f.sourceScanner }
func (f *Finding) FirstSeenAt() time.Time          { return f.firstSeenAt }
func (f *Finding) LastSeenAt() time.Time           { return f.lastSeenAt }
func (f *Finding) RiskModelVersion() string        { return f.riskModelVersion }
func (f *Finding) RiskScore() float64              { return f.riskScore }
func (f *Finding) Justification() string           { return f.justification }

func ReconstituteFinding(
	id FindingID,
	workspaceID WorkspaceID,
	projectID ProjectID,
	scanID ScanID,
	assetID AssetID,
	fingerprint FindingFingerprint,
	title string,
	description string,
	severity Severity,
	confidence Confidence,
	status FindingStatus,
	sourceScanner string,
	firstSeenAt time.Time,
	lastSeenAt time.Time,
	riskModelVersion string,
	riskScore float64,
	justification string,
) *Finding {
	return &Finding{
		id:               id,
		workspaceID:      workspaceID,
		projectID:        projectID,
		scanID:           scanID,
		assetID:          assetID,
		fingerprint:      fingerprint,
		title:            title,
		description:      description,
		severity:         severity,
		confidence:       confidence,
		status:           status,
		sourceScanner:    sourceScanner,
		firstSeenAt:      firstSeenAt,
		lastSeenAt:       lastSeenAt,
		riskModelVersion: riskModelVersion,
		riskScore:        riskScore,
		justification:    justification,
	}
}

func (f *Finding) RecordObservation(scanID ScanID, at time.Time) {
	f.scanID = scanID
	f.lastSeenAt = at.UTC()
}

func (f *Finding) AcceptRisk(justification string, at time.Time) error {
	trimmed := strings.TrimSpace(justification)
	if trimmed == "" {
		return ErrEmptyJustification
	}
	f.status = FindingStatusAcceptedRisk
	f.justification = trimmed
	f.lastSeenAt = at.UTC()
	return nil
}

func (f *Finding) MarkFalsePositive(justification string, at time.Time) error {
	trimmed := strings.TrimSpace(justification)
	if trimmed == "" {
		return ErrEmptyJustification
	}
	f.status = FindingStatusFalsePositive
	f.justification = trimmed
	f.lastSeenAt = at.UTC()
	return nil
}

func (f *Finding) Resolve(at time.Time) error {
	f.status = FindingStatusResolved
	f.lastSeenAt = at.UTC()
	return nil
}

// RiskScoringService calculates explainable risk scores without external I/O.
type RiskScoringService struct{}

func (s RiskScoringService) Calculate(sev Severity, conf Confidence, crit AssetCriticality) float64 {
	base := sev.BaseWeight()
	score := base * conf.Multiplier() * crit.Multiplier()
	if score > 10.0 {
		return 10.0
	}
	if score < 0.0 {
		return 0.0
	}
	return score
}
