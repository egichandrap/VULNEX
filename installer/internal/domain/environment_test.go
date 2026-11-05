package domain_test

import (
	"testing"

	"github.com/vulnex/vulnex/installer/internal/domain"
)

func TestDiagnosticReport_Summary(t *testing.T) {
	report := domain.NewDiagnosticReport()
	report.AddCheck(domain.CheckResult{
		Name:   "OS Check",
		Status: domain.CheckPass,
	})
	report.AddCheck(domain.CheckResult{
		Name:        "Docker Check",
		Status:      domain.CheckWarn,
		Message:     "Daemon slow",
		Remediation: "Check daemon logs",
	})

	if !report.HasWarnings() {
		t.Error("expected report to have warnings")
	}
	if report.HasFailures() {
		t.Error("expected report not to have failures")
	}

	report.AddCheck(domain.CheckResult{
		Name:        "Disk Space",
		Status:      domain.CheckFail,
		Message:     "Only 200MB free",
		Remediation: "Free up at least 10GB",
	})

	if !report.HasFailures() {
		t.Error("expected report to have failures")
	}
	if report.Status != domain.CheckFail {
		t.Errorf("expected overall status fail, got %s", report.Status)
	}
}
