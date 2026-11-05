package scanner

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/vulnex/vulnex/backend/internal/scanning/domain"
)

const sampleZapJSON = `{
  "@version": "2.14.0",
  "@generated": "Mon, 1 Jan 2026 12:00:00",
  "site": [
    {
      "@name": "https://target.example.com",
      "@host": "target.example.com",
      "@port": "443",
      "@ssl": "true",
      "alerts": [
        {
          "pluginid": "10038",
          "alertRef": "10038",
          "alert": "Content Security Policy (CSP) Header Not Set",
          "name": "Content Security Policy (CSP) Header Not Set",
          "riskcode": "2",
          "confidence": "2",
          "riskdesc": "Medium (High)",
          "desc": "Content Security Policy (CSP) is an added layer of security that helps to detect and mitigate certain types of attacks.",
          "instances": [
            {
              "uri": "https://target.example.com/login",
              "method": "GET",
              "param": "",
              "attack": "",
              "evidence": ""
            }
          ],
          "count": "1",
          "solution": "Ensure that your web server, load balancer, etc. sets the Content-Security-Policy header.",
          "reference": "https://developer.mozilla.org/en-US/docs/Web/HTTP/CSP",
          "cweid": "693",
          "wascid": "15"
        },
        {
          "pluginid": "10020",
          "alertRef": "10020",
          "alert": "Anti-CSRF Tokens Check",
          "name": "Anti-CSRF Tokens Check",
          "riskcode": "3",
          "confidence": "3",
          "riskdesc": "High (High)",
          "desc": "A form was found without an anti-CSRF token.",
          "instances": [
            {
              "uri": "https://target.example.com/transfer",
              "method": "POST"
            }
          ],
          "solution": "Include anti-CSRF token in all state-changing forms.",
          "cweid": "352"
        }
      ]
    }
  ]
}`

type mockZapCommandExecutor struct {
	output   string
	exitCode int
	err      error
}

func (m *mockZapCommandExecutor) Execute(ctx context.Context, cmd string, args []string, maxBytes int64) (string, int, error) {
	return m.output, m.exitCode, m.err
}

func TestZapAdapter_ProfileValidation(t *testing.T) {
	adapter := NewZapAdapter(&mockZapCommandExecutor{})

	validProfiles := []string{"zap-baseline", "baseline"}
	for _, p := range validProfiles {
		if err := adapter.ValidateProfile(context.Background(), p); err != nil {
			t.Errorf("expected profile %q to be valid, got: %v", p, err)
		}
	}

	invalidProfiles := []string{"deep-crawl", "aggressive-active", "", "zap-attack"}
	for _, p := range invalidProfiles {
		if err := adapter.ValidateProfile(context.Background(), p); err == nil {
			t.Errorf("expected profile %q to be rejected, got nil", p)
		}
	}
}

func TestZapAdapter_ArgumentValidation(t *testing.T) {
	adapter := NewZapAdapter(&mockZapCommandExecutor{})

	tests := []struct {
		name       string
		req        domain.ScanExecutionRequest
		wantErr    bool
		wantTarget string
	}{
		{
			name: "valid https target",
			req: domain.ScanExecutionRequest{
				TargetHost:  "https://target.example.com",
				ProfileName: "zap-baseline",
			},
			wantErr:    false,
			wantTarget: "https://target.example.com",
		},
		{
			name: "valid http target with path and port",
			req: domain.ScanExecutionRequest{
				TargetHost:  "http://api.local:8080/v1",
				ProfileName: "baseline",
			},
			wantErr:    false,
			wantTarget: "http://api.local:8080/v1",
		},
		{
			name: "empty target URL fails",
			req: domain.ScanExecutionRequest{
				TargetHost:  "",
				ProfileName: "zap-baseline",
			},
			wantErr: true,
		},
		{
			name: "invalid non-http scheme fails",
			req: domain.ScanExecutionRequest{
				TargetHost:  "ftp://target.example.com",
				ProfileName: "zap-baseline",
			},
			wantErr: true,
		},
		{
			name: "missing scheme fails",
			req: domain.ScanExecutionRequest{
				TargetHost:  "target.example.com",
				ProfileName: "zap-baseline",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args, err := adapter.BuildArguments(tt.req)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil args: %v", args)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			found := false
			for i, arg := range args {
				if arg == "-t" && i+1 < len(args) && args[i+1] == tt.wantTarget {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("arguments %v do not contain -t %s", args, tt.wantTarget)
			}
		})
	}
}

func TestZapAdapter_Execution(t *testing.T) {
	mock := &mockZapCommandExecutor{
		output:   sampleZapJSON,
		exitCode: 1, // ZAP returns 1 or 2 when findings are discovered (not a fatal system crash)
	}
	adapter := NewZapAdapter(mock)

	req := domain.ScanExecutionRequest{
		TargetHost:  "https://target.example.com",
		ProfileName: "zap-baseline",
		Timeout:     30 * time.Second,
	}

	artifact, err := adapter.ExecuteScan(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected execution error: %v", err)
	}

	if artifact.ScannerName != "zap" {
		t.Errorf("expected scanner name 'zap', got %s", artifact.ScannerName)
	}
	data, _ := io.ReadAll(artifact.OutputReader)
	if string(data) != sampleZapJSON {
		t.Errorf("expected matching payload")
	}
}

func TestZapParser_ExtractsAlertsSafely(t *testing.T) {
	parser := NewZapParser()
	obs, err := parser.Parse(strings.NewReader(sampleZapJSON), 5*1024*1024)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if len(obs) != 2 {
		t.Fatalf("expected 2 observations, got %d", len(obs))
	}

	csp := obs[0]
	if csp.Title != "Content Security Policy (CSP) Header Not Set" {
		t.Errorf("unexpected title: %s", csp.Title)
	}
	if csp.Severity != "MEDIUM" {
		t.Errorf("expected MEDIUM severity, got %s", csp.Severity)
	}
	if csp.SourceScanner != "zap" {
		t.Errorf("expected source scanner zap, got %s", csp.SourceScanner)
	}
	if csp.VulnerabilityKey != "ZAP-10038" {
		t.Errorf("expected VulnerabilityKey ZAP-10038, got %s", csp.VulnerabilityKey)
	}
	if csp.LocationKey != "https://target.example.com/login" {
		t.Errorf("expected LocationKey https://target.example.com/login, got %s", csp.LocationKey)
	}

	csrf := obs[1]
	if csrf.Severity != "HIGH" {
		t.Errorf("expected HIGH severity, got %s", csrf.Severity)
	}
	if csrf.VulnerabilityKey != "ZAP-10020" {
		t.Errorf("expected VulnerabilityKey ZAP-10020, got %s", csrf.VulnerabilityKey)
	}
}

func TestZapParser_SizeLimitAndTruncation(t *testing.T) {
	parser := NewZapParser()
	// Pass limit smaller than payload
	_, err := parser.Parse(strings.NewReader(sampleZapJSON), 50)
	if err == nil {
		t.Fatal("expected parse error on truncated JSON stream, got nil")
	}
}
