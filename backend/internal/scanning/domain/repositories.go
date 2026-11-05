package domain

import (
	"context"
	"time"
)

// ScanRepository defines persistence operations for the Scan aggregate.
type ScanRepository interface {
	Save(ctx context.Context, scan *Scan) error
	FindByID(ctx context.Context, id ScanID) (*Scan, error)
	FindByWorkspace(ctx context.Context, workspaceID WorkspaceID) ([]*Scan, error)
	Update(ctx context.Context, scan *Scan) error
}

type JobID struct {
	value string
}

func (j JobID) String() string {
	return j.value
}

// ScanJob represents an executable worker task in the queue.
type ScanJob struct {
	ID             JobID
	ScanID         ScanID
	Status         ScanStatus
	Attempt        int
	LeaseOwner     string
	LeaseExpiresAt *time.Time
	AvailableAt    time.Time
}

// ScanJobQueuePort defines non-blocking lease claiming and heartbeat operations.
type ScanJobQueuePort interface {
	Enqueue(ctx context.Context, job *ScanJob) error
	ClaimNext(ctx context.Context, workerID string, leaseDuration time.Duration) (*ScanJob, error)
	Heartbeat(ctx context.Context, jobID JobID, workerID string, extension time.Duration) error
	Complete(ctx context.Context, jobID JobID) error
	Fail(ctx context.Context, jobID JobID, reason string) error
}
