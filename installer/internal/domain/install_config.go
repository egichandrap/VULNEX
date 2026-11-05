package domain

import (
	"errors"
	"net"
	"strings"
)

type DeploymentMode string

const (
	DeploymentModeLocal  DeploymentMode = "local"
	DeploymentModeServer DeploymentMode = "server"
)

type IngressType string

const (
	IngressEmbeddedCaddy IngressType = "caddy"
	IngressExternalProxy IngressType = "external"
)

// InstallConfig holds user selections for provisioning VULNEX.
type InstallConfig struct {
	Mode             DeploymentMode `json:"mode"`
	DomainName       string         `json:"domain_name,omitempty"`
	Ingress          IngressType    `json:"ingress,omitempty"`
	TrustedProxies   string         `json:"trusted_proxies,omitempty"`
	DataDir          string         `json:"data_dir"`
	Version          string         `json:"version,omitempty"`
	DatabasePassword string         `json:"database_password,omitempty"`
	WebPort          int            `json:"web_port,omitempty"`
	APIPort          int            `json:"api_port,omitempty"`
	AutoStart        bool           `json:"auto_start,omitempty"`
}

// Validate ensures all configuration invariants are satisfied.
func (c *InstallConfig) Validate() error {
	if c.Mode != DeploymentModeLocal && c.Mode != DeploymentModeServer {
		return errors.New("invalid deployment mode: must be 'local' or 'server'")
	}
	if c.Mode == DeploymentModeServer && strings.TrimSpace(c.DomainName) == "" {
		return errors.New("server mode requires a domain name")
	}

	if c.Mode == DeploymentModeServer && c.Ingress == IngressExternalProxy && strings.TrimSpace(c.TrustedProxies) != "" {
		cidrs := strings.Split(c.TrustedProxies, ",")
		for _, raw := range cidrs {
			trimmed := strings.TrimSpace(raw)
			if trimmed == "" {
				continue
			}
			if _, _, err := net.ParseCIDR(trimmed); err != nil {
				// Allow single IP address as well
				if ip := net.ParseIP(trimmed); ip == nil {
					return errors.New("invalid trusted proxy CIDR or IP: " + trimmed)
				}
			}
		}
	}

	return nil
}

// Redacted returns a deep copy with sensitive credentials obscured.
func (c InstallConfig) Redacted() *InstallConfig {
	copyCfg := c
	if copyCfg.DatabasePassword != "" {
		copyCfg.DatabasePassword = "********"
	}
	return &copyCfg
}
