package domain

import (
	"testing"
)

func TestCanonicalAssetValue_Domain(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		wantErr  bool
	}{
		{"SECURITY.Example.Com", "security.example.com", false},
		{"vulnex.org.", "vulnex.org", false},
		{"sub-domain.test.co.uk", "sub-domain.test.co.uk", false},
		{"http://example.com", "", true}, // protocol invalid for domain type
		{"-invalid.domain", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := NewCanonicalAssetValue(AssetTypeDomain, tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error for %q, got nil", tt.input)
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error for %q: %v", tt.input, err)
				}
				if got.String() != tt.expected {
					t.Errorf("Canonical value for %q = %q, want %q", tt.input, got.String(), tt.expected)
				}
			}
		})
	}
}

func TestCanonicalAssetValue_IP(t *testing.T) {
	tests := []struct {
		assetType AssetType
		input     string
		expected  string
		wantErr   bool
	}{
		{AssetTypeIPv4, "192.168.1.1", "192.168.1.1", false},
		{AssetTypeIPv4, "010.000.000.001", "", true}, // ambiguous leading zeros rejected to prevent octal SSRF
		{AssetTypeIPv4, "999.999.999.999", "", true},
		{AssetTypeIPv4, "2001:db8::1", "", true},                                         // IPv6 passed as IPv4
		{AssetTypeIPv6, "2001:0db8:0000:0000:0000:0000:0000:0001", "2001:db8::1", false}, // normalized compressed
		{AssetTypeIPv6, "::1", "::1", false},
		{AssetTypeIPv6, "not-an-ip", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := NewCanonicalAssetValue(tt.assetType, tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error for %q, got nil", tt.input)
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error for %q: %v", tt.input, err)
				}
				if got.String() != tt.expected {
					t.Errorf("Canonical IP for %q = %q, want %q", tt.input, got.String(), tt.expected)
				}
			}
		})
	}
}

func TestCanonicalAssetValue_URL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		wantErr  bool
	}{
		{"HTTPS://Example.Com:8443/Path/To/Resource?query=1#frag", "https://example.com:8443/Path/To/Resource", false},
		{"http://api.vulnex.local/", "http://api.vulnex.local", false}, // trailing slash normalized
		{"ftp://files.example.com", "", true},                          // non-http scheme rejected
		{"just-a-path/without-scheme", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := NewCanonicalAssetValue(AssetTypeURL, tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error for %q, got nil", tt.input)
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error for %q: %v", tt.input, err)
				}
				if got.String() != tt.expected {
					t.Errorf("Canonical URL for %q = %q, want %q", tt.input, got.String(), tt.expected)
				}
			}
		})
	}
}

func TestNewAsset_CreationAndInvariants(t *testing.T) {
	projectID, _ := GenerateProjectID()
	canonicalVal, err := NewCanonicalAssetValue(AssetTypeDomain, "App.Vulnex.Org")
	if err != nil {
		t.Fatalf("Failed to create canonical value: %v", err)
	}

	asset, err := NewAsset(projectID, AssetTypeDomain, canonicalVal, "Production Web App", CriticalityHigh)
	if err != nil {
		t.Fatalf("Failed to create asset: %v", err)
	}

	if asset.CanonicalValue().String() != "app.vulnex.org" {
		t.Errorf("Expected lowercase canonical value, got: %s", asset.CanonicalValue().String())
	}
	if asset.Criticality() != CriticalityHigh {
		t.Errorf("Expected CriticalityHigh, got: %s", asset.Criticality())
	}
	if asset.Status() != AssetStatusActive {
		t.Errorf("Expected AssetStatusActive, got: %s", asset.Status())
	}

	// Archive
	asset.Archive()
	if asset.Status() != AssetStatusArchived {
		t.Errorf("Expected status to be archived, got: %s", asset.Status())
	}
}
