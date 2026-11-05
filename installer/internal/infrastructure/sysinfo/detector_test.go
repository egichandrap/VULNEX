package sysinfo_test

import (
	"context"
	"runtime"
	"testing"

	"github.com/vulnex/vulnex/installer/internal/infrastructure/sysinfo"
)

func TestSysinfoDetector_DetectOS(t *testing.T) {
	detector := sysinfo.NewDetector()
	osName, arch, err := detector.DetectOS(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if osName != runtime.GOOS {
		t.Errorf("expected OS %s, got %s", runtime.GOOS, osName)
	}
	if arch != runtime.GOARCH {
		t.Errorf("expected Arch %s, got %s", runtime.GOARCH, arch)
	}
}

func TestSysinfoDetector_CheckPort(t *testing.T) {
	detector := sysinfo.NewDetector()
	// An ephemeral port or high random port is usually free
	available, err := detector.CheckPort(context.Background(), 49151)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !available {
		t.Log("Note: port 49151 was occupied")
	}
}
