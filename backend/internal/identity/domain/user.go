package domain

import (
	"crypto/rand"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"
)

var (
	ErrInvalidEmail         = errors.New("invalid email address format")
	ErrEmptyDisplayName     = errors.New("display name cannot be empty")
	ErrEmptyPasswordHash    = errors.New("password hash cannot be empty")
	ErrUserAlreadySuspended = errors.New("user is already suspended")
	ErrUserAlreadyActive    = errors.New("user is already active")
)

// UserID is a unique identifier value object for a User.
type UserID struct {
	value string
}

func NewUserID(val string) (UserID, error) {
	if val == "" {
		return UserID{}, errors.New("user ID cannot be empty")
	}
	return UserID{value: val}, nil
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

// Email is a validated and normalized email address value object.
type Email struct {
	value string
}

func NewEmail(raw string) (Email, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return Email{}, ErrInvalidEmail
	}

	addr, err := mail.ParseAddress(trimmed)
	if err != nil || addr.Address != trimmed {
		// Strict check: reject partial matches or name wrappers
		return Email{}, ErrInvalidEmail
	}

	parts := strings.Split(addr.Address, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" || !strings.Contains(parts[1], ".") {
		return Email{}, ErrInvalidEmail
	}

	normalized := strings.ToLower(addr.Address)
	return Email{value: normalized}, nil
}

func (e Email) String() string {
	return e.value
}

type UserStatus string

const (
	UserStatusActive    UserStatus = "ACTIVE"
	UserStatusSuspended UserStatus = "SUSPENDED"
)

// User is the Aggregate Root for identity and credentials.
type User struct {
	id           UserID
	email        Email
	passwordHash string
	displayName  string
	status       UserStatus
	createdAt    time.Time
	updatedAt    time.Time
}

// NewUser creates a new active User aggregate.
func NewUser(email Email, passwordHash, displayName string) (*User, error) {
	if strings.TrimSpace(displayName) == "" {
		return nil, ErrEmptyDisplayName
	}
	if strings.TrimSpace(passwordHash) == "" {
		return nil, ErrEmptyPasswordHash
	}

	id, err := GenerateUserID()
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	return &User{
		id:           id,
		email:        email,
		passwordHash: passwordHash,
		displayName:  strings.TrimSpace(displayName),
		status:       UserStatusActive,
		createdAt:    now,
		updatedAt:    now,
	}, nil
}

// ReconstituteUser reconstitutes a User from persistence.
func ReconstituteUser(id UserID, email Email, passwordHash, displayName string, status UserStatus, createdAt, updatedAt time.Time) *User {
	return &User{
		id:           id,
		email:        email,
		passwordHash: passwordHash,
		displayName:  displayName,
		status:       status,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}
}

func (u *User) ID() UserID {
	return u.id
}

func (u *User) Email() Email {
	return u.email
}

func (u *User) PasswordHash() string {
	return u.passwordHash
}

func (u *User) DisplayName() string {
	return u.displayName
}

func (u *User) Status() UserStatus {
	return u.status
}

func (u *User) CreatedAt() time.Time {
	return u.createdAt
}

func (u *User) UpdatedAt() time.Time {
	return u.updatedAt
}

func (u *User) Suspend() {
	u.status = UserStatusSuspended
	u.updatedAt = time.Now().UTC()
}

func (u *User) Activate() {
	u.status = UserStatusActive
	u.updatedAt = time.Now().UTC()
}

func (u *User) ChangePassword(newHash string) error {
	if strings.TrimSpace(newHash) == "" {
		return ErrEmptyPasswordHash
	}
	u.passwordHash = newHash
	u.updatedAt = time.Now().UTC()
	return nil
}
