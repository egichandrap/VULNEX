package cli

import (
	"errors"

	"github.com/vulnex/vulnex/installer/internal/domain"
)

// ResolveExitCode translates an execution error to a standard VULNEX CLI exit code (0..9).
func ResolveExitCode(err error) int {
	if err == nil {
		return int(domain.ExitSuccess)
	}

	var domErr *domain.DomainError
	if errors.As(err, &domErr) {
		return int(domErr.ExitCode())
	}

	return int(domain.ExitGeneralFailure)
}
