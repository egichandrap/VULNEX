package domain

import (
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrInvalidValidityPeriod        = errors.New("validity period is invalid: validUntil must be after validFrom")
	ErrEmptyAuthorizedParty         = errors.New("authorized party cannot be empty")
	ErrEmptyEvidenceReference       = errors.New("evidence reference cannot be empty")
	ErrEmptyEvidenceChecksum        = errors.New("evidence checksum cannot be empty")
	ErrAuthorizationAlreadyRevoked  = errors.New("authorization has already been revoked and cannot be revoked again")
)

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

type AuthorizationStatus string

const (
	AuthStatusActive  AuthorizationStatus = "ACTIVE"
	AuthStatusRevoked AuthorizationStatus = "REVOKED"
)

type ScopeRuleType string

const (
	RuleTypeInclude ScopeRuleType = "INCLUDE"
	RuleTypeExclude ScopeRuleType = "EXCLUDE"
)

type ScopeTargetType string

const (
	ScopeTargetDomain ScopeTargetType = "DOMAIN"
	ScopeTargetIP     ScopeTargetType = "IP"
	ScopeTargetCIDR   ScopeTargetType = "CIDR"
	ScopeTargetURL    ScopeTargetType = "URL"
)

type ScopeRule struct {
	ID              string
	RuleType        ScopeRuleType
	TargetType      ScopeTargetType
	CanonicalTarget string
	PortRange       string
	Notes           string
}

// Authorization is the Aggregate Root governing legal and operational authorization.
type Authorization struct {
	id                AuthorizationID
	projectID         ProjectID
	status            AuthorizationStatus
	authorizedParty   string
	approverUserID    UserID
	evidenceReference string
	evidenceChecksum  string
	validFrom         time.Time
	validUntil        time.Time
	allowedMethods    []string
	revokedAt         *time.Time
	revocationReason  string
	scopeRules        []ScopeRule
	createdAt         time.Time
}

func NewAuthorization(
	id AuthorizationID,
	projectID ProjectID,
	authorizedParty string,
	approverID UserID,
	evidenceRef string,
	evidenceChecksum string,
	validFrom time.Time,
	validUntil time.Time,
	allowedMethods []string,
) (*Authorization, error) {
	if strings.TrimSpace(authorizedParty) == "" {
		return nil, ErrEmptyAuthorizedParty
	}
	if strings.TrimSpace(evidenceRef) == "" {
		return nil, ErrEmptyEvidenceReference
	}
	if strings.TrimSpace(evidenceChecksum) == "" {
		return nil, ErrEmptyEvidenceChecksum
	}
	if !validUntil.After(validFrom) {
		return nil, ErrInvalidValidityPeriod
	}

	return &Authorization{
		id:                id,
		projectID:         projectID,
		status:            AuthStatusActive,
		authorizedParty:   strings.TrimSpace(authorizedParty),
		approverUserID:    approverID,
		evidenceReference: strings.TrimSpace(evidenceRef),
		evidenceChecksum:  strings.TrimSpace(evidenceChecksum),
		validFrom:         validFrom.UTC(),
		validUntil:        validUntil.UTC(),
		allowedMethods:    allowedMethods,
		revokedAt:         nil,
		scopeRules:        make([]ScopeRule, 0),
		createdAt:         time.Now().UTC(),
	}, nil
}

func (a *Authorization) ID() AuthorizationID {
	return a.id
}

func (a *Authorization) ProjectID() ProjectID {
	return a.projectID
}

func (a *Authorization) Status() AuthorizationStatus {
	return a.status
}

func (a *Authorization) ValidFrom() time.Time {
	return a.validFrom
}

func (a *Authorization) ValidUntil() time.Time {
	return a.validUntil
}

func (a *Authorization) IsValidAt(at time.Time) bool {
	if a.status != AuthStatusActive {
		return false
	}
	utc := at.UTC()
	return !utc.Before(a.validFrom) && !utc.After(a.validUntil)
}

func (a *Authorization) Revoke(reason string) error {
	if a.status == AuthStatusRevoked {
		return ErrAuthorizationAlreadyRevoked
	}
	a.status = AuthStatusRevoked
	now := time.Now().UTC()
	a.revokedAt = &now
	a.revocationReason = strings.TrimSpace(reason)
	return nil
}

func (a *Authorization) AddScopeRule(rule ScopeRule) {
	a.scopeRules = append(a.scopeRules, rule)
}

func (a *Authorization) ScopeRules() []ScopeRule {
	res := make([]ScopeRule, len(a.scopeRules))
	copy(res, a.scopeRules)
	return res
}
