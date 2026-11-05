package scanner

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/vulnex/vulnex/backend/internal/scanning/domain"
)

var (
	ErrEmptyTargetURL   = errors.New("target URL cannot be empty")
	ErrInvalidTargetURL = errors.New("target URL must be a valid http or https URL")
)

// ZapAdapter implements ScannerAdapterPort for OWASP ZAP baseline web vulnerability scanning.
type ZapAdapter struct {
	executor CommandExecutor
}

func NewZapAdapter(executor CommandExecutor) *ZapAdapter {
	return &ZapAdapter{executor: executor}
}

func (a *ZapAdapter) Name() string {
	return "zap"
}

func (a *ZapAdapter) ValidateProfile(ctx context.Context, profileName string) error {
	switch profileName {
	case "zap-baseline", "baseline":
		return nil
	default:
		return ErrUnsupportedProfile
	}
}

// BuildArguments derives strictly typed arguments for zap-baseline.py without shell interpolation.
func (a *ZapAdapter) BuildArguments(req domain.ScanExecutionRequest) ([]string, error) {
	trimmed := strings.TrimSpace(req.TargetHost)
	if trimmed == "" {
		return nil, ErrEmptyTargetURL
	}

	parsed, err := url.ParseRequestURI(trimmed)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return nil, ErrInvalidTargetURL
	}

	if err := a.ValidateProfile(context.Background(), req.ProfileName); err != nil {
		return nil, err
	}

	// zap-baseline.py arguments:
	// -t: target URL
	// -J: JSON report output
	// -m: maximum crawl depth / minutes
	args := []string{
		"zap-baseline.py",
		"-t", trimmed,
		"-J", "zap_report.json",
		"-m", "10",
	}

	return args, nil
}

func (a *ZapAdapter) ExecuteScan(ctx context.Context, req domain.ScanExecutionRequest) (*domain.RawArtifact, error) {
	if err := a.ValidateProfile(ctx, req.ProfileName); err != nil {
		return nil, err
	}

	args, err := a.BuildArguments(req)
	if err != nil {
		return nil, fmt.Errorf("failed to build zap arguments: %w", err)
	}

	timeout := req.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Minute
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	maxBytes := req.MaxOutputBytes
	if maxBytes <= 0 {
		maxBytes = 50 * 1024 * 1024 // 50MB bounded output limit (ADR-007)
	}

	start := time.Now()
	stdout, exitCode, execErr := a.executor.Execute(ctxTimeout, "zap-baseline.py", args[1:], maxBytes)

	// OWASP ZAP baseline exit codes:
	// 0: pass (no issues)
	// 1: warning (low/medium findings)
	// 2: failure (high/critical findings)
	// 3: tool execution failure / syntax error
	if execErr != nil && exitCode > 2 {
		return nil, fmt.Errorf("zap execution failed (exit %d): %w", exitCode, execErr)
	}

	return &domain.RawArtifact{
		ScannerName:    a.Name(),
		ExitCode:       exitCode,
		OutputReader:   strings.NewReader(stdout),
		ExecutionTime:  time.Since(start),
		Truncated:      false,
		ToolVersion:    "2.14",
		ProfileVersion: "1.0",
	}, nil
}
