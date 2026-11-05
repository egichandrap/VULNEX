package domain

import (
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrEmptyWorkspaceName       = errors.New("workspace name cannot be empty")
	ErrWorkspaceNameTooLong     = errors.New("workspace name exceeds maximum length of 120 characters")
	ErrMemberAlreadyExists      = errors.New("user is already a member of this workspace")
	ErrMemberNotFound           = errors.New("member not found in workspace")
	ErrCannotRemoveSoleOwner    = errors.New("cannot remove the sole owner of the workspace")
	ErrUnauthorizedPolicyChange = errors.New("only workspace owners or administrators may alter network policies")
)

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

// UserID represents an identity reference within the workspace context.
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

func generateUUID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

type Role string

const (
	RoleOwner  Role = "OWNER"
	RoleAdmin  Role = "ADMIN"
	RoleMember Role = "MEMBER"
	RoleViewer Role = "VIEWER"
)

type WorkspaceMember struct {
	id        string
	userID    UserID
	role      Role
	createdAt time.Time
}

func ReconstituteWorkspaceMember(id string, userID UserID, role Role, createdAt time.Time) WorkspaceMember {
	return WorkspaceMember{
		id:        id,
		userID:    userID,
		role:      role,
		createdAt: createdAt,
	}
}

func (m *WorkspaceMember) ID() string {
	return m.id
}

func (m *WorkspaceMember) UserID() UserID {
	return m.userID
}

func (m *WorkspaceMember) Role() Role {
	return m.role
}

func (m *WorkspaceMember) CreatedAt() time.Time {
	return m.createdAt
}

type WorkspaceStatus string

const (
	WorkspaceStatusActive    WorkspaceStatus = "ACTIVE"
	WorkspaceStatusSuspended WorkspaceStatus = "SUSPENDED"
)

// Workspace is the Aggregate Root governing multi-tenant assessment ownership and policies.
type Workspace struct {
	id                      WorkspaceID
	name                    string
	ownerUserID             UserID
	allowPrivateNetworkScan bool // ADR-015
	status                  WorkspaceStatus
	members                 []WorkspaceMember
	createdAt               time.Time
	updatedAt               time.Time
}

func NewWorkspace(name string, ownerID UserID) (*Workspace, error) {
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return nil, ErrEmptyWorkspaceName
	}
	if len(trimmedName) > 120 {
		return nil, ErrWorkspaceNameTooLong
	}

	id, err := GenerateWorkspaceID()
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	ownerMember := WorkspaceMember{
		id:        generateUUID(),
		userID:    ownerID,
		role:      RoleOwner,
		createdAt: now,
	}

	return &Workspace{
		id:                      id,
		name:                    trimmedName,
		ownerUserID:             ownerID,
		allowPrivateNetworkScan: false, // Default: false (ADR-015)
		status:                  WorkspaceStatusActive,
		members:                 []WorkspaceMember{ownerMember},
		createdAt:               now,
		updatedAt:               now,
	}, nil
}

func (w *Workspace) ID() WorkspaceID {
	return w.id
}

func (w *Workspace) CreatedAt() time.Time {
	return w.createdAt
}

func (w *Workspace) UpdatedAt() time.Time {
	return w.updatedAt
}

func ReconstituteWorkspace(
	id WorkspaceID,
	name string,
	ownerUserID UserID,
	allowPrivateNetworkScan bool,
	status WorkspaceStatus,
	members []WorkspaceMember,
	createdAt time.Time,
	updatedAt time.Time,
) *Workspace {
	return &Workspace{
		id:                      id,
		name:                    name,
		ownerUserID:             ownerUserID,
		allowPrivateNetworkScan: allowPrivateNetworkScan,
		status:                  status,
		members:                 members,
		createdAt:               createdAt,
		updatedAt:               updatedAt,
	}
}

func (w *Workspace) Name() string {
	return w.name
}

func (w *Workspace) OwnerUserID() UserID {
	return w.ownerUserID
}

func (w *Workspace) AllowPrivateNetworkScan() bool {
	return w.allowPrivateNetworkScan
}

func (w *Workspace) Status() WorkspaceStatus {
	return w.status
}

func (w *Workspace) Members() []WorkspaceMember {
	res := make([]WorkspaceMember, len(w.members))
	copy(res, w.members)
	return res
}

func (w *Workspace) AddMember(userID UserID, role Role) error {
	for _, m := range w.members {
		if m.userID.String() == userID.String() {
			return ErrMemberAlreadyExists
		}
	}

	now := time.Now().UTC()
	w.members = append(w.members, WorkspaceMember{
		id:        generateUUID(),
		userID:    userID,
		role:      role,
		createdAt: now,
	})
	w.updatedAt = now
	return nil
}

func (w *Workspace) RemoveMember(userID UserID) error {
	idx := -1
	for i, m := range w.members {
		if m.userID.String() == userID.String() {
			idx = i
			break
		}
	}

	if idx == -1 {
		return ErrMemberNotFound
	}

	// Invariant: Cannot remove sole owner
	if w.members[idx].role == RoleOwner {
		ownerCount := 0
		for _, m := range w.members {
			if m.role == RoleOwner {
				ownerCount++
			}
		}
		if ownerCount <= 1 {
			return ErrCannotRemoveSoleOwner
		}
	}

	w.members = append(w.members[:idx], w.members[idx+1:]...)
	w.updatedAt = time.Now().UTC()
	return nil
}

func (w *Workspace) EnablePrivateNetworkScanning(actorUserID UserID, justification string) error {
	if strings.TrimSpace(justification) == "" {
		return errors.New("justification is required to enable private network scanning")
	}

	actorRole, ok := w.getRole(actorUserID)
	if !ok || (actorRole != RoleOwner && actorRole != RoleAdmin) {
		return ErrUnauthorizedPolicyChange
	}

	w.allowPrivateNetworkScan = true
	w.updatedAt = time.Now().UTC()
	return nil
}

func (w *Workspace) DisablePrivateNetworkScanning() {
	w.allowPrivateNetworkScan = false
	w.updatedAt = time.Now().UTC()
}

func (w *Workspace) getRole(userID UserID) (Role, bool) {
	for _, m := range w.members {
		if m.userID.String() == userID.String() {
			return m.role, true
		}
	}
	return "", false
}
