package application

import (
	"context"
	"fmt"

	"github.com/vulnex/vulnex/installer/internal/application/ports"
	"github.com/vulnex/vulnex/installer/internal/domain"
)

const (
	minMemoryBytes = 2 * 1024 * 1024 * 1024  // 2 GB min
	recMemoryBytes = 4 * 1024 * 1024 * 1024  // 4 GB rec
	minDiskBytes   = 2 * 1024 * 1024 * 1024  // 2 GB min (container footprint ~500 MB)
	recDiskBytes   = 5 * 1024 * 1024 * 1024  // 5 GB rec
)

// DiagnoseEnvironmentUseCase evaluates host environment, container runtime, and resources.
type DiagnoseEnvironmentUseCase struct {
	detector ports.EnvironmentDetector
}

func NewDiagnoseEnvironmentUseCase(detector ports.EnvironmentDetector) *DiagnoseEnvironmentUseCase {
	return &DiagnoseEnvironmentUseCase{detector: detector}
}

func (uc *DiagnoseEnvironmentUseCase) Execute(ctx context.Context) (*domain.DiagnosticReport, error) {
	report := domain.NewDiagnosticReport()

	// 1. OS & Architecture Check
	osName, arch, err := uc.detector.DetectOS(ctx)
	if err != nil {
		report.AddCheck(domain.CheckResult{
			Name:        "Operating System",
			Status:      domain.CheckFail,
			Message:     fmt.Sprintf("Failed to detect OS: %v", err),
			Remediation: "Verify standard system environment",
		})
	} else {
		isOSSupported := (osName == "linux" || osName == "darwin")
		isArchSupported := (arch == "amd64" || arch == "arm64")

		if isOSSupported && isArchSupported {
			report.AddCheck(domain.CheckResult{
				Name:    "Operating System & Architecture",
				Status:  domain.CheckPass,
				Message: fmt.Sprintf("%s/%s (supported)", osName, arch),
			})
		} else {
			report.AddCheck(domain.CheckResult{
				Name:        "Operating System & Architecture",
				Status:      domain.CheckFail,
				Message:     fmt.Sprintf("%s/%s is not officially supported", osName, arch),
				Remediation: "Run VULNEX on Linux (amd64/arm64) or macOS (amd64/arm64)",
			})
		}
	}

	// 2. Docker Engine Check
	dockerOk, dockerVer, dockerErr := uc.detector.CheckDocker(ctx)
	if !dockerOk {
		msg := "Docker Engine is not running or not installed"
		remed := "Install and start Docker Engine (https://docs.docker.com/engine/install/)"
		if dockerErr != nil {
			msg = fmt.Sprintf("Docker daemon error: %v", dockerErr)
			remed = "Ensure user has permission to /var/run/docker.sock (`sudo usermod -aG docker $USER`)"
		}
		report.AddCheck(domain.CheckResult{
			Name:        "Docker Engine",
			Status:      domain.CheckFail,
			Message:     msg,
			Remediation: remed,
		})
	} else {
		report.AddCheck(domain.CheckResult{
			Name:    "Docker Engine",
			Status:  domain.CheckPass,
			Message: fmt.Sprintf("Running (v%s)", dockerVer),
		})
	}

	// 3. Docker Compose v2 Check
	composeOk, composeVer, _ := uc.detector.CheckCompose(ctx)
	if !composeOk {
		report.AddCheck(domain.CheckResult{
			Name:        "Docker Compose Plugin",
			Status:      domain.CheckFail,
			Message:     "Docker Compose v2 plugin is missing",
			Remediation: "Install docker-compose-plugin (https://docs.docker.com/compose/install/)",
		})
	} else {
		report.AddCheck(domain.CheckResult{
			Name:    "Docker Compose Plugin",
			Status:  domain.CheckPass,
			Message: fmt.Sprintf("Available (%s)", composeVer),
		})
	}

	// 4. Memory & Disk Resources
	memBytes, diskBytes, resErr := uc.detector.CheckResources(ctx)
	if resErr != nil {
		report.AddCheck(domain.CheckResult{
			Name:    "Hardware Resources",
			Status:  domain.CheckWarn,
			Message: fmt.Sprintf("Could not query resource statistics: %v", resErr),
		})
	} else {
		memGB := float64(memBytes) / (1024 * 1024 * 1024)
		if memBytes < minMemoryBytes {
			report.AddCheck(domain.CheckResult{
				Name:        "System Memory",
				Status:      domain.CheckFail,
				Message:     fmt.Sprintf("Insufficient RAM: %.1f GB (min: 2 GB)", memGB),
				Remediation: "Allocate at least 2 GB of memory to run services",
			})
		} else if memBytes < recMemoryBytes {
			report.AddCheck(domain.CheckResult{
				Name:        "System Memory",
				Status:      domain.CheckWarn,
				Message:     fmt.Sprintf("%.1f GB RAM available (recommended: 4 GB)", memGB),
				Remediation: "Heavy security assessments may require additional RAM",
			})
		} else {
			report.AddCheck(domain.CheckResult{
				Name:    "System Memory",
				Status:  domain.CheckPass,
				Message: fmt.Sprintf("%.1f GB RAM available", memGB),
			})
		}

		diskGB := float64(diskBytes) / (1024 * 1024 * 1024)
		if diskBytes < minDiskBytes {
			report.AddCheck(domain.CheckResult{
				Name:        "Available Disk Space",
				Status:      domain.CheckFail,
				Message:     fmt.Sprintf("Insufficient disk space: %.1f GB (min: %d GB)", diskGB, minDiskBytes/(1024*1024*1024)),
				Remediation: "Free up storage space on the installation drive",
			})
		} else if diskBytes < recDiskBytes {
			report.AddCheck(domain.CheckResult{
				Name:        "Available Disk Space",
				Status:      domain.CheckWarn,
				Message:     fmt.Sprintf("%.1f GB available (recommended: %d GB)", diskGB, recDiskBytes/(1024*1024*1024)),
				Remediation: "Additional disk space recommended for ongoing vulnerability scans and reporting data",
			})
		} else {
			report.AddCheck(domain.CheckResult{
				Name:    "Available Disk Space",
				Status:  domain.CheckPass,
				Message: fmt.Sprintf("%.1f GB available", diskGB),
			})
		}
	}

	// 5. Port Availability (Dashboard 3000)
	port3000Ok, _ := uc.detector.CheckPort(ctx, 3000)
	if !port3000Ok {
		report.AddCheck(domain.CheckResult{
			Name:        "Port 3000 (Dashboard)",
			Status:      domain.CheckWarn,
			Message:     "Port 3000 is already in use by another process",
			Remediation: "Stop conflicting service or configure alternate WEB_PORT",
		})
	} else {
		report.AddCheck(domain.CheckResult{
			Name:    "Port 3000 (Dashboard)",
			Status:  domain.CheckPass,
			Message: "Port 3000 is free",
		})
	}

	// 6. Registry Network Connectivity
	regOk, _ := uc.detector.CheckRegistry(ctx)
	if !regOk {
		report.AddCheck(domain.CheckResult{
			Name:        "Container Registry Connectivity",
			Status:      domain.CheckWarn,
			Message:     "Unable to ping public registry endpoints",
			Remediation: "Check firewall / internet connectivity for image pulling",
		})
	} else {
		report.AddCheck(domain.CheckResult{
			Name:    "Container Registry Connectivity",
			Status:  domain.CheckPass,
			Message: "Registry endpoints are reachable",
		})
	}

	return report, nil
}
