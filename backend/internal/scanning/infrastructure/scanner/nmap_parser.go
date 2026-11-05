package scanner

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"
)

type Observation struct {
	Title            string
	Description      string
	Severity         string
	Confidence       string
	SourceScanner    string
	VulnerabilityKey string
	LocationKey      string
}

// NmapXML structures for parsing <nmaprun>
type nmapRunXML struct {
	XMLName xml.Name      `xml:"nmaprun"`
	Hosts   []nmapHostXML `xml:"host"`
}

type nmapHostXML struct {
	Status    nmapStatusXML   `xml:"status"`
	Addresses []nmapAddrXML   `xml:"address"`
	Ports     nmapPortsXML    `xml:"ports"`
}

type nmapStatusXML struct {
	State  string `xml:"state,attr"`
	Reason string `xml:"reason,attr"`
}

type nmapAddrXML struct {
	Addr     string `xml:"addr,attr"`
	AddrType string `xml:"addrtype,attr"`
}

type nmapPortsXML struct {
	Ports []nmapPortXML `xml:"port"`
}

type nmapPortXML struct {
	Protocol string         `xml:"protocol,attr"`
	PortID   int            `xml:"portid,attr"`
	State    nmapStateXML   `xml:"state"`
	Service  nmapServiceXML `xml:"service"`
}

type nmapStateXML struct {
	State string `xml:"state,attr"`
}

type nmapServiceXML struct {
	Name    string `xml:"name,attr"`
	Product string `xml:"product,attr"`
	Version string `xml:"version,attr"`
}

// NmapParser parses bounded XML streams into normalized domain observations.
type NmapParser struct{}

func NewNmapParser() *NmapParser {
	return &NmapParser{}
}

func (p *NmapParser) Parse(reader io.Reader, maxBytes int64) ([]Observation, error) {
	if reader == nil {
		return nil, errors.New("reader cannot be nil")
	}

	// Read with explicit size limit to prevent memory exhaustion
	limitReader := io.LimitReader(reader, maxBytes)

	var run nmapRunXML
	decoder := xml.NewDecoder(limitReader)
	if err := decoder.Decode(&run); err != nil {
		return nil, fmt.Errorf("failed to decode nmap XML: %w", err)
	}

	var observations []Observation

	for _, host := range run.Hosts {
		if host.Status.State != "up" {
			continue
		}

		for _, port := range host.Ports.Ports {
			if port.State.State != "open" {
				continue
			}

			loc := fmt.Sprintf("%s/%d", port.Protocol, port.PortID)
			title := fmt.Sprintf("Open Port %d %s", port.PortID, strings.ToUpper(port.Protocol))

			var descParts []string
			if port.Service.Name != "" {
				descParts = append(descParts, fmt.Sprintf("Service: %s", port.Service.Name))
			}
			if port.Service.Product != "" {
				descParts = append(descParts, fmt.Sprintf("Product: %s", port.Service.Product))
			}
			if port.Service.Version != "" {
				descParts = append(descParts, fmt.Sprintf("Version: %s", port.Service.Version))
			}

			desc := strings.Join(descParts, ", ")
			if desc == "" {
				desc = fmt.Sprintf("Port %d is open and listening", port.PortID)
			}

			observations = append(observations, Observation{
				Title:            title,
				Description:      desc,
				Severity:         "INFO",
				Confidence:       "CONFIRMED",
				SourceScanner:    "nmap",
				VulnerabilityKey: fmt.Sprintf("open-port-%s-%d", port.Protocol, port.PortID),
				LocationKey:      loc,
			})
		}
	}

	return observations, nil
}
