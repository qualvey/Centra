package rule

import (
	"context"
	"testing"

	"eventguard/internal/core"
	"eventguard/internal/storage"
)

func TestIPThresholdRuleTriggersAtThreshold(t *testing.T) {
	rule := NewIPThresholdRule(IPThresholdConfig{
		EventType: "singbox.reality_invalid_handshake",
		Threshold: 3,
	})
	store := storage.NewMemoryStore()
	event := core.Event{
		EventType: "singbox.reality_invalid_handshake",
		IP:        "45.227.254.152",
	}

	for i := 1; i <= 2; i++ {
		triggers, err := rule.Evaluate(context.Background(), event, store)
		if err != nil {
			t.Fatal(err)
		}
		if len(triggers) != 0 {
			t.Fatalf("iteration %d triggered early", i)
		}
	}

	triggers, err := rule.Evaluate(context.Background(), event, store)
	if err != nil {
		t.Fatal(err)
	}
	if len(triggers) != 1 {
		t.Fatalf("triggers = %d, want 1", len(triggers))
	}
	if triggers[0].Count != 3 {
		t.Fatalf("count = %d, want 3", triggers[0].Count)
	}
}
