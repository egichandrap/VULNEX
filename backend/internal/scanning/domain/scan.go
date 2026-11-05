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
