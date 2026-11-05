package storage_test

import (
	"context"
	"testing"

	"github.com/vulnex/vulnex/installer/internal/infrastructure/storage"
)

func TestFileLock_AcquireAndRelease(t *testing.T) {
	tempDir := t.TempDir()
	locker := storage.NewFileLock()
	ctx := context.Background()

	unlock1, err := locker.AcquireLock(ctx, tempDir)
	if err != nil {
		t.Fatalf("first AcquireLock failed: %v", err)
	}

	// Second acquire should fail because lock is held
	_, err = locker.AcquireLock(ctx, tempDir)
	if err == nil {
		t.Fatal("expected second AcquireLock to fail, got nil")
	}

	// Release first lock
	if err := unlock1(); err != nil {
		t.Fatalf("unlock failed: %v", err)
	}

	// Third acquire should succeed now
	unlock2, err := locker.AcquireLock(ctx, tempDir)
	if err != nil {
		t.Fatalf("subsequent AcquireLock failed: %v", err)
	}
	_ = unlock2()
}
