package domain

import (
	"crypto/rand"
	"errors"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var (
	ErrInvalidAssetType      = errors.New("invalid or unsupported asset type")
	ErrInvalidAssetFormat    = errors.New("asset value fails canonical format validation")
	ErrEmptyCanonicalValue   = errors.New("canonical asset value cannot be empty")
	ErrAssetAlreadyArchived  = errors.New("asset is already archived")
)

var domainRegex = regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)

type AssetID struct {
	value string
}

func NewAssetID(val string) (AssetID, error) {
	if strings.TrimSpace(val) == "" {
		return AssetID{}, errors.New("asset ID cannot be empty")
	}
	return AssetID{value: strings.TrimSpace(val)}, nil
}

func GenerateAssetID() (AssetID, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return AssetID{}, fmt.Errorf("failed to generate UUID: %w", err)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	val := fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
	return AssetID{value: val}, nil
}

func (a AssetID) String() string {
	return a.value
}

type ProjectID struct {
	value string
}

func NewProjectID(val string) (ProjectID, error) {
	if strings.TrimSpace(val) == "" {
		return ProjectID{}, errors.New("project ID cannot be empty")
	}
	return ProjectID{value: strings.TrimSpace(val)}, nil
}

func GenerateProjectID() (ProjectID, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return ProjectID{}, fmt.Errorf("failed to generate UUID: %w", err)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	val := fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
	return ProjectID{value: val}, nil
}

func (p ProjectID) String() string {
	return p.value
}

type AssetType string

const (
	AssetTypeDomain   AssetType = "DOMAIN"
	AssetTypeHostname AssetType = "HOSTNAME"
	AssetTypeIPv4     AssetType = "IPV4"
	AssetTypeIPv6     AssetType = "IPV6"
	AssetTypeURL      AssetType = "URL"
)

type AssetCriticality string

const (
	CriticalityLow      AssetCriticality = "LOW"
	CriticalityMedium   AssetCriticality = "MEDIUM"
	CriticalityHigh     AssetCriticality = "HIGH"
	CriticalityCritical AssetCriticality = "CRITICAL"
)

type AssetStatus string

const (
	AssetStatusActive   AssetStatus = "ACTIVE"
	AssetStatusArchived AssetStatus = "ARCHIVED"
)

// CanonicalAssetValue is an immutable, strictly validated representation of a target asset.
type CanonicalAssetValue struct {
	value string
}

func NewCanonicalAssetValue(assetType AssetType, raw string) (CanonicalAssetValue, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return CanonicalAssetValue{}, ErrEmptyCanonicalValue
	}

	switch assetType {
	case AssetTypeDomain, AssetTypeHostname:
		trimmed = strings.TrimSuffix(trimmed, ".")
		lower := strings.ToLower(trimmed)
		if strings.Contains(lower, "://") || strings.Contains(lower, "/") {
			return CanonicalAssetValue{}, fmt.Errorf("%w: domain must not contain scheme or path", ErrInvalidAssetFormat)
		}
		if !domainRegex.MatchString(lower) {
			return CanonicalAssetValue{}, fmt.Errorf("%w: invalid domain syntax", ErrInvalidAssetFormat)
		}
		return CanonicalAssetValue{value: lower}, nil

	case AssetTypeIPv4:
		ip := net.ParseIP(trimmed)
		if ip == nil || ip.To4() == nil {
			return CanonicalAssetValue{}, fmt.Errorf("%w: invalid IPv4 address", ErrInvalidAssetFormat)
		}
		return CanonicalAssetValue{value: ip.To4().String()}, nil

	case AssetTypeIPv6:
		ip := net.ParseIP(trimmed)
		if ip == nil || ip.To4() != nil {
			return CanonicalAssetValue{}, fmt.Errorf("%w: invalid IPv6 address", ErrInvalidAssetFormat)
		}
		return CanonicalAssetValue{value: ip.String()}, nil

	case AssetTypeURL:
		parsed, err := url.Parse(trimmed)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			return CanonicalAssetValue{}, fmt.Errorf("%w: URL must have valid http/https scheme and host", ErrInvalidAssetFormat)
		}
		// Canonical URL: lowercase host, strip query and fragment for root asset identity
		canonicalHost := strings.ToLower(parsed.Host)
		canonicalPath := parsed.Path
		if canonicalPath == "/" {
			canonicalPath = ""
		}
		canonicalURL := fmt.Sprintf("%s://%s%s", parsed.Scheme, canonicalHost, canonicalPath)
		return CanonicalAssetValue{value: canonicalURL}, nil

	default:
		return CanonicalAssetValue{}, ErrInvalidAssetType
	}
}

func (c CanonicalAssetValue) String() string {
	return c.value
}

func ReconstituteCanonicalAssetValue(val string) CanonicalAssetValue {
	return CanonicalAssetValue{value: val}
}

// Asset is the Aggregate Root representing an assessment target.
type Asset struct {
	id             AssetID
	projectID      ProjectID
	assetType      AssetType
	canonicalValue CanonicalAssetValue
	displayValue   string
	criticality    AssetCriticality
	status         AssetStatus
	createdAt      time.Time
	updatedAt      time.Time
}

func NewAsset(projectID ProjectID, assetType AssetType, canonicalVal CanonicalAssetValue, displayValue string, crit AssetCriticality) (*Asset, error) {
	id, err := GenerateAssetID()
	if err != nil {
		return nil, err
	}

	display := strings.TrimSpace(displayValue)
	if display == "" {
		display = canonicalVal.String()
	}

	now := time.Now().UTC()
	return &Asset{
		id:             id,
		projectID:      projectID,
		assetType:      assetType,
		canonicalValue: canonicalVal,
		displayValue:   display,
		criticality:    crit,
		status:         AssetStatusActive,
		createdAt:      now,
		updatedAt:      now,
	}, nil
}

func (a *Asset) ID() AssetID {
	return a.id
}

func (a *Asset) CreatedAt() time.Time {
	return a.createdAt
}

func (a *Asset) UpdatedAt() time.Time {
	return a.updatedAt
}

func ReconstituteAsset(
	id AssetID,
	projectID ProjectID,
	assetType AssetType,
	canonicalVal CanonicalAssetValue,
	displayValue string,
	crit AssetCriticality,
	status AssetStatus,
	createdAt time.Time,
	updatedAt time.Time,
) *Asset {
	return &Asset{
		id:             id,
		projectID:      projectID,
		assetType:      assetType,
		canonicalValue: canonicalVal,
		displayValue:   displayValue,
		criticality:    crit,
		status:         status,
		createdAt:      createdAt,
		updatedAt:      updatedAt,
	}
}

func (a *Asset) ProjectID() ProjectID {
	return a.projectID
}

func (a *Asset) AssetType() AssetType {
	return a.assetType
}

func (a *Asset) CanonicalValue() CanonicalAssetValue {
	return a.canonicalValue
}

func (a *Asset) DisplayValue() string {
	return a.displayValue
}

func (a *Asset) Criticality() AssetCriticality {
	return a.criticality
}

func (a *Asset) Status() AssetStatus {
	return a.status
}

func (a *Asset) Archive() {
	a.status = AssetStatusArchived
	a.updatedAt = time.Now().UTC()
}
