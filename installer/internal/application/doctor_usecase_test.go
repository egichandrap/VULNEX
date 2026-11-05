package application_test

import (
	"context"
	"testing"

	"github.com/vulnex/vulnex/installer/internal/application"
	"github.com/vulnex/vulnex/installer/internal/domain"
)

type mockEnvDetector struct {
	os        string
	arch      string
	dockerOk  bool
	dockerVer string
	dockerErr error
	composeOk bool
	compVer   string
	compErr   error
	memBytes  uint64
	diskBytes uint64
	portOk    bool
	regOk     bool
}

func (m *mockEnvDetector) DetectOS(ctx context.Context) (string, string, error) {
	return m.os, m.arch, nil
}
func (m *mockEnvDetector) CheckDocker(ctx context.Context) (bool, string, error) {
	return m.dockerOk, m.dockerVer, m.dockerErr
}
func (m *mockEnvDetector) CheckCompose(ctx context.Context) (bool, string, error) {
	return m.composeOk, m.compVer, m.compErr
}
func (m *mockEnvDetector) CheckResources(ctx context.Context) (uint64, uint64, error) {
	return m.memBytes, m.diskBytes, nil
}
func (m *mockEnvDetector) CheckPort(ctx context.Context, port int) (bool, error) {
	return m.portOk, nil
}
func (m *mockEnvDetector) CheckRegistry(ctx context.Context) (bool, error) {
	return m.regOk, nil
}

func TestDiagnoseEnvironmentUseCase_AllPass(t *testing.T) {
	detector := &mockEnvDetector{
		os:        "linux",
		arch:      "amd64",
		dockerOk:  true,
		dockerVer: "26.1.1",
		composeOk: true,
		compVer:   "v2.27.0",
		memBytes:  8 * 1024 * 1024 * 1024,
		diskBytes: 50 * 1024 * 1024 * 1024,
		portOk:    true,
		regOk:     true,
	}

	uc := application.NewDiagnoseEnvironmentUseCase(detector)
	report, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.Status != domain.CheckPass {
		t.Errorf("expected report status PASS, got %s", report.Status)
	}
	if report.HasFailures() {
		t.Error("expected 0 failures")
	}
}

func TestDiagnoseEnvironmentUseCase_DockerMissing(t *testing.T) {
	detector := &mockEnvDetector{
		os:        "linux",
		arch:      "amd64",
		dockerOk:  false,
		composeOk: false,
		memBytes:  8 * 1024 * 1024 * 1024,
		diskBytes: 50 * 1024 * 1024 * 1024,
		portOk:    true,
		regOk:     true,
	}

	uc := application.NewDiagnoseEnvironmentUseCase(detector)
	report, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.Status != domain.CheckFail {
		t.Errorf("expected report status FAIL, got %s", report.Status)
	}
	if !report.HasFailures() {
		t.Error("expected report to have failures")
	}
}
