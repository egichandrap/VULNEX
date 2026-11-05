package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/vulnex/vulnex/backend/internal/platform/postgres"
	"github.com/vulnex/vulnex/backend/internal/scanning/domain"
)

var (
	ErrNoJobAvailable = errors.New("no scan job available for claiming")
	ErrJobNotFound    = errors.New("scan job not found")
	ErrLeaseLost      = errors.New("scan job lease lost or expired")
)

// ScanJobQueue implements domain.ScanJobQueuePort using PostgreSQL FOR UPDATE SKIP LOCKED.
type ScanJobQueue struct {
	db postgres.DBTX
}

func NewScanJobQueue(db postgres.DBTX) *ScanJobQueue {
	return &ScanJobQueue{db: db}
}

func (q *ScanJobQueue) Enqueue(ctx context.Context, job *domain.ScanJob) error {
	query := `
		INSERT INTO scan_jobs (id, scan_id, status, attempt, available_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, clock_timestamp(), clock_timestamp())
	`
	_, err := q.db.ExecContext(
		ctx,
		query,
		job.ID.String(),
		job.ScanID.String(),
		string(job.Status),
		job.Attempt,
		job.AvailableAt,
	)
	if err != nil {
		return fmt.Errorf("failed to enqueue scan job: %w", err)
	}
	return nil
}

func (q *ScanJobQueue) ClaimNext(ctx context.Context, workerID string, leaseDuration time.Duration) (*domain.ScanJob, error) {
	if leaseDuration <= 0 {
		leaseDuration = 5 * time.Minute
	}

	leaseExpiresAt := time.Now().UTC().Add(leaseDuration)

	// Atomic claim via subquery with FOR UPDATE SKIP LOCKED
	query := `
		UPDATE scan_jobs
		SET status = 'RUNNING',
		    attempt = attempt + 1,
		    lease_owner = $1,
		    lease_expires_at = $2,
		    updated_at = clock_timestamp()
		WHERE id = (
			SELECT id
			FROM scan_jobs
			WHERE (status = 'QUEUED' AND available_at <= clock_timestamp())
			   OR (status = 'RUNNING' AND lease_expires_at < clock_timestamp())
			ORDER BY created_at ASC
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		)
		RETURNING id, scan_id, status, attempt, lease_owner, lease_expires_at, available_at
	`

	row := q.db.QueryRowContext(ctx, query, workerID, leaseExpiresAt)

	var (
		idStr       string
		scanIDStr   string
		statusStr   string
		attempt     int
		leaseOwner  sql.NullString
		expiresAt   *time.Time
		availableAt time.Time
	)

	err := row.Scan(&idStr, &scanIDStr, &statusStr, &attempt, &leaseOwner, &expiresAt, &availableAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoJobAvailable
		}
		return nil, fmt.Errorf("failed to claim next scan job: %w", err)
	}

	jobID, err := domain.NewJobID(idStr)
	if err != nil {
		return nil, err
	}

	scanID, err := domain.NewScanID(scanIDStr)
	if err != nil {
		return nil, err
	}

	return &domain.ScanJob{
		ID:             jobID,
		ScanID:         scanID,
		Status:         domain.ScanStatus(statusStr),
		Attempt:        attempt,
		LeaseOwner:     leaseOwner.String,
		LeaseExpiresAt: expiresAt,
		AvailableAt:    availableAt,
	}, nil
}

func (q *ScanJobQueue) Heartbeat(ctx context.Context, jobID domain.JobID, workerID string, extension time.Duration) error {
	newExpiration := time.Now().UTC().Add(extension)

	query := `
		UPDATE scan_jobs
		SET lease_expires_at = $1, updated_at = clock_timestamp()
		WHERE id = $2 AND lease_owner = $3 AND status = 'RUNNING'
	`
	res, err := q.db.ExecContext(ctx, query, newExpiration, jobID.String(), workerID)
	if err != nil {
		return fmt.Errorf("failed to heartbeat job lease: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check heartbeat rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrLeaseLost
	}

	return nil
}

func (q *ScanJobQueue) Complete(ctx context.Context, jobID domain.JobID) error {
	query := `
		UPDATE scan_jobs
		SET status = 'COMPLETED', lease_owner = NULL, lease_expires_at = NULL, updated_at = clock_timestamp()
		WHERE id = $1
	`
	res, err := q.db.ExecContext(ctx, query, jobID.String())
	if err != nil {
		return fmt.Errorf("failed to mark job completed: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check completion rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrJobNotFound
	}

	return nil
}

func (q *ScanJobQueue) Fail(ctx context.Context, jobID domain.JobID, reason string) error {
	query := `
		UPDATE scan_jobs
		SET status = 'FAILED', last_error_code = $1, lease_owner = NULL, lease_expires_at = NULL, updated_at = clock_timestamp()
		WHERE id = $2
	`
	res, err := q.db.ExecContext(ctx, query, reason, jobID.String())
	if err != nil {
		return fmt.Errorf("failed to mark job failed: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check failure rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrJobNotFound
	}

	return nil
}
