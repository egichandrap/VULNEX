package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/vulnex/vulnex/installer/internal/domain"
)

const lockFileName = ".vulnex.lock"

// FileLock manages process concurrency using flock.
type FileLock struct{}

func NewFileLock() *FileLock {
	return &FileLock{}
}

// AcquireLock attempts to obtain a non-blocking exclusive advisory lock on the target data directory.
func (l *FileLock) AcquireLock(ctx context.Context, dataDir string) (func() error, error) {
	cleanDir := filepath.Clean(dataDir)
	if err := os.MkdirAll(cleanDir, dirPerms); err != nil {
		return nil, domain.WrapError(domain.ExitGeneralFailure, "failed to create directory for lock", err)
	}

	lockPath := filepath.Join(cleanDir, lockFileName)
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, domain.WrapError(domain.ExitGeneralFailure, "failed to open lock file", err)
	}

	// Non-blocking exclusive lock
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = f.Close()
		return nil, domain.WrapError(domain.ExitGeneralFailure, "installation lock already held by another process", err)
	}

	// Record PID in lock file
	_ = f.Truncate(0)
	_, _ = fmt.Fprintf(f, "%d\n", os.Getpid())

	unlock := func() error {
		_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		_ = f.Close()
		_ = os.Remove(lockPath)
		return nil
	}

	return unlock, nil
}
