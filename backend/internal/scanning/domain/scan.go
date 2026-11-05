package domain

import (
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrScanAlreadyTerminal = errors.New("scan is in a terminal state and cannot be modified")
	ErrInvalidStateTransition = errors.New("invalid scan state transition")
	ErrEmptyScanTargets    = errors.New("scan must have at least one target")
	ErrEmptyIdempotencyKey = errors.New("idempotency key cannot be empty")
)

type ScanID struct {
	value string
}

func NewScanID(val string) (ScanID, error) {
	if strings.TrimSpace(val) == "" {
		return ScanID{}, errors.New("scan ID cannot be empty")
	}
	return ScanID{value: strings.TrimSpace(val)}, nil
}

func GenerateScanID() (ScanID, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return ScanID{}, fmt.Errorf("failed to generate UUID: %w", err)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	val := fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
	return ScanID{value: val}, nil
}

func (s ScanID) String() string {
	return s.value
}

type WorkspaceID struct {
	value string
}

func NewWorkspaceID(val string) (WorkspaceID, error) {
	if strings.TrimSpace(val) == "" {
		return WorkspaceID{}, errors.New("workspace ID cannot be empty")
	}
	return WorkspaceID{value: strings.TrimSpace(val)}, nil
}

func GenerateWorkspaceID() (WorkspaceID, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return WorkspaceID{}, fmt.Errorf("failed to generate UUID: %w", err)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	val := fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
	return WorkspaceID{value: val}, nil
}

func (w WorkspaceID) String() string {
	return w.value
}

type ProjectID struct {
	value string
}

func NewProjectID(val string) (ProjectID, error) {
	if strings.TrimSpace(val) == "" {
		return ProjectID{}, errors.New("project ID cannot be empty")
	}
	return ProjectID{value: strings.TrimSpace(val)}, nil
}

func GenerateProjectID() (ProjectID, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return ProjectID{}, fmt.Errorf("failed to generate UUID: %w", err)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	val := fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
	return ProjectID{value: val}, nil
}

func (p ProjectID) String() string {
	return p.value
}

type AuthorizationID struct {
	value string
}

func NewAuthorizationID(val string) (AuthorizationID, error) {
	if strings.TrimSpace(val) == "" {
		return AuthorizationID{}, errors.New("authorization ID cannot be empty")
	}
	return AuthorizationID{value: strings.TrimSpace(val)}, nil
}

func GenerateAuthorizationID() (AuthorizationID, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return AuthorizationID{}, fmt.Errorf("failed to generate UUID: %w", err)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	val := fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
	return AuthorizationID{value: val}, nil
}

func (a AuthorizationID) String() string {
	return a.value
}

type ProfileVersionID struct {
	value string
}

func NewProfileVersionID(val string) (ProfileVersionID, error) {
	if strings.TrimSpace(val) == "" {
		return ProfileVersionID{}, errors.New("profile version ID cannot be empty")
	}
	return ProfileVersionID{value: strings.TrimSpace(val)}, nil
}

func GenerateProfileVersionID() (ProfileVersionID, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return ProfileVersionID{}, fmt.Errorf("failed to generate UUID: %w", err)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	val := fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
	return ProfileVersionID{value: val}, nil
}

func (p ProfileVersionID) String() string {
	return p.value
}

type UserID struct {
	value string
}

func NewUserID(val string) (UserID, error) {
	if strings.TrimSpace(val) == "" {
		return UserID{}, errors.New("user ID cannot be empty")
	}
	return UserID{value: strings.TrimSpace(val)}, nil
}

func GenerateUserID() (UserID, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return UserID{}, fmt.Errorf("failed to generate UUID: %w", err)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	val := fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
	return UserID{value: val}, nil
}

func (u UserID) String() string {
	return u.value
}

type ScanStatus string

const (
	ScanStatusQueued    ScanStatus = "QUEUED"
	ScanStatusRunning   ScanStatus = "RUNNING"
	ScanStatusCompleted ScanStatus = "COMPLETED"
	ScanStatusFailed    ScanStatus = "FAILED"
	ScanStatusCancelled ScanStatus = "CANCELLED"
	ScanStatusTimedOut  ScanStatus = "TIMED_OUT"
)

// Scan is the Aggregate Root governing an assessment execution.
type Scan struct {
	id                      ScanID
	workspaceID             WorkspaceID
	projectID               ProjectID
	authorizationID         AuthorizationID
	scannerProfileVersionID ProfileVersionID
	requestedBy             UserID
	status                  ScanStatus
	idempotencyKey          string
	targetSnapshot          []string
	startedAt               *time.Time
	completedAt             *time.Time
	failureReason           string
	createdAt               time.Time
	updatedAt               time.Time
}

func NewScan(
	id ScanID,
	workspaceID WorkspaceID,
	projectID ProjectID,
	authID AuthorizationID,
	profileVersionID ProfileVersionID,
	requestedBy UserID,
	idempotencyKey string,
	targets []string,
) (*Scan, error) {
	if len(targets) == 0 {
		return nil, ErrEmptyScanTargets
	}
	if strings.TrimSpace(idempotencyKey) == "" {
		return nil, ErrEmptyIdempotencyKey
	}

	now := time.Now().UTC()
	targetCopy := make([]string, len(targets))
	copy(targetCopy, targets)

	return &Scan{
		id:                      id,
		workspaceID:             workspaceID,
		projectID:               projectID,
		authorizationID:         authID,
		scannerProfileVersionID: profileVersionID,
		requestedBy:             requestedBy,
		status:                  ScanStatusQueued,
		idempotencyKey:          strings.TrimSpace(idempotencyKey),
		targetSnapshot:          targetCopy,
		startedAt:               nil,
		completedAt:             nil,
		createdAt:               now,
		updatedAt:               now,
	}, nil
}

func (s *Scan) ID() ScanID {
	return s.id
}

func (s *Scan) WorkspaceID() WorkspaceID {
	return s.workspaceID
}

func (s *Scan) ProjectID() ProjectID {
	return s.projectID
}

func (s *Scan) AuthorizationID() AuthorizationID {
	return s.authorizationID
}

func (s *Scan) ScannerProfileVersionID() ProfileVersionID {
	return s.scannerProfileVersionID
}

func (s *Scan) RequestedBy() UserID {
	return s.requestedBy
}

func (s *Scan) IdempotencyKey() string {
	return s.idempotencyKey
}

func (s *Scan) TargetSnapshot() []string {
	res := make([]string, len(s.targetSnapshot))
	copy(res, s.targetSnapshot)
	return res
}

func (s *Scan) StartedAt() *time.Time {
	return s.startedAt
}

func (s *Scan) CompletedAt() *time.Time {
	return s.completedAt
}

func (s *Scan) FailureReason() string {
	return s.failureReason
}

func (s *Scan) CreatedAt() time.Time {
	return s.createdAt
}

func (s *Scan) UpdatedAt() time.Time {
	return s.updatedAt
}

func ReconstituteScan(
	id ScanID,
	workspaceID WorkspaceID,
	projectID ProjectID,
	authID AuthorizationID,
	profileVersionID ProfileVersionID,
	requestedBy UserID,
	status ScanStatus,
	idempotencyKey string,
	targetSnapshot []string,
	startedAt *time.Time,
	completedAt *time.Time,
	failureReason string,
	createdAt time.Time,
	updatedAt time.Time,
) *Scan {
	return &Scan{
		id:                      id,
		workspaceID:             workspaceID,
		projectID:               projectID,
		authorizationID:         authID,
		scannerProfileVersionID: profileVersionID,
		requestedBy:             requestedBy,
		status:                  status,
		idempotencyKey:          idempotencyKey,
		targetSnapshot:          targetSnapshot,
		startedAt:               startedAt,
		completedAt:             completedAt,
		failureReason:           failureReason,
		createdAt:               createdAt,
		updatedAt:               updatedAt,
	}
}

func (s *Scan) Status() ScanStatus {
	return s.status
}

func (s *Scan) IsTerminal() bool {
	return s.status == ScanStatusCompleted ||
		s.status == ScanStatusFailed ||
		s.status == ScanStatusCancelled ||
		s.status == ScanStatusTimedOut
}

func (s *Scan) MarkRunning(at time.Time) error {
	if s.IsTerminal() {
		return ErrScanAlreadyTerminal
	}
	if s.status != ScanStatusQueued {
		return ErrInvalidStateTransition
	}
	s.status = ScanStatusRunning
	now := at.UTC()
	s.startedAt = &now
	s.updatedAt = now
	return nil
}

func (s *Scan) MarkCompleted(at time.Time) error {
	if s.IsTerminal() {
		return ErrScanAlreadyTerminal
	}
	if s.status != ScanStatusRunning {
		return ErrInvalidStateTransition
	}
	s.status = ScanStatusCompleted
	now := at.UTC()
	s.completedAt = &now
	s.updatedAt = now
	return nil
}

func (s *Scan) MarkFailed(reason string, at time.Time) error {
	if s.IsTerminal() {
		return ErrScanAlreadyTerminal
	}
	s.status = ScanStatusFailed
	s.failureReason = strings.TrimSpace(reason)
	now := at.UTC()
	s.completedAt = &now
	s.updatedAt = now
	return nil
}

func (s *Scan) MarkTimedOut(at time.Time) error {
	if s.IsTerminal() {
		return ErrScanAlreadyTerminal
	}
	s.status = ScanStatusTimedOut
	now := at.UTC()
	s.completedAt = &now
	s.updatedAt = now
	return nil
}

func (s *Scan) Cancel(reason string) error {
	if s.IsTerminal() {
		return ErrScanAlreadyTerminal
	}
	s.status = ScanStatusCancelled
	s.failureReason = strings.TrimSpace(reason)
	now := time.Now().UTC()
	s.completedAt = &now
	s.updatedAt = now
	return nil
}
