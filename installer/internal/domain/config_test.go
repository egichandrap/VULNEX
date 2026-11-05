package domain_test

import (
	"testing"

	"github.com/vulnex/vulnex/installer/internal/domain"
)

func TestInstallConfig_ValidationRules(t *testing.T) {
	tests := []struct {
		name    string
		config  domain.InstallConfig
		wantErr bool
	}{
		{
			name: "Valid local mode",
			config: domain.InstallConfig{
				Mode:    domain.DeploymentModeLocal,
				DataDir: "./vulnex-data",
				Version: "1.0.0",
			},
			wantErr: false,
		},
		{
			name: "Valid server mode with embedded Caddy",
			config: domain.InstallConfig{
				Mode:       domain.DeploymentModeServer,
				DomainName: "vulnex.corp.internal",
				Ingress:    domain.IngressEmbeddedCaddy,
				DataDir:    "/var/lib/vulnex",
				Version:    "1.0.0",
			},
			wantErr: false,
		},
		{
			name: "Valid server mode with external proxy and valid CIDR",
			config: domain.InstallConfig{
				Mode:           domain.DeploymentModeServer,
				DomainName:     "vulnex.corp.internal",
				Ingress:        domain.IngressExternalProxy,
				TrustedProxies: "172.16.0.0/12, 10.0.0.0/8",
				DataDir:        "/var/lib/vulnex",
				Version:        "1.0.0",
			},
			wantErr: false,
		},
		{
			name: "Server mode with invalid trusted proxies CIDR",
			config: domain.InstallConfig{
				Mode:           domain.DeploymentModeServer,
				DomainName:     "vulnex.corp.internal",
				Ingress:        domain.IngressExternalProxy,
				TrustedProxies: "not-a-valid-cidr",
				DataDir:        "/var/lib/vulnex",
			},
			wantErr: true,
		},
		{
			name: "Server mode missing domain name",
			config: domain.InstallConfig{
				Mode:    domain.DeploymentModeServer,
				DataDir: "/var/lib/vulnex",
			},
			wantErr: true,
		},
		{
			name: "Empty deployment mode",
			config: domain.InstallConfig{
				Mode: "",
			},
			wantErr: true,
		},
		{
			name: "Invalid deployment mode",
			config: domain.InstallConfig{
				Mode: "cloud-saas",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInstallConfig_Redacted(t *testing.T) {
	cfg := domain.InstallConfig{
		Mode:             domain.DeploymentModeLocal,
		DatabasePassword: "supersecretpassword123!",
		DataDir:          "./vulnex-data",
	}

	redacted := cfg.Redacted()
	if redacted.DatabasePassword == "supersecretpassword123!" {
		t.Errorf("expected DatabasePassword to be redacted, got plaintext")
	}
	if redacted.DatabasePassword != "********" {
		t.Errorf("expected '********', got %s", redacted.DatabasePassword)
	}
	// Original must not be mutated
	if cfg.DatabasePassword != "supersecretpassword123!" {
		t.Errorf("expected original password to be preserved, got %s", cfg.DatabasePassword)
	}
}
