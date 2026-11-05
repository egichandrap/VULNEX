package domain

import (
	"testing"
	"time"
)

func TestNewEmail_ValidAndNormalized(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		wantErr  bool
	}{
		{"Admin@Vulnex.Local", "admin@vulnex.local", false},
		{"security.analyst+tag@example.org", "security.analyst+tag@example.org", false},
		{"invalid-email", "", true},
		{"@no-user.com", "", true},
		{"no-domain@", "", true},
		{"", "", true},
		{"   spaces@example.com  ", "spaces@example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := NewEmail(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewEmail(%q) expected error, got nil", tt.input)
				}
			} else {
				if err != nil {
					t.Fatalf("NewEmail(%q) unexpected error: %v", tt.input, err)
				}
				if got.String() != tt.expected {
					t.Errorf("NewEmail(%q) = %q, want %q", tt.input, got.String(), tt.expected)
				}
			}
		})
	}
}

func TestNewUser_CreationAndInvariants(t *testing.T) {
	email, err := NewEmail("operator@vulnex.local")
	if err != nil {
		t.Fatalf("Failed to create email: %v", err)
	}

	user, err := NewUser(email, "hashed_secret_password", "Security Operator")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	if user.ID().String() == "" {
		t.Errorf("Expected user to have non-empty ID")
	}
	if user.Status() != UserStatusActive {
		t.Errorf("Expected initial status to be %q, got %q", UserStatusActive, user.Status())
	}
	if user.DisplayName() != "Security Operator" {
		t.Errorf("Unexpected display name: %s", user.DisplayName())
	}

	// Test state transition: Suspend
	user.Suspend()
	if user.Status() != UserStatusSuspended {
		t.Errorf("Expected status to be suspended, got %q", user.Status())
	}

	// Test state transition: Activate
	user.Activate()
	if user.Status() != UserStatusActive {
		t.Errorf("Expected status to be active, got %q", user.Status())
	}
}

func TestNewUser_EmptyDisplayNameFails(t *testing.T) {
	email, _ := NewEmail("test@example.com")
	_, err := NewUser(email, "hash", "")
	if err == nil {
		t.Errorf("Expected error when display name is empty, got nil")
	}
}

func TestSetupToken_ClaimAndInvariants(t *testing.T) {
	now := time.Now()
	token, rawSecret, err := GenerateSetupToken(24 * time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate setup token: %v", err)
	}

	if token.IsClaimed() {
		t.Errorf("Expected new setup token not to be claimed")
	}

	// 1. Claiming with invalid raw secret fails
	err = token.Claim("wrong-secret", now)
	if err == nil {
		t.Errorf("Expected claim with invalid secret to fail")
	}
	if token.IsClaimed() {
		t.Errorf("Token should not be claimed after invalid attempt")
	}

	// 2. Claiming with valid raw secret succeeds
	err = token.Claim(rawSecret, now)
	if err != nil {
		t.Fatalf("Expected valid claim to succeed, got: %v", err)
	}
	if !token.IsClaimed() {
		t.Errorf("Expected token to be marked claimed")
	}

	// 3. Claiming again (replay attack) fails
	err = token.Claim(rawSecret, now.Add(time.Minute))
	if err != ErrSetupTokenAlreadyUsed {
		t.Errorf("Expected ErrSetupTokenAlreadyUsed on second claim, got: %v", err)
	}
}

func TestSetupToken_ExpiredClaimFails(t *testing.T) {
	now := time.Now()
	token, rawSecret, err := GenerateSetupToken(1 * time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate setup token: %v", err)
	}

	// Try claiming after expiration
	claimTime := now.Add(2 * time.Hour)
	err = token.Claim(rawSecret, claimTime)
	if err != ErrSetupTokenExpired {
		t.Errorf("Expected ErrSetupTokenExpired, got: %v", err)
	}
}
