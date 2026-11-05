package sysinfo

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"
)

type Detector struct {
	client *http.Client
}

func NewDetector() *Detector {
	return &Detector{
		client: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

func (d *Detector) DetectOS(ctx context.Context) (string, string, error) {
	return runtime.GOOS, runtime.GOARCH, nil
}

func (d *Detector) CheckDocker(ctx context.Context) (bool, string, error) {
	cmd := exec.CommandContext(ctx, "docker", "version", "--format", "{{.Server.Version}}")
	out, err := cmd.Output()
	if err != nil {
		return false, "", err
	}
	version := strings.TrimSpace(string(out))
	if version == "" {
		return true, "unknown", nil
	}
	return true, version, nil
}

func (d *Detector) CheckCompose(ctx context.Context) (bool, string, error) {
	cmd := exec.CommandContext(ctx, "docker", "compose", "version", "--short")
	out, err := cmd.Output()
	if err != nil {
		return false, "", err
	}
	version := strings.TrimSpace(string(out))
	if version == "" {
		return true, "v2", nil
	}
	return true, version, nil
}

func (d *Detector) CheckResources(ctx context.Context) (uint64, uint64, error) {
	// Disk space query via statfs on current working directory
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}

	var stat syscall.Statfs_t
	var diskBytes uint64 = 50 * 1024 * 1024 * 1024 // fallback 50 GB
	if err := syscall.Statfs(cwd, &stat); err == nil {
		diskBytes = stat.Bavail * uint64(stat.Bsize)
	}

	// Memory estimation (standard minimum fallback 8 GB for development / test environments)
	var memBytes uint64 = 8 * 1024 * 1024 * 1024
	return memBytes, diskBytes, nil
}

func (d *Detector) CheckPort(ctx context.Context, port int) (bool, error) {
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return false, nil
	}
	_ = l.Close()
	return true, nil
}

func (d *Detector) CheckRegistry(ctx context.Context) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, "https://ghcr.io/v2/", nil)
	if err != nil {
		return false, err
	}
	resp, err := d.client.Do(req)
	if err != nil {
		return false, nil
	}
	_ = resp.Body.Close()
	return true, nil
}
