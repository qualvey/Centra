package rule

import (
	"context"
	"fmt"

	"eventguard/internal/core"
	"eventguard/internal/engine"
)

type IPThresholdConfig struct {
	EventType string
	Threshold int
}

type IPThresholdRule struct {
	config IPThresholdConfig
}

func NewIPThresholdRule(config IPThresholdConfig) *IPThresholdRule {
	if config.Threshold <= 0 {
		config.Threshold = 5
	}
	return &IPThresholdRule{config: config}
}

func (r *IPThresholdRule) Evaluate(ctx context.Context, event core.Event, store engine.Storage) ([]core.Trigger, error) {
	if event.IP == "" || event.EventType != r.config.EventType {
		return nil, nil
	}

	key := fmt.Sprintf("event_count:%s:%s", event.EventType, event.IP)
	count, err := store.Increment(ctx, key)
	if err != nil {
		return nil, err
	}
	if count < r.config.Threshold {
		return nil, nil
	}

	markKey := fmt.Sprintf("triggered:%s:%s", event.EventType, event.IP)
	firstTrigger, err := store.MarkOnce(ctx, markKey)
	if err != nil {
		return nil, err
	}
	if !firstTrigger {
		return nil, nil
	}

	return []core.Trigger{{
		Event:  event,
		Reason: "REALITY Invalid Handshake",
		Count:  count,
	}}, nil
}
