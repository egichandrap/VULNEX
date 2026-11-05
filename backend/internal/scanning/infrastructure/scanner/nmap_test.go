package scanner

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/vulnex/vulnex/backend/internal/scanning/domain"
)

const sampleNmapXML = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE nmaprun>
<nmaprun scanner="nmap" version="7.94">
<host starttime="1700000000" endtime="1700000010">
    <status state="up" reason="syn-ack"/>
    <address addr="93.184.216.34" addrtype="ipv4"/>
    <hostnames>
        <hostname name="example.com" type="user"/>
    </hostnames>
    <ports>
        <port protocol="tcp" portid="80">
            <state state="open" reason="syn-ack"/>
            <service name="http" product="ECS" version="bastion" method="probed" conf="10"/>
        </port>
        <port protocol="tcp" portid="443">
            <state state="open" reason="syn-ack"/>
            <service name="https" product="ECS" version="bastion" tunnel="ssl" method="probed" conf="10"/>
        </port>
        <port protocol="tcp" portid="8080">
            <state state="closed" reason="reset"/>
            <service name="http-proxy" method="table" conf="3"/>
        </port>
    </ports>
</host>
</nmaprun>`

func TestNmapParser_ExtractsOpenPortsSafely(t *testing.T) {
	parser := NewNmapParser()

	observations, err := parser.Parse(strings.NewReader(sampleNmapXML), 10*1024*1024)
	if err != nil {
		t.Fatalf("Failed to parse sample Nmap XML: %v", err)
	}

	// Should extract 2 open ports (80 and 443), ignoring closed port 8080
	if len(observations) != 2 {
		t.Fatalf("Expected 2 open port observations, got %d", len(observations))
	}

	obs80 := observations[0]
	if obs80.LocationKey != "tcp/80" {
		t.Errorf("Expected location tcp/80, got %s", obs80.LocationKey)
	}
	if obs80.Severity != "INFO" {
		t.Errorf("Expected severity INFO for open port, got %s", obs80.Severity)
	}
	if !strings.Contains(obs80.Description, "http") {
		t.Errorf("Expected description to mention service http, got %s", obs80.Description)
	}

	obs443 := observations[1]
	if obs443.LocationKey != "tcp/443" {
		t.Errorf("Expected location tcp/443, got %s", obs443.LocationKey)
	}
}

func TestNmapParser_TruncatedOutputRejectedSafely(t *testing.T) {
	parser := NewNmapParser()

	// Truncated, malformed XML
	truncatedXML := `<nmaprun scanner="nmap"><host><ports><port protocol="tcp" portid="80">`
	_, err := parser.Parse(strings.NewReader(truncatedXML), 10*1024*1024)
	if err == nil {
		t.Errorf("Expected parser to return error on truncated XML, got nil")
	}
}

func TestNmapAdapter_ArgumentValidation(t *testing.T) {
	adapter := NewNmapAdapter(nil) // executor nil for argument building test

	// 1. Valid discovery profile
	args, err := adapter.BuildArguments(domain.ScanExecutionRequest{
		TargetHost:  "example.com",
		ProfileName: "discovery",
	})
	if err != nil {
		t.Fatalf("BuildArguments failed for discovery: %v", err)
	}
	argsStr := strings.Join(args, " ")
	if !strings.Contains(argsStr, "-sn") {
		t.Errorf("Expected discovery profile to include -sn, got: %s", argsStr)
	}

	// 2. Valid port scan profile with valid port range
	args, err = adapter.BuildArguments(domain.ScanExecutionRequest{
		TargetHost:  "example.com",
		ProfileName: "top-ports",
		PortRange:   "80,443,8080-8090",
	})
	if err != nil {
		t.Fatalf("BuildArguments failed: %v", err)
	}
	argsStr = strings.Join(args, " ")
	if !strings.Contains(argsStr, "-p 80,443,8080-8090") {
		t.Errorf("Expected port range flag, got: %s", argsStr)
	}

	// 3. Command injection attempt in port range must be rejected!
	_, err = adapter.BuildArguments(domain.ScanExecutionRequest{
		TargetHost:  "example.com",
		ProfileName: "top-ports",
		PortRange:   "80; cat /etc/passwd",
	})
	if err != ErrInvalidPortRange {
		t.Errorf("Expected ErrInvalidPortRange for injected command, got: %v", err)
	}

	// 4. Unknown profile must be rejected
	_, err = adapter.BuildArguments(domain.ScanExecutionRequest{
		TargetHost:  "example.com",
		ProfileName: "arbitrary-custom-profile",
	})
	if err != ErrUnsupportedProfile {
		t.Errorf("Expected ErrUnsupportedProfile, got: %v", err)
	}
}

func TestNmapAdapter_MockExecution(t *testing.T) {
	mockExec := &mockCommandExecutor{
		output: sampleNmapXML,
	}

	adapter := NewNmapAdapter(mockExec)

	scanID, _ := domain.GenerateScanID()
	req := domain.ScanExecutionRequest{
		ScanID:         scanID,
		TargetHost:     "93.184.216.34",
		ProfileName:    "discovery",
		Timeout:        5 * time.Second,
		MaxOutputBytes: 1024 * 1024,
	}

	artifact, err := adapter.ExecuteScan(context.Background(), req)
	if err != nil {
		t.Fatalf("ExecuteScan failed: %v", err)
	}

	if artifact.ScannerName != "nmap" {
		t.Errorf("Expected scanner name nmap, got: %s", artifact.ScannerName)
	}
	if artifact.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got: %d", artifact.ExitCode)
	}
}

type mockCommandExecutor struct {
	output string
}

func (m *mockCommandExecutor) Execute(ctx context.Context, cmd string, args []string, maxBytes int64) (string, int, error) {
	return m.output, 0, nil
}
