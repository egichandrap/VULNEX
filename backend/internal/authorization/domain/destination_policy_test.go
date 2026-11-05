package domain

import (
	"net"
	"testing"
)

func TestNetworkDestinationPolicy_ProhibitedDestinations(t *testing.T) {
	policy := NetworkDestinationPolicy{}

	prohibitedTests := []struct {
		name         string
		ip           string
		allowPrivate bool
		expectedErr  error
	}{
		// Cloud Metadata endpoint
		{"AWS/GCP/Azure Metadata IP", "169.254.169.254", true, ErrCloudMetadataProhibited},
		{"AWS/GCP/Azure Metadata IP with private=false", "169.254.169.254", false, ErrCloudMetadataProhibited},

		// Loopback addresses (IPv4 & IPv6)
		{"IPv4 Loopback standard", "127.0.0.1", true, ErrLoopbackProhibited},
		{"IPv4 Loopback subnet", "127.0.0.100", false, ErrLoopbackProhibited},
		{"IPv6 Loopback", "::1", true, ErrLoopbackProhibited},

		// IPv4-mapped IPv6 pointing to loopback
		{"IPv4-mapped IPv6 Loopback", "::ffff:127.0.0.1", true, ErrLoopbackProhibited},

		// Link-Local addresses
		{"IPv4 Link-Local", "169.254.1.1", false, ErrLinkLocalProhibited},
		{"IPv6 Link-Local", "fe80::1", true, ErrLinkLocalProhibited},

		// Multicast / Broadcast
		{"IPv4 Multicast", "224.0.0.1", true, ErrMulticastProhibited},
		{"IPv6 Multicast", "ff02::1", false, ErrMulticastProhibited},
		{"IPv4 Broadcast", "255.255.255.255", false, ErrBroadcastProhibited},
	}

	for _, tt := range prohibitedTests {
		t.Run(tt.name, func(t *testing.T) {
			parsedIP := net.ParseIP(tt.ip)
			if parsedIP == nil {
				t.Fatalf("Failed to parse test IP %s", tt.ip)
			}

			err := policy.ValidateDestination(parsedIP, tt.allowPrivate)
			if err != tt.expectedErr {
				t.Errorf("ValidateDestination(%s, allowPrivate=%v) = %v; want %v", tt.ip, tt.allowPrivate, err, tt.expectedErr)
			}
		})
	}
}

func TestNetworkDestinationPolicy_PrivateNetworkGating(t *testing.T) {
	policy := NetworkDestinationPolicy{}

	privateTests := []struct {
		name string
		ip   string
	}{
		{"RFC 1918 10.0.0.0/8", "10.0.50.1"},
		{"RFC 1918 172.16.0.0/12", "172.20.10.5"},
		{"RFC 1918 192.168.0.0/16", "192.168.1.254"},
		{"IPv6 Unique Local fc00::/7", "fd00::1"},
	}

	for _, tt := range privateTests {
		t.Run(tt.name, func(t *testing.T) {
			parsedIP := net.ParseIP(tt.ip)
			if parsedIP == nil {
				t.Fatalf("Failed to parse test IP %s", tt.ip)
			}

			// When allowPrivate is false -> must fail with ErrPrivateNetworkDeniedByDefault
			err := policy.ValidateDestination(parsedIP, false)
			if err != ErrPrivateNetworkDeniedByDefault {
				t.Errorf("ValidateDestination(%s, allowPrivate=false) = %v; want %v", tt.ip, err, ErrPrivateNetworkDeniedByDefault)
			}

			// When allowPrivate is true -> must be permitted
			err = policy.ValidateDestination(parsedIP, true)
			if err != nil {
				t.Errorf("ValidateDestination(%s, allowPrivate=true) unexpected error: %v", tt.ip, err)
			}
		})
	}
}

func TestNetworkDestinationPolicy_PublicDestinationsAlwaysAllowed(t *testing.T) {
	policy := NetworkDestinationPolicy{}

	publicIPs := []string{
		"93.184.216.34",        // example.com
		"8.8.8.8",              // Google DNS
		"1.1.1.1",              // Cloudflare DNS
		"2606:4700:4700::1111", // Cloudflare IPv6
	}

	for _, ipStr := range publicIPs {
		parsedIP := net.ParseIP(ipStr)
		if parsedIP == nil {
			t.Fatalf("Failed to parse IP %s", ipStr)
		}

		err := policy.ValidateDestination(parsedIP, false)
		if err != nil {
			t.Errorf("Public destination %s should be allowed even with allowPrivate=false, got: %v", ipStr, err)
		}
	}
}
