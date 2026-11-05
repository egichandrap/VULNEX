package domain

import "fmt"

// Standardized VULNEX CLI Exit Codes per Specification Section 9
type ExitCode int

const (
	ExitSuccess                  ExitCode = 0
	ExitGeneralFailure           ExitCode = 1
	ExitInvalidConfig            ExitCode = 2
	ExitUnsupportedEnv           ExitCode = 3
	ExitRuntimeUnavailable       ExitCode = 4
	ExitArtifactVerificationFail ExitCode = 5
	ExitDeploymentFailed         ExitCode = 6
	ExitBackupRestoreFailed      ExitCode = 7
	ExitUpgradeFailed            ExitCode = 8
	ExitCancelled                ExitCode = 9
)

// DomainError represents a typed error with an associated CLI exit code.
type DomainError struct {
	Code    ExitCode
	Message string
	Err     error
}

func NewError(code ExitCode, message string) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
	}
}

func WrapError(code ExitCode, message string, err error) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

func (e *DomainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *DomainError) ExitCode() ExitCode {
	return e.Code
}

func (e *DomainError) Unwrap() error {
	return e.Err
}
