package domain

import (
	"net"
	"strings"
)

// ScopePolicy is a pure domain service evaluating target authorization rules.
// Invariant: EXCLUDE rules unconditionally override INCLUDE rules.
type ScopePolicy struct{}

// IsTargetAuthorized evaluates target against rules.
func (p ScopePolicy) IsTargetAuthorized(target string, rules []ScopeRule) bool {
	cleanTarget := strings.ToLower(strings.TrimSpace(target))
	if cleanTarget == "" {
		return false
	}

	included := false

	for _, rule := range rules {
		matches := p.matchesRule(cleanTarget, rule)
		if matches {
			if rule.RuleType == RuleTypeExclude {
				// Exclusion takes immediate and final precedence
				return false
			}
			if rule.RuleType == RuleTypeInclude {
				included = true
			}
		}
	}

	return included
}

func (p ScopePolicy) matchesRule(target string, rule ScopeRule) bool {
	cleanRuleTarget := strings.ToLower(strings.TrimSpace(rule.CanonicalTarget))

	switch rule.TargetType {
	case ScopeTargetDomain:
		// Exact match or subdomain
		if target == cleanRuleTarget {
			return true
		}
		if strings.HasSuffix(target, "."+cleanRuleTarget) {
			return true
		}
		return false

	case ScopeTargetIP:
		// Exact IP equality
		targetIP := net.ParseIP(target)
		ruleIP := net.ParseIP(cleanRuleTarget)
		if targetIP != nil && ruleIP != nil {
			return targetIP.Equal(ruleIP)
		}
		return target == cleanRuleTarget

	case ScopeTargetCIDR:
		// CIDR subnet membership
		targetIP := net.ParseIP(target)
		if targetIP == nil {
			return false
		}
		_, ipNet, err := net.ParseCIDR(cleanRuleTarget)
		if err != nil {
			return false
		}
		return ipNet.Contains(targetIP)

	case ScopeTargetURL:
		// URL prefix match
		return strings.HasPrefix(target, cleanRuleTarget)

	default:
		return target == cleanRuleTarget
	}
}
