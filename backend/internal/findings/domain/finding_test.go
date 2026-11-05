package domain

import (
	"testing"
	"time"
)

func TestFindingFingerprint_Deterministic(t *testing.T) {
	fp1 := CalculateFingerprint(
		"proj-1",
		"security.example.com",
		"nmap",
		"open-port-22",
		"tcp/22",
	)
	fp2 := CalculateFingerprint(
		"proj-1",
		"security.example.com",
		"nmap",
		"open-port-22",
		"tcp/22",
	)
	fpDiff := CalculateFingerprint(
		"proj-1",
		"security.example.com",
		"nmap",
		"open-port-80", // different port
		"tcp/80",
	)

	if fp1.String() != fp2.String() {
		t.Errorf("Fingerprints should be deterministic for identical inputs")
	}
	if fp1.String() == fpDiff.String() {
		t.Errorf("Fingerprints should differ for different finding locations")
	}
}

func TestFinding_ReobservationUpdatesLastSeen(t *testing.T) {
	now := time.Now().UTC()
	findingID, _ := GenerateFindingID()
	workspaceID, _ := GenerateWorkspaceID()
	projectID, _ := GenerateProjectID()
	scanID, _ := GenerateScanID()
	assetID, _ := GenerateAssetID()

	fp := CalculateFingerprint(projectID.String(), "192.168.1.1", "nmap", "ssh-running", "tcp/22")

	finding, err := NewFinding(
		findingID,
		workspaceID,
		projectID,
		scanID,
		assetID,
		fp,
		"SSH Service Detected",
		"OpenSSH 8.9p1 Ubuntu running on port 22",
		SeverityInfo,
		ConfidenceConfirmed,
		"nmap",
		now,
		1.0,
		"v1.0.0",
	)
	if err != nil {
		t.Fatalf("Failed to create finding: %v", err)
	}

	if finding.Status() != FindingStatusOpen {
		t.Errorf("Expected initial status to be OPEN")
	}

	// Re-observation 1 hour later in another scan
	laterScanID, _ := GenerateScanID()
	laterTime := now.Add(time.Hour)
	finding.RecordObservation(laterScanID, laterTime)

	if !finding.LastSeenAt().Equal(laterTime) {
		t.Errorf("Expected last_seen_at to be updated to %v, got: %v", laterTime, finding.LastSeenAt())
	}
	if !finding.FirstSeenAt().Equal(now) {
		t.Errorf("FirstSeenAt must remain unchanged")
	}
}

func TestFinding_StatusTransitions(t *testing.T) {
	now := time.Now().UTC()
	findingID, _ := GenerateFindingID()
	workspaceID, _ := GenerateWorkspaceID()
	projectID, _ := GenerateProjectID()
	scanID, _ := GenerateScanID()
	assetID, _ := GenerateAssetID()
	fp := CalculateFingerprint("p", "a", "s", "t", "l")

	finding, _ := NewFinding(findingID, workspaceID, projectID, scanID, assetID, fp, "T", "D", SeverityHigh, ConfidenceHigh, "zap", now, 7.5, "v1")

	// Transition to ACCEPTED_RISK requires non-empty justification
	err := finding.AcceptRisk("", now)
	if err == nil {
		t.Errorf("Expected error when accepting risk without justification")
	}

	err = finding.AcceptRisk("Internal compensating controls applied (firewall restricted)", now)
	if err != nil {
		t.Fatalf("AcceptRisk failed with valid justification: %v", err)
	}
	if finding.Status() != FindingStatusAcceptedRisk {
		t.Errorf("Expected status ACCEPTED_RISK")
	}

	// Transition to RESOLVED
	err = finding.Resolve(now.Add(time.Hour))
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}
	if finding.Status() != FindingStatusResolved {
		t.Errorf("Expected status RESOLVED")
	}
}

func TestRiskScoringService_Calculation(t *testing.T) {
	service := RiskScoringService{}

	// Critical finding on critical asset with confirmed confidence -> highest score (10.0)
	scoreHigh := service.Calculate(SeverityCritical, ConfidenceConfirmed, AssetCriticalityCritical)
	if scoreHigh < 9.0 || scoreHigh > 10.0 {
		t.Errorf("Expected critical score around 9.0-10.0, got: %f", scoreHigh)
	}

	// Low finding on low asset with low confidence -> very low score
	scoreLow := service.Calculate(SeverityLow, ConfidenceLow, AssetCriticalityLow)
	if scoreLow >= scoreHigh {
		t.Errorf("Low score (%f) must be less than high score (%f)", scoreLow, scoreHigh)
	}
}
