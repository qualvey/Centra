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
