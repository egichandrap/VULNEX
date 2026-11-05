package application

import (
	"context"
	"testing"
	"time"

	"github.com/vulnex/vulnex/backend/internal/scanning/domain"
)

type fakeScanRepository struct {
	scans map[string]*domain.Scan
}

func newFakeScanRepository() *fakeScanRepository {
	return &fakeScanRepository{scans: make(map[string]*domain.Scan)}
}

func (f *fakeScanRepository) Save(ctx context.Context, s *domain.Scan) error {
	f.scans[s.ID().String()] = s
	return nil
}

func (f *fakeScanRepository) FindByID(ctx context.Context, id domain.ScanID) (*domain.Scan, error) {
	s, ok := f.scans[id.String()]
	if !ok {
		return nil, domain.ErrInvalidStateTransition // or not found
	}
	return s, nil
}

func (f *fakeScanRepository) FindByWorkspace(ctx context.Context, wsID domain.WorkspaceID) ([]*domain.Scan, error) {
	var res []*domain.Scan
	for _, s := range f.scans {
		if s.WorkspaceID().String() == wsID.String() {
			res = append(res, s)
		}
	}
	return res, nil
}

func (f *fakeScanRepository) Update(ctx context.Context, s *domain.Scan) error {
	f.scans[s.ID().String()] = s
	return nil
}

type fakeJobQueue struct {
	enqueued []*domain.ScanJob
}

func newFakeJobQueue() *fakeJobQueue {
	return &fakeJobQueue{enqueued: make([]*domain.ScanJob, 0)}
}

func (f *fakeJobQueue) Enqueue(ctx context.Context, job *domain.ScanJob) error {
	f.enqueued = append(f.enqueued, job)
	return nil
}

func (f *fakeJobQueue) ClaimNext(ctx context.Context, workerID string, leaseDuration time.Duration) (*domain.ScanJob, error) {
	return nil, nil
}

func (f *fakeJobQueue) Heartbeat(ctx context.Context, jobID domain.JobID, workerID string, ext time.Duration) error {
	return nil
}

func (f *fakeJobQueue) Complete(ctx context.Context, jobID domain.JobID) error {
	return nil
}

func (f *fakeJobQueue) Fail(ctx context.Context, jobID domain.JobID, reason string) error {
	return nil
}

type fakeAuthGate struct {
	isValid bool
	rules   []ScopeRuleDTO
}

func (f fakeAuthGate) VerifyAuthorization(ctx context.Context, authID string) (bool, []ScopeRuleDTO, error) {
	return f.isValid, f.rules, nil
}

func TestRequestScanUseCase_Success(t *testing.T) {
	ctx := context.Background()
	scanRepo := newFakeScanRepository()
	jobQueue := newFakeJobQueue()
	authGate := fakeAuthGate{
		isValid: true,
		rules: []ScopeRuleDTO{
			{RuleType: "INCLUDE", TargetType: "DOMAIN", Target: "security.example.com"},
		},
	}

	useCase := NewRequestScanUseCase(scanRepo, jobQueue, authGate)

	scanID, _ := domain.GenerateScanID()
	wsID, _ := domain.GenerateWorkspaceID()
	projectID, _ := domain.GenerateProjectID()
	authID, _ := domain.GenerateAuthorizationID()
	profileID, _ := domain.GenerateProfileVersionID()
	userID, _ := domain.GenerateUserID()

	res, err := useCase.Execute(ctx, RequestScanCommand{
		ScanID:           scanID.String(),
		WorkspaceID:      wsID.String(),
		ProjectID:        projectID.String(),
		AuthorizationID:  authID.String(),
		ProfileVersionID: profileID.String(),
		RequestedBy:      userID.String(),
		IdempotencyKey:   "idem-scan-1",
		Targets:          []string{"security.example.com"},
	})

	if err != nil {
		t.Fatalf("RequestScanUseCase failed: %v", err)
	}

	if res.Status != string(domain.ScanStatusQueued) {
		t.Errorf("Expected status QUEUED, got: %s", res.Status)
	}

	if len(jobQueue.enqueued) != 1 {
		t.Errorf("Expected 1 job enqueued in queue, got %d", len(jobQueue.enqueued))
	}
}

func TestRequestScanUseCase_OutOfScopeRejected(t *testing.T) {
	ctx := context.Background()
	scanRepo := newFakeScanRepository()
	jobQueue := newFakeJobQueue()
	authGate := fakeAuthGate{
		isValid: true,
		rules: []ScopeRuleDTO{
			{RuleType: "INCLUDE", TargetType: "DOMAIN", Target: "allowed.example.com"},
			{RuleType: "EXCLUDE", TargetType: "DOMAIN", Target: "restricted.example.com"},
		},
	}

	useCase := NewRequestScanUseCase(scanRepo, jobQueue, authGate)

	// Target 1: outside scope
	_, err := useCase.Execute(ctx, RequestScanCommand{
		Targets: []string{"unauthorized.org"},
	})
	if err != ErrTargetOutOfScope {
		t.Errorf("Expected ErrTargetOutOfScope, got: %v", err)
	}

	// Target 2: explicitly excluded
	_, err = useCase.Execute(ctx, RequestScanCommand{
		Targets: []string{"restricted.example.com"},
	})
	if err != ErrTargetOutOfScope {
		t.Errorf("Expected ErrTargetOutOfScope for excluded target, got: %v", err)
	}
}

func TestRequestScanUseCase_InvalidAuthorizationRejected(t *testing.T) {
	ctx := context.Background()
	scanRepo := newFakeScanRepository()
	jobQueue := newFakeJobQueue()
	authGate := fakeAuthGate{
		isValid: false, // Expired or revoked
	}

	useCase := NewRequestScanUseCase(scanRepo, jobQueue, authGate)

	_, err := useCase.Execute(ctx, RequestScanCommand{
		Targets: []string{"security.example.com"},
	})
	if err != ErrAuthorizationInvalidOrExpired {
		t.Errorf("Expected ErrAuthorizationInvalidOrExpired, got: %v", err)
	}
}
