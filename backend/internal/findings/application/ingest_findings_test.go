package application

import (
	"context"
	"testing"
	"time"

	"github.com/vulnex/vulnex/backend/internal/findings/domain"
)

type fakeFindingRepository struct {
	findings map[string]*domain.Finding // keyed by fingerprint
}

func newFakeFindingRepository() *fakeFindingRepository {
	return &fakeFindingRepository{findings: make(map[string]*domain.Finding)}
}

func (f *fakeFindingRepository) Save(ctx context.Context, finding *domain.Finding) error {
	f.findings[finding.Fingerprint().String()] = finding
	return nil
}

func (f *fakeFindingRepository) FindByID(ctx context.Context, id domain.FindingID) (*domain.Finding, error) {
	for _, finding := range f.findings {
		if finding.ID().String() == id.String() {
			return finding, nil
		}
	}
	return nil, nil
}

func (f *fakeFindingRepository) FindByFingerprint(ctx context.Context, projectID domain.ProjectID, fp domain.FindingFingerprint) (*domain.Finding, error) {
	finding, ok := f.findings[fp.String()]
	if !ok {
		return nil, nil
	}
	return finding, nil
}

func (f *fakeFindingRepository) ListByProject(ctx context.Context, projectID domain.ProjectID) ([]*domain.Finding, error) {
	var res []*domain.Finding
	for _, finding := range f.findings {
		res = append(res, finding)
	}
	return res, nil
}

func (f *fakeFindingRepository) Update(ctx context.Context, finding *domain.Finding) error {
	f.findings[finding.Fingerprint().String()] = finding
	return nil
}

func TestIngestFindingsUseCase_NewFindingIngested(t *testing.T) {
	ctx := context.Background()
	repo := newFakeFindingRepository()
	riskService := domain.RiskScoringService{}

	useCase := NewIngestFindingsUseCase(repo, riskService)

	now := time.Now().UTC()
	cmd := IngestFindingsCommand{
		WorkspaceID: "ws-1",
		ProjectID:   "proj-1",
		ScanID:      "scan-1",
		AssetID:     "asset-1",
		Observations: []RawObservationDTO{
			{
				Title:            "Open Port 443 HTTPS",
				Description:      "Nginx 1.24 web server",
				Severity:         "MEDIUM",
				Confidence:       "CONFIRMED",
				SourceScanner:    "nmap",
				VulnerabilityKey: "port-443-open",
				LocationKey:      "tcp/443",
				AssetCriticality: "HIGH",
			},
		},
		DiscoveredAt: now,
	}

	res, err := useCase.Execute(ctx, cmd)
	if err != nil {
		t.Fatalf("IngestFindingsUseCase failed: %v", err)
	}

	if res.NewCount != 1 || res.UpdatedCount != 0 {
		t.Errorf("Expected 1 new finding, 0 updated; got new=%d, updated=%d", res.NewCount, res.UpdatedCount)
	}

	if len(repo.findings) != 1 {
		t.Fatalf("Expected 1 finding in repo, got %d", len(repo.findings))
	}
}

func TestIngestFindingsUseCase_DeduplicationOnRepeatDiscovery(t *testing.T) {
	ctx := context.Background()
	repo := newFakeFindingRepository()
	riskService := domain.RiskScoringService{}

	useCase := NewIngestFindingsUseCase(repo, riskService)

	t1 := time.Now().UTC().Add(-2 * time.Hour)
	observation := RawObservationDTO{
		Title:            "Exposed MySQL Port 3306",
		Description:      "MySQL 8.0 listening publicly",
		Severity:         "HIGH",
		Confidence:       "CONFIRMED",
		SourceScanner:    "nmap",
		VulnerabilityKey: "port-3306-open",
		LocationKey:      "tcp/3306",
		AssetCriticality: "CRITICAL",
	}

	// 1. First Scan ingests observation
	cmd1 := IngestFindingsCommand{
		WorkspaceID:  "ws-1",
		ProjectID:    "proj-1",
		ScanID:       "scan-1",
		AssetID:      "asset-1",
		Observations: []RawObservationDTO{observation},
		DiscoveredAt: t1,
	}
	_, _ = useCase.Execute(ctx, cmd1)

	// 2. Second Scan (2 hours later) detects the identical issue
	t2 := time.Now().UTC()
	cmd2 := IngestFindingsCommand{
		WorkspaceID:  "ws-1",
		ProjectID:    "proj-1",
		ScanID:       "scan-2",
		AssetID:      "asset-1",
		Observations: []RawObservationDTO{observation},
		DiscoveredAt: t2,
	}
	res2, err := useCase.Execute(ctx, cmd2)
	if err != nil {
		t.Fatalf("Second ingestion failed: %v", err)
	}

	// Deduplication Invariant: 0 new findings, 1 updated finding!
	if res2.NewCount != 0 || res2.UpdatedCount != 1 {
		t.Errorf("Deduplication failed: expected 0 new, 1 updated; got new=%d, updated=%d", res2.NewCount, res2.UpdatedCount)
	}

	// Repository must STILL contain only 1 finding record
	if len(repo.findings) != 1 {
		t.Errorf("Repository must contain exactly 1 deduplicated finding, got %d", len(repo.findings))
	}

	// Finding's last_seen_at must be updated to t2
	for _, finding := range repo.findings {
		if !finding.LastSeenAt().Equal(t2) {
			t.Errorf("Expected last_seen_at to be %v, got %v", t2, finding.LastSeenAt())
		}
		if !finding.FirstSeenAt().Equal(t1) {
			t.Errorf("FirstSeenAt must remain unchanged at %v, got %v", t1, finding.FirstSeenAt())
		}
	}
}
