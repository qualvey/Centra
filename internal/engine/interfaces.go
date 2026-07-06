package engine

import (
	"context"

	"eventguard/internal/core"
)

type Reader interface {
	ReadLine(ctx context.Context) (string, error)
}

type Parser interface {
	Parse(line string) (core.Event, bool, error)
}

type Rule interface {
	Evaluate(ctx context.Context, event core.Event, store Storage) ([]core.Trigger, error)
}

type Action interface {
	Execute(ctx context.Context, trigger core.Trigger) error
}

type Storage interface {
	Increment(ctx context.Context, key string) (int, error)
}
