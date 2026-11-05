package domain

import (
	"testing"
)

func TestNewWorkspace_CreationAndOwnerMembership(t *testing.T) {
	ownerID, _ := GenerateUserID()
	ws, err := NewWorkspace("Primary Security Team", ownerID)
	if err != nil {
		t.Fatalf("Failed to create workspace: %v", err)
	}

	if ws.Name() != "Primary Security Team" {
		t.Errorf("Unexpected workspace name: %s", ws.Name())
	}
	if ws.OwnerUserID().String() != ownerID.String() {
		t.Errorf("Owner ID mismatch")
	}
	if ws.AllowPrivateNetworkScan() {
		t.Errorf("Expected allow_private_network_scan to be false by default (ADR-015)")
	}

	// Invariant: Workspace must have at least one owner member
	members := ws.Members()
	if len(members) != 1 {
		t.Fatalf("Expected 1 initial member, got %d", len(members))
	}
	if members[0].UserID().String() != ownerID.String() || members[0].Role() != RoleOwner {
		t.Errorf("Expected initial member to be Owner")
	}
}

func TestWorkspace_AddAndRemoveMember(t *testing.T) {
	ownerID, _ := GenerateUserID()
	ws, _ := NewWorkspace("Test Workspace", ownerID)

	memberID, _ := GenerateUserID()
	err := ws.AddMember(memberID, RoleMember)
	if err != nil {
		t.Fatalf("Failed to add member: %v", err)
	}

	if len(ws.Members()) != 2 {
		t.Errorf("Expected 2 members, got %d", len(ws.Members()))
	}

	// Duplicate member rejected
	err = ws.AddMember(memberID, RoleAdmin)
	if err != ErrMemberAlreadyExists {
		t.Errorf("Expected ErrMemberAlreadyExists, got %v", err)
	}

	// Remove member
	err = ws.RemoveMember(memberID)
	if err != nil {
		t.Fatalf("Failed to remove member: %v", err)
	}
	if len(ws.Members()) != 1 {
		t.Errorf("Expected 1 member remaining, got %d", len(ws.Members()))
	}

	// Invariant: Cannot remove sole owner
	err = ws.RemoveMember(ownerID)
	if err != ErrCannotRemoveSoleOwner {
		t.Errorf("Expected ErrCannotRemoveSoleOwner, got %v", err)
	}
}

func TestWorkspace_PrivateNetworkScanPolicy(t *testing.T) {
	ownerID, _ := GenerateUserID()
	ws, _ := NewWorkspace("Test Workspace", ownerID)

	memberID, _ := GenerateUserID()
	_ = ws.AddMember(memberID, RoleMember)

	// Member cannot enable private network scanning
	err := ws.EnablePrivateNetworkScanning(memberID, "Testing internal RFC 1918 subnet")
	if err != ErrUnauthorizedPolicyChange {
		t.Errorf("Expected ErrUnauthorizedPolicyChange for non-admin, got %v", err)
	}
	if ws.AllowPrivateNetworkScan() {
		t.Errorf("Private network scanning should remain false")
	}

	// Owner can enable private network scanning
	err = ws.EnablePrivateNetworkScanning(ownerID, "Authorized internal assessment")
	if err != nil {
		t.Fatalf("Owner failed to enable private scanning: %v", err)
	}
	if !ws.AllowPrivateNetworkScan() {
		t.Errorf("Expected AllowPrivateNetworkScan to be true")
	}

	// Disable private network scanning
	ws.DisablePrivateNetworkScanning()
	if ws.AllowPrivateNetworkScan() {
		t.Errorf("Expected AllowPrivateNetworkScan to be false")
	}
}
