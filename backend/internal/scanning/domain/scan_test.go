package domain

import (
	"testing"
	"time"
)

func TestScan_LifecycleAndStateTransitions(t *testing.T) {
	now := time.Now().UTC()
	scanID, _ := GenerateScanID()
	workspaceID, _ := GenerateWorkspaceID()
	projectID, _ := GenerateProjectID()
	authID, _ := GenerateAuthorizationID()
	profileID, _ := GenerateProfileVersionID()
	userID, _ := GenerateUserID()

	targets := []string{"example.com", "198.51.100.1"}

	scan, err := NewScan(
		scanID,
		workspaceID,
		projectID,
		authID,
		profileID,
		userID,
		"idem-key-12345",
		targets,
	)
	if err != nil {
		t.Fatalf("Failed to create scan: %v", err)
	}

	// 1. Initial State must be QUEUED
	if scan.Status() != ScanStatusQueued {
		t.Errorf("Expected initial status %s, got %s", ScanStatusQueued, scan.Status())
	}

	// 2. Transition QUEUED -> RUNNING succeeds
	err = scan.MarkRunning(now)
	if err != nil {
		t.Fatalf("MarkRunning failed: %v", err)
	}
	if scan.Status() != ScanStatusRunning {
		t.Errorf("Expected status %s, got %s", ScanStatusRunning, scan.Status())
	}

	// 3. Transition RUNNING -> COMPLETED succeeds
	completedTime := now.Add(5 * time.Minute)
	err = scan.MarkCompleted(completedTime)
	if err != nil {
		t.Fatalf("MarkCompleted failed: %v", err)
	}
	if scan.Status() != ScanStatusCompleted {
		t.Errorf("Expected status %s, got %s", ScanStatusCompleted, scan.Status())
	}

	// 4. Invariant: Terminal state cannot transition to any other state
	err = scan.MarkRunning(completedTime.Add(time.Minute))
	if err != ErrScanAlreadyTerminal {
		t.Errorf("Expected ErrScanAlreadyTerminal when attempting to restart completed scan, got: %v", err)
	}

	err = scan.Cancel("Cancelling completed scan")
	if err != ErrScanAlreadyTerminal {
		t.Errorf("Expected ErrScanAlreadyTerminal when cancelling completed scan, got: %v", err)
	}
}

func TestScan_CancellationFromQueued(t *testing.T) {
	scanID, _ := GenerateScanID()
	workspaceID, _ := GenerateWorkspaceID()
	projectID, _ := GenerateProjectID()
	authID, _ := GenerateAuthorizationID()
	profileID, _ := GenerateProfileVersionID()
	userID, _ := GenerateUserID()

	scan, _ := NewScan(scanID, workspaceID, projectID, authID, profileID, userID, "idem-2", []string{"test.com"})

	// Cancellation while QUEUED
	err := scan.Cancel("User requested abort")
	if err != nil {
		t.Fatalf("Cancel from QUEUED failed: %v", err)
	}
	if scan.Status() != ScanStatusCancelled {
		t.Errorf("Expected status %s, got %s", ScanStatusCancelled, scan.Status())
	}

	// Cannot transition cancelled scan to RUNNING
	err = scan.MarkRunning(time.Now())
	if err != ErrScanAlreadyTerminal {
		t.Errorf("Expected ErrScanAlreadyTerminal, got: %v", err)
	}
}

func TestScan_TimeoutTransition(t *testing.T) {
	scanID, _ := GenerateScanID()
	workspaceID, _ := GenerateWorkspaceID()
	projectID, _ := GenerateProjectID()
	authID, _ := GenerateAuthorizationID()
	profileID, _ := GenerateProfileVersionID()
	userID, _ := GenerateUserID()

	scan, _ := NewScan(scanID, workspaceID, projectID, authID, profileID, userID, "idem-3", []string{"test.com"})
	_ = scan.MarkRunning(time.Now())

	// Timeout transition
	err := scan.MarkTimedOut(time.Now())
	if err != nil {
		t.Fatalf("MarkTimedOut failed: %v", err)
	}
	if scan.Status() != ScanStatusTimedOut {
		t.Errorf("Expected status %s, got %s", ScanStatusTimedOut, scan.Status())
	}
}

func TestScan_EmptyTargetsFails(t *testing.T) {
	scanID, _ := GenerateScanID()
	workspaceID, _ := GenerateWorkspaceID()
	projectID, _ := GenerateProjectID()
	authID, _ := GenerateAuthorizationID()
	profileID, _ := GenerateProfileVersionID()
	userID, _ := GenerateUserID()

	_, err := NewScan(scanID, workspaceID, projectID, authID, profileID, userID, "idem-4", []string{})
	if err != ErrEmptyScanTargets {
		t.Errorf("Expected ErrEmptyScanTargets, got: %v", err)
	}
}
