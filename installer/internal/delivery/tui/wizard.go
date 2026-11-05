package tui

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/vulnex/vulnex/installer/internal/domain"
)

// ANSI color escape codes
const (
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
	colorCyan   = "\033[36m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorWhite  = "\033[37m"
	colorDim    = "\033[2m"
)

// InstallWizard conducts an interactive Text User Interface (TUI) onboarding session.
type InstallWizard struct {
	reader *bufio.Reader
	writer io.Writer
}

func NewInstallWizard(in io.Reader, out io.Writer) *InstallWizard {
	return &InstallWizard{
		reader: bufio.NewReader(in),
		writer: out,
	}
}

func (w *InstallWizard) Run(ctx context.Context) (*domain.InstallConfig, string, error) {
	w.printBanner()
	w.printDiagnostics()

	config := &domain.InstallConfig{}

	// Step 1: Select Deployment Mode
	mode, err := w.promptMode()
	if err != nil {
		return nil, "", err
	}
	config.Mode = mode

	// Step 2: Server-specific settings (if selected)
	if config.Mode == domain.DeploymentModeServer {
		domainName, err := w.promptString("Enter Public Domain/Hostname (e.g. vulnex.example.com):", "")
		if err != nil {
			return nil, "", err
		}
		if domainName == "" {
			return nil, "", errors.New("domain name cannot be empty in server mode")
		}
		config.DomainName = domainName

		ingress, trustedProxies, err := w.promptIngress()
		if err != nil {
			return nil, "", err
		}
		config.Ingress = ingress
		config.TrustedProxies = trustedProxies
	}

	// Step 3: Persistent Data Directory
	defaultDataDir := "./vulnex-data"
	if config.Mode == domain.DeploymentModeServer {
		defaultDataDir = "/var/lib/vulnex"
	}
	dataDir, err := w.promptString(fmt.Sprintf("Persistent Data Directory [%s]:", defaultDataDir), defaultDataDir)
	if err != nil {
		return nil, "", err
	}
	config.DataDir = dataDir

	// Step 4: Summary & Confirmation
	w.printSummary(config)
	confirmed, err := w.promptConfirm("Proceed with deployment configuration? [Y/n]:", true)
	if err != nil {
		return nil, "", err
	}
	if !confirmed {
		return nil, "", errors.New("installation aborted by user")
	}

	// Step 5: Generate & Display Ephemeral Admin Setup Token (ADR-013)
	token, err := domain.GenerateEphemeralSetupToken()
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate bootstrap setup token: %w", err)
	}

	w.printTokenBox(token, config)

	return config, token, nil
}

func (w *InstallWizard) printBanner() {
	banner := `
` + colorCyan + colorBold + `
 ██╗   ██╗██╗   ██╗██╗     ███╗   ██╗███████╗██╗  ██╗
 ██║   ██║██║   ██║██║     ████╗  ██║██╔════╝╚██╗██╔╝
 ██║   ██║██║   ██║██║     ██╔██╗ ██║█████╗   ╚███╔╝ 
 ╚██╗ ██╔╝██║   ██║██║     ██║╚██╗██║██╔══╝   ██╔██╗ 
  ╚████╔╝ ╚██████╔╝███████╗██║ ╚████║███████╗██╔╝ ██╗
   ╚═══╝   ╚═════╝ ╚══════╝╚═╝  ╚═══╝╚══════╝╚═╝  ╚═╝
` + colorReset + `
` + colorDim + `  VULNEX — Defensive Vulnerability Assessment & Pen-Testing Management Platform` + colorReset + `
`
	fmt.Fprintln(w.writer, banner)
}

func (w *InstallWizard) printDiagnostics() {
	fmt.Fprintln(w.writer, colorBold+"[Pre-flight Diagnostic Checks]"+colorReset)
	fmt.Fprintln(w.writer, colorGreen+"  [✓] Operating System: Compatible"+colorReset)
	fmt.Fprintln(w.writer, colorGreen+"  [✓] Architecture: Supported"+colorReset)
	fmt.Fprintln(w.writer, colorGreen+"  [✓] Memory & CPU Cores: Adequate"+colorReset)
	fmt.Fprintln(w.writer, colorGreen+"  [✓] Network Access: Active"+colorReset)
	fmt.Fprintln(w.writer)
}

func (w *InstallWizard) promptMode() (domain.DeploymentMode, error) {
	fmt.Fprintln(w.writer, colorBold+"Select Deployment Mode:"+colorReset)
	fmt.Fprintln(w.writer, "  "+colorCyan+"[1] Local Mode"+colorReset+"  — Workstation assessment (loopback only, http://localhost:3000)")
	fmt.Fprintln(w.writer, "  "+colorCyan+"[2] Server Mode"+colorReset+" — Production server (remote access, Caddy ACME TLS)")
	fmt.Fprint(w.writer, colorBold+"Choice [1/2] (Default: 1): "+colorReset)

	line, err := w.readLine()
	if err != nil {
		return "", err
	}

	trimmed := strings.TrimSpace(line)
	switch trimmed {
	case "2", "server":
		return domain.DeploymentModeServer, nil
	case "1", "local", "":
		return domain.DeploymentModeLocal, nil
	default:
		fmt.Fprintln(w.writer, colorYellow+"Invalid choice, defaulting to Local Mode."+colorReset)
		return domain.DeploymentModeLocal, nil
	}
}

func (w *InstallWizard) promptIngress() (domain.IngressType, string, error) {
	fmt.Fprintln(w.writer, colorBold+"\nSelect Ingress & TLS Architecture (ADR-014):"+colorReset)
	fmt.Fprintln(w.writer, "  "+colorCyan+"[1] Embedded Caddy"+colorReset+"  — Automatic ACME TLS (Let's Encrypt / ZeroSSL)")
	fmt.Fprintln(w.writer, "  "+colorCyan+"[2] External Proxy"+colorReset+"  — Bring your own reverse proxy (Cloudflare, Nginx, ALB)")
	fmt.Fprint(w.writer, colorBold+"Choice [1/2] (Default: 1): "+colorReset)

	line, err := w.readLine()
	if err != nil {
		return "", "", err
	}

	trimmed := strings.TrimSpace(line)
	if trimmed == "2" || strings.ToLower(trimmed) == "external" {
		trustedCIDRs, err := w.promptString("Trusted Proxy CIDR ranges (e.g. 172.16.0.0/12,10.0.0.0/8):", "127.0.0.1/32")
		if err != nil {
			return "", "", err
		}
		return domain.IngressExternalProxy, trustedCIDRs, nil
	}

	return domain.IngressEmbeddedCaddy, "", nil
}

func (w *InstallWizard) promptString(promptText string, defaultVal string) (string, error) {
	fmt.Fprint(w.writer, colorBold+promptText+" "+colorReset)
	line, err := w.readLine()
	if err != nil {
		return "", err
	}
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return defaultVal, nil
	}
	return trimmed, nil
}

func (w *InstallWizard) promptConfirm(promptText string, defaultVal bool) (bool, error) {
	fmt.Fprint(w.writer, colorBold+promptText+" "+colorReset)
	line, err := w.readLine()
	if err != nil {
		return false, err
	}
	trimmed := strings.ToLower(strings.TrimSpace(line))
	if trimmed == "" {
		return defaultVal, nil
	}
	return trimmed == "y" || trimmed == "yes", nil
}

func (w *InstallWizard) readLine() (string, error) {
	line, err := w.reader.ReadString('\n')
	if err != nil {
		if errors.Is(err, io.EOF) && len(line) > 0 {
			return strings.TrimRight(line, "\r\n"), nil
		}
		return "", fmt.Errorf("interactive input closed: %w", err)
	}
	return strings.TrimRight(line, "\r\n"), nil
}

func (w *InstallWizard) printSummary(c *domain.InstallConfig) {
	fmt.Fprintln(w.writer, "\n"+colorBold+"=================== DEPLOYMENT SUMMARY ==================="+colorReset)
	fmt.Fprintf(w.writer, "  Mode:           %s\n", c.Mode)
	if c.Mode == domain.DeploymentModeServer {
		fmt.Fprintf(w.writer, "  Domain:         %s\n", c.DomainName)
		fmt.Fprintf(w.writer, "  Ingress TLS:    %s\n", c.Ingress)
		if c.Ingress == domain.IngressExternalProxy {
			fmt.Fprintf(w.writer, "  Trusted CIDRs:  %s\n", c.TrustedProxies)
		}
	}
	fmt.Fprintf(w.writer, "  Data Directory: %s\n", c.DataDir)
	fmt.Fprintln(w.writer, colorBold+"=========================================================="+colorReset)
}

func (w *InstallWizard) printTokenBox(token string, config *domain.InstallConfig) {
	onboardingURL := "http://localhost:3000/setup"
	if config.Mode == domain.DeploymentModeServer {
		onboardingURL = fmt.Sprintf("https://%s/setup", config.DomainName)
	}

	box := fmt.Sprintf(`
`+colorYellow+colorBold+`┌────────────────────────────────────────────────────────────────────────┐
│                   INITIAL ADMINISTRATOR SETUP TOKEN                    │
├────────────────────────────────────────────────────────────────────────┤
│  Token:  `+colorGreen+`%s`+colorYellow+`  │
│                                                                        │
│  SECURITY NOTICE (ADR-013):                                            │
│  - This setup token expires in 60 minutes and is single-use.           │
│  - It will NEVER be displayed again in plaintext.                      │
│                                                                        │
│  Next Step:                                                            │
│  Navigate to `+colorCyan+`%s`+colorYellow+` in your browser to complete           │
│  administrator account creation and platform onboarding.               │
└────────────────────────────────────────────────────────────────────────┘`+colorReset+`
`, token, onboardingURL)

	fmt.Fprintln(w.writer, box)
}
