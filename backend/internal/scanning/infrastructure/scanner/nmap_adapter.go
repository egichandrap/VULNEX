package scanner

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/vulnex/vulnex/backend/internal/scanning/domain"
)

var (
	ErrUnsupportedProfile = errors.New("unsupported or unapproved scanner profile")
	ErrInvalidPortRange   = errors.New("invalid port range syntax (must match 80,443,1000-2000)")
	ErrEmptyTargetHost    = errors.New("target host cannot be empty")
)

var portRangeRegex = regexp.MustCompile(`^[0-9]+(-[0-9]+)?(,[0-9]+(-[0-9]+)?)*$`)

// CommandExecutor abstracts low-level process execution for isolated scanner containers.
type CommandExecutor interface {
	Execute(ctx context.Context, cmd string, args []string, maxBytes int64) (stdout string, exitCode int, err error)
}

// NmapAdapter implements ScannerAdapterPort for constrained Nmap network assessment.
type NmapAdapter struct {
	executor CommandExecutor
}

func NewNmapAdapter(executor CommandExecutor) *NmapAdapter {
	return &NmapAdapter{executor: executor}
}

func (a *NmapAdapter) Name() string {
	return "nmap"
}

func (a *NmapAdapter) ValidateProfile(ctx context.Context, profileName string) error {
	switch profileName {
	case "discovery", "top-ports", "common-web-ports":
		return nil
	default:
		return ErrUnsupportedProfile
	}
}

// BuildArguments derives strictly typed, allowlisted command arguments without shell interpolation.
func (a *NmapAdapter) BuildArguments(req domain.ScanExecutionRequest) ([]string, error) {
	if strings.TrimSpace(req.TargetHost) == "" {
		return nil, ErrEmptyTargetHost
	}

	var args []string

	switch req.ProfileName {
	case "discovery":
		// Fast host discovery only: ping sweep, no port scan
		args = append(args, "-sn", "-PE", "-PS80,443")

	case "top-ports":
		// TCP connect scan on top ports
		args = append(args, "-sT", "-T3")
		if req.PortRange != "" {
			if !portRangeRegex.MatchString(req.PortRange) {
				return nil, ErrInvalidPortRange
			}
			args = append(args, "-p", req.PortRange)
		} else {
			args = append(args, "--top-ports", "100")
		}

	case "common-web-ports":
		args = append(args, "-sT", "-T3", "-p", "80,443,8080,8443")

	default:
		return nil, ErrUnsupportedProfile
	}

	// Always output machine-readable XML to stdout for bounded parsing
	args = append(args, "-oX", "-", req.TargetHost)
	return args, nil
}

func (a *NmapAdapter) ExecuteScan(ctx context.Context, req domain.ScanExecutionRequest) (*domain.RawArtifact, error) {
	if err := a.ValidateProfile(ctx, req.ProfileName); err != nil {
		return nil, err
	}

	args, err := a.BuildArguments(req)
	if err != nil {
		return nil, fmt.Errorf("failed to build nmap arguments: %w", err)
	}

	timeout := req.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}

	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	maxBytes := req.MaxOutputBytes
	if maxBytes <= 0 {
		maxBytes = 50 * 1024 * 1024 // 50MB default ceiling
	}

	start := time.Now()
	output, exitCode, err := a.executor.Execute(execCtx, "nmap", args, maxBytes)
	elapsed := time.Since(start)

	if err != nil {
		return nil, fmt.Errorf("nmap execution failed: %w", err)
	}

	return &domain.RawArtifact{
		ScannerName:    "nmap",
		ExitCode:       exitCode,
		OutputReader:   strings.NewReader(output),
		ExecutionTime:  elapsed,
		Truncated:      false,
		ToolVersion:    "7.94",
		ProfileVersion: req.ProfileName + "-v1.0.0",
	}, nil
}
