package domain

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
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

func NewJobID(val string) (JobID, error) {
	if val == "" {
		return JobID{}, errors.New("job ID cannot be empty")
	}
	return JobID{value: val}, nil
}

func GenerateJobID() (JobID, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return JobID{}, fmt.Errorf("failed to generate UUID: %w", err)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	val := fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
	return JobID{value: val}, nil
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
