package singbox

import "testing"

func TestParseRealityInvalidHandshake(t *testing.T) {
	parser := NewParser()
	line := "2026-07-07T00:00:00+08:00 host sing-box[123]: WARN reality: invalid handshake from 45.227.254.152:443"

	event, ok, err := parser.Parse(line)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected parser to emit event")
	}
	if event.EventType != RealityInvalidHandshake {
		t.Fatalf("event type = %q, want %q", event.EventType, RealityInvalidHandshake)
	}
	if event.IP != "45.227.254.152" {
		t.Fatalf("ip = %q", event.IP)
	}
	if event.Service != "sing-box" {
		t.Fatalf("service = %q", event.Service)
	}
}

func TestParseRealityProcessedInvalidConnection(t *testing.T) {
	parser := NewParser()
	line := "2026-07-06T21:47:27+08:00 tw sing-box[2631710]: ERROR inbound/vless[vless-in]: process connection from 45.194.67.28:51078: TLS handshake: REALITY: processed invalid connection"

	event, ok, err := parser.Parse(line)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected parser to emit event")
	}
	if event.EventType != RealityInvalidHandshake {
		t.Fatalf("event type = %q, want %q", event.EventType, RealityInvalidHandshake)
	}
	if event.IP != "45.194.67.28" {
		t.Fatalf("ip = %q", event.IP)
	}
}

func TestParseIgnoresUnrelatedLine(t *testing.T) {
	parser := NewParser()

	_, ok, err := parser.Parse("sshd[123]: failed password for root from 203.0.113.10")
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("expected unrelated line to be ignored")
	}
}
