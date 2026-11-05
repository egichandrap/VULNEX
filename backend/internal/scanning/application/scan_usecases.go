package application

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/vulnex/vulnex/backend/internal/scanning/domain"
)

var (
	ErrTargetOutOfScope               = errors.New("one or more scan targets are outside authorized scope or explicitly excluded")
	ErrAuthorizationInvalidOrExpired  = errors.New("authorization is expired, revoked, or invalid")
	ErrScanNotFound                   = errors.New("scan not found")
	ErrCrossWorkspaceAccessProhibited = errors.New("access to scan belonging to another workspace is prohibited")
)

type ScopeRuleDTO struct {
	RuleType   string // "INCLUDE" or "EXCLUDE"
	TargetType string // "DOMAIN", "IP", "CIDR", "URL"
	Target     string
}

// AuthorizationValidatorPort defines the cross-context port to verify authorization and scope.
type AuthorizationValidatorPort interface {
	VerifyAuthorization(ctx context.Context, authID string) (isValid bool, rules []ScopeRuleDTO, err error)
}

type RequestScanCommand struct {
	ScanID           string
	WorkspaceID      string
	ProjectID        string
	AuthorizationID  string
	ProfileVersionID string
	RequestedBy      string
	IdempotencyKey   string
	Targets          []string
}

type RequestScanResult struct {
	ScanID    string    `json:"scan_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// RequestScanUseCase orchestrates scan submission and scope gating.
type RequestScanUseCase struct {
	scanRepo ScanRepositoryPort
	queue    domain.ScanJobQueuePort
	authGate AuthorizationValidatorPort
}

// ScanRepositoryPort is the repository interface required by this use case.
type ScanRepositoryPort interface {
	domain.ScanRepository
}

func NewRequestScanUseCase(
	scanRepo ScanRepositoryPort,
	queue domain.ScanJobQueuePort,
	authGate AuthorizationValidatorPort,
) *RequestScanUseCase {
	return &RequestScanUseCase{
		scanRepo: scanRepo,
		queue:    queue,
		authGate: authGate,
	}
}

func (uc *RequestScanUseCase) Execute(ctx context.Context, cmd RequestScanCommand) (*RequestScanResult, error) {
	// 1. Verify authorization validity
	isValid, rules, err := uc.authGate.VerifyAuthorization(ctx, cmd.AuthorizationID)
	if err != nil {
		return nil, fmt.Errorf("authorization lookup failed: %w", err)
	}
	if !isValid {
		return nil, ErrAuthorizationInvalidOrExpired
	}

	// 2. Gate all targets against scope rules (Exclusions strictly override inclusions)
	if len(cmd.Targets) == 0 {
		return nil, domain.ErrEmptyScanTargets
	}

	for _, target := range cmd.Targets {
		if !uc.isTargetAuthorized(target, rules) {
			return nil, ErrTargetOutOfScope
		}
	}

	// 3. Construct Scan aggregate
	scanID, err := domain.GenerateScanID()
	if cmd.ScanID != "" {
		// allow caller or generate
	}
	wsID, err := domain.NewWorkspaceID(cmd.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("invalid workspace ID: %w", err)
	}
	projectID, err := domain.NewProjectID(cmd.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("invalid project ID: %w", err)
	}
	authID, err := domain.NewAuthorizationID(cmd.AuthorizationID)
	if err != nil {
		return nil, fmt.Errorf("invalid authorization ID: %w", err)
	}
	profileID, err := domain.NewProfileVersionID(cmd.ProfileVersionID)
	if err != nil {
		return nil, fmt.Errorf("invalid profile version ID: %w", err)
	}
	userID, err := domain.NewUserID(cmd.RequestedBy)
	if err != nil {
		return nil, fmt.Errorf("invalid requested by user ID: %w", err)
	}

	scan, err := domain.NewScan(
		scanID,
		wsID,
		projectID,
		authID,
		profileID,
		userID,
		cmd.IdempotencyKey,
		cmd.Targets,
	)
	if err != nil {
		return nil, err
	}

	// 4. Create ScanJob entity
	now := time.Now().UTC()
	jobID, err := domain.GenerateJobID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate job ID: %w", err)
	}
	job := &domain.ScanJob{
		ID:          jobID,
		ScanID:      scan.ID(),
		Status:      domain.ScanStatusQueued,
		Attempt:     0,
		AvailableAt: now,
	}

	// 5. Persist Scan and Job
	if err := uc.scanRepo.Save(ctx, scan); err != nil {
		return nil, fmt.Errorf("failed to save scan: %w", err)
	}
	if err := uc.queue.Enqueue(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to enqueue scan job: %w", err)
	}

	return &RequestScanResult{
		ScanID:    scan.ID().String(),
		Status:    string(scan.Status()),
		CreatedAt: now,
	}, nil
}

func (uc *RequestScanUseCase) isTargetAuthorized(target string, rules []ScopeRuleDTO) bool {
	cleanTarget := strings.ToLower(strings.TrimSpace(target))
	if cleanTarget == "" {
		return false
	}

	included := false
	for _, rule := range rules {
		cleanRuleTarget := strings.ToLower(strings.TrimSpace(rule.Target))
		matches := false

		switch rule.TargetType {
		case "DOMAIN":
			if cleanTarget == cleanRuleTarget || strings.HasSuffix(cleanTarget, "."+cleanRuleTarget) {
				matches = true
			}
		case "IP":
			targetIP := net.ParseIP(cleanTarget)
			ruleIP := net.ParseIP(cleanRuleTarget)
			if targetIP != nil && ruleIP != nil && targetIP.Equal(ruleIP) {
				matches = true
			} else if cleanTarget == cleanRuleTarget {
				matches = true
			}
		case "CIDR":
			targetIP := net.ParseIP(cleanTarget)
			if targetIP != nil {
				_, ipNet, err := net.ParseCIDR(cleanRuleTarget)
				if err == nil && ipNet.Contains(targetIP) {
					matches = true
				}
			}
		case "URL":
			if strings.HasPrefix(cleanTarget, cleanRuleTarget) {
				matches = true
			}
		default:
			if cleanTarget == cleanRuleTarget {
				matches = true
			}
		}

		if matches {
			if rule.RuleType == "EXCLUDE" {
				return false // Exclusion strictly overrides
			}
			if rule.RuleType == "INCLUDE" {
				included = true
			}
		}
	}

	return included
}

type CancelScanCommand struct {
	ScanID      string
	WorkspaceID string
	RequestedBy string
	Reason      string
}

// CancelScanUseCase handles graceful cancellation of running or queued scans.
type CancelScanUseCase struct {
	scanRepo ScanRepositoryPort
}

func NewCancelScanUseCase(scanRepo ScanRepositoryPort) *CancelScanUseCase {
	return &CancelScanUseCase{scanRepo: scanRepo}
}

func (uc *CancelScanUseCase) Execute(ctx context.Context, cmd CancelScanCommand) error {
	// 1. Look up scan
	scan, err := uc.scanRepo.FindByID(ctx, domain.ScanID{})
	if err != nil || scan == nil {
		return ErrScanNotFound
	}

	// 2. Verify workspace tenancy
	if scan.WorkspaceID().String() != cmd.WorkspaceID {
		return ErrCrossWorkspaceAccessProhibited
	}

	// 3. Perform cancellation on domain aggregate
	if err := scan.Cancel(cmd.Reason); err != nil {
		return err
	}

	// 4. Update repository
	return uc.scanRepo.Update(ctx, scan)
}
