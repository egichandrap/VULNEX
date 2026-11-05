package domain

import (
	"context"
	"io"
	"time"
)

// ScanExecutionRequest contains the parameters for an isolated scanner execution.
// Invariant: Targets and options are strictly validated prior to construction.
type ScanExecutionRequest struct {
	ScanID         ScanID
	TargetHost     string
	ResolvedIP     string
	ProfileName    string
	PortRange      string
	Timeout        time.Duration
	MaxOutputBytes int64
}

// RawArtifact contains the raw scanner output captured under bounded limits.
type RawArtifact struct {
	ScannerName    string
	ExitCode       int
	OutputReader   io.Reader
	ExecutionTime  time.Duration
	Truncated      bool
	ToolVersion    string
	ProfileVersion string
}

// ScannerAdapterPort defines the contract for external scanner invocation.
type ScannerAdapterPort interface {
	Name() string
	ValidateProfile(ctx context.Context, profileName string) error
	ExecuteScan(ctx context.Context, req ScanExecutionRequest) (*RawArtifact, error)
}
