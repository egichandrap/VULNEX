package application

import (
	"context"
	"fmt"
	"time"

	"github.com/vulnex/vulnex/backend/internal/findings/domain"
)

type RawObservationDTO struct {
	Title            string
	Description      string
	Severity         string // INFO, LOW, MEDIUM, HIGH, CRITICAL
	Confidence       string // LOW, MEDIUM, HIGH, CONFIRMED
	SourceScanner    string
	VulnerabilityKey string
	LocationKey      string
	AssetCriticality string // LOW, MEDIUM, HIGH, CRITICAL
}

type IngestFindingsCommand struct {
	WorkspaceID  string
	ProjectID    string
	ScanID       string
	AssetID      string
	Observations []RawObservationDTO
	DiscoveredAt time.Time
}

type IngestFindingsResult struct {
	NewCount     int
	UpdatedCount int
}

// IngestFindingsUseCase processes raw scanner observations with deterministic deduplication.
type IngestFindingsUseCase struct {
	repo        domain.FindingRepository
	riskService domain.RiskScoringService
}

func NewIngestFindingsUseCase(
	repo domain.FindingRepository,
	riskService domain.RiskScoringService,
) *IngestFindingsUseCase {
	return &IngestFindingsUseCase{
		repo:        repo,
		riskService: riskService,
	}
}

func (uc *IngestFindingsUseCase) Execute(ctx context.Context, cmd IngestFindingsCommand) (*IngestFindingsResult, error) {
	result := &IngestFindingsResult{}

	workspaceID, _ := domain.GenerateWorkspaceID()
	projectID, _ := domain.GenerateProjectID()
	scanID, _ := domain.GenerateScanID()
	assetID, _ := domain.GenerateAssetID()

	for _, obs := range cmd.Observations {
		// 1. Calculate deterministic fingerprint
		fingerprint := domain.CalculateFingerprint(
			cmd.ProjectID,
			cmd.AssetID,
			obs.SourceScanner,
			obs.VulnerabilityKey,
			obs.LocationKey,
		)

		// 2. Query repository for existing finding by fingerprint (Deduplication)
		existing, err := uc.repo.FindByFingerprint(ctx, projectID, fingerprint)
		if err != nil {
			return nil, fmt.Errorf("failed to query finding by fingerprint: %w", err)
		}

		if existing != nil {
			// Deduplication hit: update last_seen_at
			existing.RecordObservation(scanID, cmd.DiscoveredAt)
			if err := uc.repo.Update(ctx, existing); err != nil {
				return nil, fmt.Errorf("failed to update deduplicated finding: %w", err)
			}
			result.UpdatedCount++
		} else {
			// New observation: compute explainable risk score and create Finding aggregate
			sev := domain.Severity(obs.Severity)
			conf := domain.Confidence(obs.Confidence)
			crit := domain.AssetCriticality(obs.AssetCriticality)

			riskScore := uc.riskService.Calculate(sev, conf, crit)
			findingID, _ := domain.GenerateFindingID()

			newFinding, err := domain.NewFinding(
				findingID,
				workspaceID,
				projectID,
				scanID,
				assetID,
				fingerprint,
				obs.Title,
				obs.Description,
				sev,
				conf,
				obs.SourceScanner,
				cmd.DiscoveredAt,
				riskScore,
				"v1.0.0",
			)
			if err != nil {
				return nil, fmt.Errorf("failed to initialize finding: %w", err)
			}

			if err := uc.repo.Save(ctx, newFinding); err != nil {
				return nil, fmt.Errorf("failed to save new finding: %w", err)
			}
			result.NewCount++
		}
	}

	return result, nil
}
