package domain

import (
	"errors"
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
	Mode           DeploymentMode
	DomainName     string
	Ingress        IngressType
	TrustedProxies string
	DataDir        string
	AutoStart      bool
}

func (c *InstallConfig) Validate() error {
	if c.Mode != DeploymentModeLocal && c.Mode != DeploymentModeServer {
		return errors.New("invalid deployment mode: must be 'local' or 'server'")
	}
	if c.Mode == DeploymentModeServer && strings.TrimSpace(c.DomainName) == "" {
		return errors.New("server mode requires a domain name")
	}
	return nil
}
