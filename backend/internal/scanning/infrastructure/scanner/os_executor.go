package scanner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
)

// OSCommandExecutor runs commands directly as an OS process without shell interpolation.
type OSCommandExecutor struct{}

func NewOSCommandExecutor() *OSCommandExecutor {
	return &OSCommandExecutor{}
}

func (e *OSCommandExecutor) Execute(ctx context.Context, cmd string, args []string, maxBytes int64) (string, int, error) {
	if maxBytes <= 0 {
		maxBytes = 50 * 1024 * 1024 // 50MB bounded limit
	}

	execCmd := exec.CommandContext(ctx, cmd, args...)

	stdoutPipe, err := execCmd.StdoutPipe()
	if err != nil {
		return "", -1, fmt.Errorf("failed to open stdout pipe: %w", err)
	}

	var stderrBuf bytes.Buffer
	execCmd.Stderr = &stderrBuf

	if err := execCmd.Start(); err != nil {
		return "", -1, fmt.Errorf("failed to start process: %w", err)
	}

	limitReader := io.LimitReader(stdoutPipe, maxBytes)
	var stdoutBuf bytes.Buffer
	_, copyErr := io.Copy(&stdoutBuf, limitReader)

	waitErr := execCmd.Wait()

	exitCode := 0
	if waitErr != nil {
		var exitErr *exec.ExitError
		if errors.As(waitErr, &exitErr) {
			exitCode = exitErr.ExitCode()
		} else {
			return stdoutBuf.String(), -1, waitErr
		}
	}

	if copyErr != nil {
		return stdoutBuf.String(), exitCode, fmt.Errorf("error reading stdout: %w", copyErr)
	}

	return stdoutBuf.String(), exitCode, nil
}
