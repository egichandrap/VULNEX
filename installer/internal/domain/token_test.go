package domain

import (
	"encoding/hex"
	"testing"
)

func TestGenerateEphemeralSetupToken(t *testing.T) {
	token, err := GenerateEphemeralSetupToken()
	if err != nil {
		t.Fatalf("unexpected token generation error: %v", err)
	}

	if len(token) != 64 {
		t.Errorf("expected 64 hex characters (32 bytes), got length %d", len(token))
	}

	decoded, err := hex.DecodeString(token)
	if err != nil {
		t.Fatalf("token is not valid hex: %v", err)
	}

	if len(decoded) != 32 {
		t.Errorf("expected 32 raw bytes, got %d", len(decoded))
	}

	token2, _ := GenerateEphemeralSetupToken()
	if token == token2 {
		t.Errorf("subsequent tokens must be unique and non-colliding")
	}
}

func TestInstallConfig_Validation(t *testing.T) {
	validLocal := InstallConfig{
		Mode:    DeploymentModeLocal,
		DataDir: "./vulnex-data",
	}
	if err := validLocal.Validate(); err != nil {
		t.Errorf("expected local config to be valid, got: %v", err)
	}

	validServer := InstallConfig{
		Mode:       DeploymentModeServer,
		DomainName: "vulnex.corp.internal",
		Ingress:    IngressEmbeddedCaddy,
	}
	if err := validServer.Validate(); err != nil {
		t.Errorf("expected server config with domain to be valid, got: %v", err)
	}

	serverMissingDomain := InstallConfig{
		Mode:       DeploymentModeServer,
		DomainName: "",
	}
	if err := serverMissingDomain.Validate(); err == nil {
		t.Error("expected error for server mode without domain name, got nil")
	}

	invalidMode := InstallConfig{
		Mode: "invalid-mode",
	}
	if err := invalidMode.Validate(); err == nil {
		t.Error("expected error for invalid mode, got nil")
	}
}
