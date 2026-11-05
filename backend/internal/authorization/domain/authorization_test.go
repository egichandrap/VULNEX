package domain

import (
	"testing"
	"time"
)

func TestAuthorization_CreationAndInvariants(t *testing.T) {
	now := time.Now().UTC()
	validFrom := now.Add(-1 * time.Hour)
	validUntil := now.Add(24 * time.Hour)

	authID, _ := GenerateAuthorizationID()
	projectID, _ := GenerateProjectID()
	approverID, _ := GenerateUserID()

	auth, err := NewAuthorization(
		authID,
		projectID,
		"Acme Corp Security Officer",
		approverID,
		"docs/auth/acme-pentest-signed.pdf",
		"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		validFrom,
		validUntil,
		[]string{"PORT_SCAN", "WEB_PASSIVE"},
	)
	if err != nil {
		t.Fatalf("Failed to create authorization: %v", err)
	}

	if !auth.IsValidAt(now) {
		t.Errorf("Expected authorization to be valid now")
	}

	// Invariant: Expired authorization
	if auth.IsValidAt(validUntil.Add(time.Hour)) {
		t.Errorf("Authorization should not be valid after validUntil")
	}

	// Invariant: Before validFrom
	if auth.IsValidAt(validFrom.Add(-time.Hour)) {
		t.Errorf("Authorization should not be valid before validFrom")
	}

	// Invariant: Terminal revocation
	err = auth.Revoke("Assessment aborted by client")
	if err != nil {
		t.Fatalf("Revocation failed: %v", err)
	}
	if auth.IsValidAt(now) {
		t.Errorf("Revoked authorization must never be valid")
	}

	// Re-revoking fails
	err = auth.Revoke("Another reason")
	if err != ErrAuthorizationAlreadyRevoked {
		t.Errorf("Expected ErrAuthorizationAlreadyRevoked, got: %v", err)
	}
}

func TestAuthorization_InvalidPeriodFails(t *testing.T) {
	now := time.Now().UTC()
	authID, _ := GenerateAuthorizationID()
	projectID, _ := GenerateProjectID()
	approverID, _ := GenerateUserID()

	// validUntil before validFrom
	_, err := NewAuthorization(
		authID,
		projectID,
		"Party",
		approverID,
		"ref",
		"checksum",
		now.Add(time.Hour),
		now,
		[]string{"PORT_SCAN"},
	)
	if err != ErrInvalidValidityPeriod {
		t.Errorf("Expected ErrInvalidValidityPeriod, got: %v", err)
	}
}

func TestScopePolicy_ExclusionsOverrideInclusions(t *testing.T) {
	policy := ScopePolicy{}

	rules := []ScopeRule{
		{
			RuleType:        RuleTypeInclude,
			TargetType:      ScopeTargetDomain,
			CanonicalTarget: "example.com",
		},
		{
			RuleType:        RuleTypeInclude,
			TargetType:      ScopeTargetCIDR,
			CanonicalTarget: "198.51.100.0/24",
		},
		{
			RuleType:        RuleTypeExclude,
			TargetType:      ScopeTargetDomain,
			CanonicalTarget: "billing.example.com",
		},
		{
			RuleType:        RuleTypeExclude,
			TargetType:      ScopeTargetIP,
			CanonicalTarget: "198.51.100.50",
		},
	}

	// Included domain
	if !policy.IsTargetAuthorized("example.com", rules) {
		t.Errorf("Expected 'example.com' to be authorized")
	}

	// Excluded subdomain (exclusion must strictly override)
	if policy.IsTargetAuthorized("billing.example.com", rules) {
		t.Errorf("Expected 'billing.example.com' to be rejected by exclusion")
	}

	// Included IP within CIDR
	if !policy.IsTargetAuthorized("198.51.100.20", rules) {
		t.Errorf("Expected '198.51.100.20' to be authorized via CIDR")
	}

	// Excluded exact IP within included CIDR (exclusion overrides)
	if policy.IsTargetAuthorized("198.51.100.50", rules) {
		t.Errorf("Expected '198.51.100.50' to be rejected by explicit exclusion")
	}

	// Target not in any rule
	if policy.IsTargetAuthorized("unrelated.org", rules) {
		t.Errorf("Expected 'unrelated.org' to be rejected (not included)")
	}
}
