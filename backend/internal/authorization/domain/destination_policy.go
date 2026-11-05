package domain

import (
	"errors"
	"net"
)

var (
	ErrLoopbackProhibited            = errors.New("destination is a loopback address and is strictly prohibited")
	ErrLinkLocalProhibited           = errors.New("destination is a link-local address and is strictly prohibited")
	ErrCloudMetadataProhibited       = errors.New("destination is cloud metadata endpoint (169.254.169.254) and is strictly prohibited")
	ErrMulticastProhibited           = errors.New("destination is a multicast address and is prohibited")
	ErrBroadcastProhibited           = errors.New("destination is a broadcast address and is prohibited")
	ErrPrivateNetworkDeniedByDefault = errors.New("destination is in a private network (RFC 1918 / RFC 4193) and is denied by default")
)

var (
	cloudMetadataIP = net.ParseIP("169.254.169.254")

	// RFC 1918 IPv4 Private Subnets
	privateCIDR10  = mustParseCIDR("10.0.0.0/8")
	privateCIDR172 = mustParseCIDR("172.16.0.0/12")
	privateCIDR192 = mustParseCIDR("192.168.0.0/16")

	// RFC 4193 IPv6 Unique Local Subnet
	privateCIDRFC00 = mustParseCIDR("fc00::/7")
)

func mustParseCIDR(s string) *net.IPNet {
	_, ipNet, err := net.ParseCIDR(s)
	if err != nil {
		panic("invalid static CIDR: " + s)
	}
	return ipNet
}

// NetworkDestinationPolicy is a pure domain service that enforces SSRF and DNS rebinding protections.
// Invariant: Cloud metadata, loopback, link-local, and multicast are permanently prohibited.
// Private networks are prohibited unless allowPrivate is true.
type NetworkDestinationPolicy struct{}

func (p NetworkDestinationPolicy) ValidateDestination(ip net.IP, allowPrivate bool) error {
	if ip == nil {
		return errors.New("destination IP cannot be nil")
	}

	// Unmap IPv4-mapped IPv6 (e.g. ::ffff:127.0.0.1 -> 127.0.0.1)
	unmappedIP := ip
	if v4 := ip.To4(); v4 != nil {
		unmappedIP = v4
	}

	// 1. Cloud Metadata Check (Permanent Prohibition)
	if unmappedIP.Equal(cloudMetadataIP) {
		return ErrCloudMetadataProhibited
	}

	// 2. Loopback Check (Permanent Prohibition)
	if unmappedIP.IsLoopback() {
		return ErrLoopbackProhibited
	}

	// 3. Multicast Check (Permanent Prohibition)
	if unmappedIP.IsMulticast() {
		return ErrMulticastProhibited
	}

	// 4. Link-Local Unicast Check (Permanent Prohibition)
	if unmappedIP.IsLinkLocalUnicast() || unmappedIP.IsLinkLocalMulticast() {
		return ErrLinkLocalProhibited
	}

	// 5. Broadcast Check
	if unmappedIP.Equal(net.IPv4bcast) {
		return ErrBroadcastProhibited
	}

	// 6. Private Network Check (Gated by Workspace Policy, ADR-015)
	if p.isPrivate(unmappedIP) {
		if !allowPrivate {
			return ErrPrivateNetworkDeniedByDefault
		}
	}

	return nil
}

func (p NetworkDestinationPolicy) isPrivate(ip net.IP) bool {
	if v4 := ip.To4(); v4 != nil {
		return privateCIDR10.Contains(v4) || privateCIDR172.Contains(v4) || privateCIDR192.Contains(v4)
	}
	return privateCIDRFC00.Contains(ip)
}
