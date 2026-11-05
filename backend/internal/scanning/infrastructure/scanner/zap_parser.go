package scanner

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

type zapReportJSON struct {
	Site []zapSiteJSON `json:"site"`
}

type zapSiteJSON struct {
	Name   string         `json:"@name"`
	Host   string         `json:"@host"`
	Port   string         `json:"@port"`
	Alerts []zapAlertJSON `json:"alerts"`
}

type zapAlertJSON struct {
	PluginID   string             `json:"pluginid"`
	AlertRef   string             `json:"alertRef"`
	Alert      string             `json:"alert"`
	Name       string             `json:"name"`
	RiskCode   string             `json:"riskcode"`
	Confidence string             `json:"confidence"`
	RiskDesc   string             `json:"riskdesc"`
	Desc       string             `json:"desc"`
	Solution   string             `json:"solution"`
	Reference  string             `json:"reference"`
	CWEID      string             `json:"cweid"`
	WASCID     string             `json:"wascid"`
	Instances  []zapAlertInstance `json:"instances"`
}

type zapAlertInstance struct {
	URI    string `json:"uri"`
	Method string `json:"method"`
	Param  string `json:"param"`
}

// ZapParser parses bounded OWASP ZAP JSON reports into normalized domain observations.
type ZapParser struct{}

func NewZapParser() *ZapParser {
	return &ZapParser{}
}

func (p *ZapParser) Parse(reader io.Reader, maxBytes int64) ([]Observation, error) {
	if reader == nil {
		return nil, errors.New("reader cannot be nil")
	}

	limitReader := io.LimitReader(reader, maxBytes)

	var report zapReportJSON
	decoder := json.NewDecoder(limitReader)
	if err := decoder.Decode(&report); err != nil {
		return nil, fmt.Errorf("failed to decode zap JSON: %w", err)
	}

	var observations []Observation

	for _, site := range report.Site {
		for _, alert := range site.Alerts {
			title := alert.Alert
			if title == "" {
				title = alert.Name
			}

			severity := normalizeZapSeverity(alert.RiskCode)
			confidence := normalizeZapConfidence(alert.Confidence)

			location := site.Name
			if len(alert.Instances) > 0 && alert.Instances[0].URI != "" {
				location = alert.Instances[0].URI
			}

			vulnKey := fmt.Sprintf("ZAP-%s", alert.PluginID)
			if alert.PluginID == "" {
				vulnKey = fmt.Sprintf("ZAP-%s", alert.AlertRef)
			}

			var descParts []string
			if alert.Desc != "" {
				descParts = append(descParts, alert.Desc)
			}
			if alert.Solution != "" {
				descParts = append(descParts, fmt.Sprintf("Solution: %s", alert.Solution))
			}
			if alert.CWEID != "" {
				descParts = append(descParts, fmt.Sprintf("CWE-%s", alert.CWEID))
			}

			observations = append(observations, Observation{
				Title:            title,
				Description:      strings.Join(descParts, "\n\n"),
				Severity:         severity,
				Confidence:       confidence,
				SourceScanner:    "zap",
				VulnerabilityKey: vulnKey,
				LocationKey:      location,
			})
		}
	}

	return observations, nil
}

func normalizeZapSeverity(riskCode string) string {
	switch strings.TrimSpace(riskCode) {
	case "0":
		return "INFORMATIONAL"
	case "1":
		return "LOW"
	case "2":
		return "MEDIUM"
	case "3":
		return "HIGH"
	case "4":
		return "CRITICAL"
	default:
		return "INFORMATIONAL"
	}
}

func normalizeZapConfidence(conf string) string {
	switch strings.TrimSpace(conf) {
	case "1":
		return "LOW"
	case "2":
		return "MEDIUM"
	case "3":
		return "HIGH"
	case "4":
		return "CERTAIN"
	default:
		return "LOW"
	}
}
