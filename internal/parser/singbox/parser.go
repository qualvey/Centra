package singbox

import (
	"net/netip"
	"regexp"
	"strings"
	"time"

	"eventguard/internal/core"
)

const RealityInvalidHandshake = "singbox.reality_invalid_handshake"

var ipPattern = regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(line string) (core.Event, bool, error) {
	lower := strings.ToLower(line)
	if !strings.Contains(lower, "sing-box") && !strings.Contains(lower, "singbox") {
		return core.Event{}, false, nil
	}

	ip := extractIP(line)
	eventType := classify(lower)
	if eventType == "" || ip == "" {
		return core.Event{}, false, nil
	}

	return core.Event{
		Timestamp: time.Now().UTC(),
		Source:    "journalctl",
		Service:   "sing-box",
		EventType: eventType,
		Level:     extractLevel(lower),
		IP:        ip,
		Message:   line,
		Metadata: map[string]string{
			"parser": "singbox",
		},
	}, true, nil
}

func classify(lower string) string {
	if strings.Contains(lower, "reality") && strings.Contains(lower, "invalid") && strings.Contains(lower, "handshake") {
		return RealityInvalidHandshake
	}
	if strings.Contains(lower, "tls handshake") && strings.Contains(lower, "reality") && strings.Contains(lower, "processed invalid connection") {
		return RealityInvalidHandshake
	}
	return ""
}

func extractIP(line string) string {
	candidates := ipPattern.FindAllString(line, -1)
	for _, candidate := range candidates {
		addr, err := netip.ParseAddr(candidate)
		if err == nil && addr.Is4() {
			return candidate
		}
	}
	return ""
}

func extractLevel(lower string) string {
	for _, level := range []string{"debug", "info", "warn", "error", "fatal"} {
		if strings.Contains(lower, level) {
			return level
		}
	}
	return ""
}
