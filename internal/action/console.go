package action

import (
	"context"
	"fmt"
	"io"

	"eventguard/internal/core"
)

type ConsoleSuggestion struct {
	writer io.Writer
}

func NewConsoleSuggestion(writer io.Writer) *ConsoleSuggestion {
	return &ConsoleSuggestion{writer: writer}
}

func (a *ConsoleSuggestion) Execute(ctx context.Context, trigger core.Trigger) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	_, err := fmt.Fprintf(a.writer, "Need Block:\n\n%s\n\nReason:\n\n%s\n\nCount:\n\n%d\n\n", trigger.Event.IP, trigger.Reason, trigger.Count)
	return err
}
