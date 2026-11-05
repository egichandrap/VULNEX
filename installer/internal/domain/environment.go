package domain

type CheckStatus string

const (
	CheckPass CheckStatus = "PASS"
	CheckWarn CheckStatus = "WARN"
	CheckFail CheckStatus = "FAIL"
)

// CheckResult represents an individual diagnostic assertion.
type CheckResult struct {
	Name        string      `json:"name"`
	Status      CheckStatus `json:"status"`
	Message     string      `json:"message,omitempty"`
	Remediation string      `json:"remediation,omitempty"`
}

// DiagnosticReport collects results from environment and prerequisite checks.
type DiagnosticReport struct {
	Status CheckStatus   `json:"status"`
	Checks []CheckResult `json:"checks"`
}

func NewDiagnosticReport() *DiagnosticReport {
	return &DiagnosticReport{
		Status: CheckPass,
		Checks: make([]CheckResult, 0),
	}
}

func (r *DiagnosticReport) AddCheck(c CheckResult) {
	r.Checks = append(r.Checks, c)
	if c.Status == CheckFail {
		r.Status = CheckFail
	} else if c.Status == CheckWarn && r.Status != CheckFail {
		r.Status = CheckWarn
	}
}

func (r *DiagnosticReport) HasWarnings() bool {
	for _, c := range r.Checks {
		if c.Status == CheckWarn {
			return true
		}
	}
	return false
}

func (r *DiagnosticReport) HasFailures() bool {
	for _, c := range r.Checks {
		if c.Status == CheckFail {
			return true
		}
	}
	return false
}
